package pipe

import (
    "context"
    "fmt"
    "sync"
)

// Runner starts nodes, handles cancellation and waits for completion.
type Runner struct {
    g *Graph
}

func NewRunner(g *Graph) *Runner { return &Runner{g: g} }

func (r *Runner) Run(ctx context.Context) error {
    if r.g == nil {
        return fmt.Errorf("nil graph")
    }
    if err := r.g.materialize(); err != nil {
        return err
    }

    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    var wg sync.WaitGroup
    errs := make(chan error, len(r.g.nodes))

    for _, n := range r.g.nodes {
        wg.Add(1)
        go func(n Node) {
            defer wg.Done()
            if err := n.Start(ctx); err != nil {
                select {
                case errs <- err:
                default:
                }
                cancel()
            }
        }(n)
    }

    // Wait and ensure outputs are closed to unblock downstreams.
    done := make(chan struct{})
    go func() {
        wg.Wait()
        close(done)
    }()

    select {
    case <-done:
        // all good
    case err := <-errs:
        if err != nil {
            <-done
            return err
        }
    case <-ctx.Done():
        <-done
        return ctx.Err()
    }

    return nil
}


