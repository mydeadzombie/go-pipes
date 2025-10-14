package nodes

import (
	"context"

	"go-pipes/pkg/pipe"
)

// Tee duplicates items from input to two outputs: out1 and out2.
// Backpressure applies if either downstream is slow.
type Tee struct {
	pipe.BaseNode
}

func NewTee(id string) *Tee { return &Tee{BaseNode: pipe.BaseNode{IDValue: id}} }

func (n *Tee) Start(ctx context.Context) error {
	defer n.CloseOutputs()
	in, _ := n.GetInput("in")
	out1, _ := n.GetOutput("out1")
	out2, _ := n.GetOutput("out2")
	if in == nil || out1 == nil || out2 == nil {
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case v, ok := <-in:
			if !ok {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case out1 <- v:
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case out2 <- v:
			}
		}
	}
}
