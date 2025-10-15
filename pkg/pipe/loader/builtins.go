package loader

import (
	"fmt"
	"go-pipes/pkg/pipe"
	"go-pipes/pkg/pipe/nodes"
)

type Defaults struct {
	Dir     string
	Workers int
	Quiet   bool
}

type BuiltinFactory func(id string, cfg map[string]any, d Defaults) (pipe.Node, error)

func getString(cfg map[string]any, key, def string) string {
	if v, ok := cfg[key].(string); ok && v != "" {
		return v
	}
	return def
}

func getBool(cfg map[string]any, key string, def bool) bool {
	if v, ok := cfg[key].(bool); ok {
		return v
	}
	return def
}

func getInt(cfg map[string]any, key string, def int) int {
	if v, ok := cfg[key].(int); ok {
		return v
	}
	if v, ok := cfg[key].(float64); ok {
		return int(v)
	}
	return def
}

func getStringList(cfg map[string]any, key string, def []string) []string {
	if s, ok := cfg[key].(string); ok && s != "" {
		return []string{s}
	}
	if xs, ok := cfg[key].([]any); ok {
		out := make([]string, 0, len(xs))
		for _, x := range xs {
			if s, ok := x.(string); ok {
				out = append(out, s)
			}
		}
		if len(out) > 0 {
			return out
		}
	}
	return def
}

func builtinFactories() map[string]BuiltinFactory {
	return map[string]BuiltinFactory{
        "stdin_source": func(id string, cfg map[string]any, d Defaults) (pipe.Node, error) {
            if id == "" { return nil, fmt.Errorf("empty id") }
            prompt := getString(cfg, "prompt", "Enter file path: ")
            allowEmpty := getBool(cfg, "allowEmpty", false)
            return nodes.NewStdinSource(id, prompt, allowEmpty), nil
        },
		"file_walker": func(id string, cfg map[string]any, d Defaults) (pipe.Node, error) {
			if id == "" {
				return nil, fmt.Errorf("empty id")
			}
			dirs := getStringList(cfg, "dir", nil)
			if len(dirs) == 0 {
				if d.Dir != "" {
					dirs = []string{d.Dir}
				} else {
					dirs = []string{"."}
				}
			}
			workers := getInt(cfg, "workers", 1)
			n := nodes.NewFileWalker(id, dirs...)
			n.Workers = workers
			return n, nil
		},
		"md5_hasher": func(id string, cfg map[string]any, d Defaults) (pipe.Node, error) {
			if id == "" {
				return nil, fmt.Errorf("empty id")
			}
			workers := getInt(cfg, "workers", d.Workers)
			return nodes.NewMD5Hasher(id, workers), nil
		},
		"printer": func(id string, cfg map[string]any, d Defaults) (pipe.Node, error) {
			if id == "" {
				return nil, fmt.Errorf("empty id")
			}
			quiet := getBool(cfg, "quiet", d.Quiet)
			workers := getInt(cfg, "workers", 1)
			n := nodes.NewPrinter(id, quiet)
			n.Workers = workers
			return n, nil
		},
		"file_sink": func(id string, cfg map[string]any, d Defaults) (pipe.Node, error) {
			if id == "" {
				return nil, fmt.Errorf("empty id")
			}
			path := getString(cfg, "path", "output.txt")
			append := getBool(cfg, "append", false)
			workers := getInt(cfg, "workers", 1)
			n := nodes.NewFileSink(id, path, append)
			n.Workers = workers
			return n, nil
		},
		"tee": func(id string, cfg map[string]any, d Defaults) (pipe.Node, error) {
			if id == "" {
				return nil, fmt.Errorf("empty id")
			}
			return nodes.NewTee(id), nil
		},
	}
}

// Builtins returns a registry pre-populated with standard nodes.
func Builtins(dirDefault string, workersDefault int, quietDefault bool) *Registry {
	reg := NewRegistry()
	defaults := Defaults{Dir: dirDefault, Workers: workersDefault, Quiet: quietDefault}
	for typ, factory := range builtinFactories() {
		t := typ
		f := factory
		reg.Register(t, func(spec NodeSpec) (pipe.Node, error) {
			return f(spec.ID, spec.Config, defaults)
		})
	}
	return reg
}
