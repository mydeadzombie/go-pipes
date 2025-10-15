package nodes

import (
    "bufio"
    "context"
    "fmt"
    "os"
    "path/filepath"
    "strings"

    "go-pipes/pkg/pipe"
)

type StdinSource struct {
    pipe.BaseNode
    Prompt     string
    AllowEmpty bool
}

func NewStdinSource(id string, prompt string, allowEmpty bool) *StdinSource {
    return &StdinSource{BaseNode: pipe.BaseNode{IDValue: id}, Prompt: prompt, AllowEmpty: allowEmpty}
}

func (n *StdinSource) Start(ctx context.Context) error {
    defer n.CloseOutputs()
    out, _ := n.GetOutput("paths")
    if out == nil {
        return nil
    }
    if n.Prompt == "" {
        n.Prompt = "Enter file path: "
    }
    reader := bufio.NewReader(os.Stdin)
    fmt.Fprint(os.Stdout, n.Prompt)
    line, err := reader.ReadString('\n')
    if err != nil {
        return err
    }
    path := strings.TrimSpace(line)
    if path == "" && !n.AllowEmpty {
        return fmt.Errorf("empty input")
    }
    if path != "" {
        if !filepath.IsAbs(path) {
            if abs, err := filepath.Abs(path); err == nil { path = abs }
        }
        if fi, err := os.Stat(path); err != nil {
            return err
        } else if fi.IsDir() {
            return fmt.Errorf("path is a directory: %s", path)
        }
        select {
        case <-ctx.Done():
            return ctx.Err()
        case out <- path:
        }
    }
    return nil
}


