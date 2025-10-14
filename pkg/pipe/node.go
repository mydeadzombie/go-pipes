package pipe

import (
	"context"
	"sync"
)

// Node represents a processing unit in the pipeline graph.
// Nodes have named input and output ports. Implementations should:
// - read from all configured inputs until they are closed or ctx is cancelled
// - write results to outputs
// - close their outputs before returning from Start
type Node interface {
	ID() string

	// Port metadata
	InPorts() []string
	OutPorts() []string

	// Connections
	SetInput(port string, ch <-chan any)
	SetOutput(port string, ch chan any)
	GetInput(port string) (<-chan any, bool)
	GetOutput(port string) (chan any, bool)

	// Lifecycle
	Start(ctx context.Context) error
	CloseOutputs()
}

// BaseNode provides common storage for ports and a helper to close all outputs.
type BaseNode struct {
	IDValue string
	In      map[string]<-chan any
	Out     map[string]chan any
	once    sync.Once
}

func (b *BaseNode) ID() string { return b.IDValue }

func (b *BaseNode) InPorts() []string {
	ports := make([]string, 0, len(b.In))
	for p := range b.In {
		ports = append(ports, p)
	}
	return ports
}

func (b *BaseNode) OutPorts() []string {
	ports := make([]string, 0, len(b.Out))
	for p := range b.Out {
		ports = append(ports, p)
	}
	return ports
}

func (b *BaseNode) SetInput(port string, ch <-chan any) {
	if b.In == nil {
		b.In = make(map[string]<-chan any)
	}
	b.In[port] = ch
}

func (b *BaseNode) GetInput(port string) (<-chan any, bool) {
	if b.In == nil {
		return nil, false
	}
	ch, ok := b.In[port]
	return ch, ok
}

func (b *BaseNode) GetOutput(port string) (chan any, bool) {
	if b.Out == nil {
		return nil, false
	}
	ch, ok := b.Out[port]
	return ch, ok
}

func (b *BaseNode) SetOutput(port string, ch chan any) {
	if b.Out == nil {
		b.Out = make(map[string]chan any)
	}
	b.Out[port] = ch
}

func (b *BaseNode) CloseOutputs() {
	b.once.Do(func() {
		for _, ch := range b.Out {
			close(ch)
		}
	})
}

// (Typed helper adapters were removed in rollback)
