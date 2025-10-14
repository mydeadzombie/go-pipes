package nodes

import (
	"context"
	"fmt"
	"sync"

	"go-pipes/pkg/pipe"
)

type Printer struct {
	pipe.BaseNode
	Quiet   bool
	Workers int
}

func NewPrinter(id string, quiet bool) *Printer {
	return &Printer{BaseNode: pipe.BaseNode{IDValue: id}, Quiet: quiet, Workers: 1}
}

func (n *Printer) Start(ctx context.Context) error {
	defer n.CloseOutputs()
	in, _ := n.GetInput("in")
	if in == nil {
		return nil
	}
	workers := n.Workers
	if workers <= 0 {
		workers = 1
	}
	var wg sync.WaitGroup
	wg.Add(workers)
	for wid := 1; wid <= workers; wid++ {
		go func(wid int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case v, ok := <-in:
					if !ok {
						return
					}
					if n.Quiet {
						continue
					}
					switch r := v.(type) {
					case MD5Result:
						if r.Err != nil {
							fmt.Printf("worker=%d ERROR: %s: %v\n", wid, r.Path, r.Err)
						} else {
							fmt.Printf("worker=%d %s\n", wid, r.String())
						}
					default:
						fmt.Printf("worker=%d %v\n", wid, r)
					}
				}
			}
		}(wid)
	}
	wg.Wait()
	return nil
}
