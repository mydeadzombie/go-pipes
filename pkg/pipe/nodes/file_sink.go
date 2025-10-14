package nodes

import (
    "bufio"
    "context"
    "fmt"
    "os"

    "go-pipes/pkg/pipe"
)

type FileSink struct {
    pipe.BaseNode
    Path   string
    Append bool
    Workers int
}

func NewFileSink(id, path string, append bool) *FileSink {
    return &FileSink{BaseNode: pipe.BaseNode{IDValue: id}, Path: path, Append: append, Workers: 1}
}

func (n *FileSink) Start(ctx context.Context) error {
    defer n.CloseOutputs()
    in, _ := n.GetInput("in")
    if in == nil {
        return nil
    }
    if n.Path == "" {
        n.Path = "md5-output.txt"
    }
    flag := os.O_CREATE | os.O_WRONLY
    if n.Append {
        flag |= os.O_APPEND
    } else {
        flag |= os.O_TRUNC
    }
    f, err := os.OpenFile(n.Path, flag, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    w := bufio.NewWriter(f)
    defer w.Flush()

    workers := n.Workers
    if workers <= 1 {
        // single-threaded streaming write
        for {
            select {
            case <-ctx.Done():
                return ctx.Err()
            case v, ok := <-in:
                if !ok {
                    return nil
                }
                switch r := v.(type) {
                case MD5Result:
                    if r.Err != nil {
                        fmt.Fprintf(w, "ERROR: %s: %v\n", r.Path, r.Err)
                    } else {
                        fmt.Fprintf(w, "%s\n", r.String())
                    }
                default:
                    fmt.Fprintf(w, "%v\n", r)
                }
            }
        }
    }

    // Multi-worker: group output by sink worker id
    type buf = []string
    bufs := make([]buf, workers)

    done := make(chan struct{}, workers)
    for wid := 1; wid <= workers; wid++ {
        idx := wid - 1
        go func(wid int, idx int) {
            for {
                select {
                case <-ctx.Done():
                    done <- struct{}{}
                    return
                case v, ok := <-in:
                    if !ok {
                        done <- struct{}{}
                        return
                    }
                    switch r := v.(type) {
                    case MD5Result:
                        if r.Err != nil {
                            bufs[idx] = append(bufs[idx], fmt.Sprintf("worker=%d ERROR: %s: %v", wid, r.Path, r.Err))
                        } else {
                            bufs[idx] = append(bufs[idx], fmt.Sprintf("worker=%d %s", wid, r.String()))
                        }
                    default:
                        bufs[idx] = append(bufs[idx], fmt.Sprintf("worker=%d %v", wid, r))
                    }
                }
            }
        }(wid, idx)
    }

    // wait workers to finish (on input close or ctx cancel)
    for i := 0; i < workers; i++ {
        <-done
    }

    // write grouped sections in order
    for wid := 1; wid <= workers; wid++ {
        idx := wid - 1
        if len(bufs[idx]) == 0 { continue }
        fmt.Fprintf(w, "== worker %d ==\n", wid)
        for _, line := range bufs[idx] {
            fmt.Fprintln(w, line)
        }
    }
    return nil
}


