package pipe

import (
	"fmt"
)

type edge struct {
	from   Node
	out    string
	to     Node
	in     string
	buffer int
}

// Graph holds nodes and their wiring.
type Graph struct {
	nodes []Node
	edges []edge
}

func NewGraph() *Graph { return &Graph{} }

func (g *Graph) Add(nodes ...Node) {
	g.nodes = append(g.nodes, nodes...)
}

// Connect wires from.out -> to.in with a buffered channel.
func (g *Graph) Connect(from Node, outPort string, to Node, inPort string, buffer int) error {
	if from == nil || to == nil {
		return fmt.Errorf("nil node in Connect")
	}
	if buffer < 0 {
		buffer = 0
	}
	g.edges = append(g.edges, edge{from: from, out: outPort, to: to, in: inPort, buffer: buffer})
	return nil
}

// materialize creates channels and assigns them to node ports.
func (g *Graph) materialize() error {
	// Map of fromNodeID:outPort to channel for potential multiple downstreams
	type key struct{ id, port string }
	outChans := make(map[key]chan any)

	for _, e := range g.edges {
		k := key{id: e.from.ID(), port: e.out}
		ch, ok := outChans[k]
		if !ok {
			// create new channel for this output
			ch = make(chan any, e.buffer)
			e.from.SetOutput(e.out, ch)
			outChans[k] = ch
		}
		// connect input
		e.to.SetInput(e.in, ch)
	}
	return nil
}
