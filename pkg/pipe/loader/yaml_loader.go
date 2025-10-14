package loader

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"go-pipes/pkg/pipe"
)

// YAML schema structures
type PipelineSpec struct {
	Nodes []NodeSpec `yaml:"nodes"`
	Edges []EdgeSpec `yaml:"edges"`
}

type NodeSpec struct {
	ID     string         `yaml:"id"`
	Type   string         `yaml:"type"`
	Config map[string]any `yaml:"config"`
}

type EdgeSpec struct {
	From   string `yaml:"from"` // nodeID.out
	To     string `yaml:"to"`   // nodeID.in
	Buffer int    `yaml:"buffer"`
}

// NodeFactory creates a pipe.Node from a NodeSpec.
type NodeFactory func(spec NodeSpec) (pipe.Node, error)

type Registry struct {
	factories map[string]NodeFactory
}

func NewRegistry() *Registry { return &Registry{factories: map[string]NodeFactory{}} }

func (r *Registry) Register(nodeType string, f NodeFactory) { r.factories[nodeType] = f }

func (r *Registry) Build(spec NodeSpec) (pipe.Node, error) {
	f, ok := r.factories[spec.Type]
	if !ok {
		return nil, fmt.Errorf("unknown node type: %s", spec.Type)
	}
	return f(spec)
}

// LoadFromReader reads YAML, constructs a Graph using provided registry.
func LoadFromReader(r io.Reader, reg *Registry) (*pipe.Graph, error) {
	var spec PipelineSpec
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&spec); err != nil {
		return nil, err
	}
	g := pipe.NewGraph()
	idToNode := make(map[string]pipe.Node)
	for _, ns := range spec.Nodes {
		if ns.ID == "" {
			return nil, fmt.Errorf("node id is required")
		}
		if _, exists := idToNode[ns.ID]; exists {
			return nil, fmt.Errorf("duplicate node id: %s", ns.ID)
		}
		n, err := reg.Build(ns)
		if err != nil {
			return nil, err
		}
		g.Add(n)
		idToNode[ns.ID] = n
	}
	for _, es := range spec.Edges {
		var fromID, fromPort string
		var toID, toPort string
		if idx := strings.LastIndex(es.From, "."); idx <= 0 || idx >= len(es.From)-1 {
			return nil, fmt.Errorf("invalid from format %q; expected node.port", es.From)
		} else {
			fromID = es.From[:idx]
			fromPort = es.From[idx+1:]
		}
		if idx := strings.LastIndex(es.To, "."); idx <= 0 || idx >= len(es.To)-1 {
			return nil, fmt.Errorf("invalid to format %q; expected node.port", es.To)
		} else {
			toID = es.To[:idx]
			toPort = es.To[idx+1:]
		}
		from := idToNode[fromID]
		to := idToNode[toID]
		if from == nil || to == nil {
			return nil, fmt.Errorf("unknown node id in edge: %q -> %q", es.From, es.To)
		}
		if err := g.Connect(from, fromPort, to, toPort, es.Buffer); err != nil {
			return nil, err
		}
	}
	return g, nil
}

func LoadFromFile(path string, reg *Registry) (*pipe.Graph, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return LoadFromReader(f, reg)
}
