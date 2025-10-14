package nodes

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"

	"go-pipes/pkg/pipe"
)

type FileWalker struct {
	pipe.BaseNode
	Dirs    []string
	Workers int
}

func NewFileWalker(id string, dirOrDirs ...string) *FileWalker {
	dirs := dirOrDirs
	if len(dirs) == 0 {
		dirs = []string{"."}
	}
	return &FileWalker{BaseNode: pipe.BaseNode{IDValue: id}, Dirs: dirs, Workers: 1}
}

func (n *FileWalker) Start(ctx context.Context) error {
	defer n.CloseOutputs()
	out, _ := n.GetOutput("files")
	if out == nil {
		return nil
	}

	// Worker routine that walks one root directory with symlink safety
	walkOne := func(ctx context.Context, root string) error {
		if root == "" {
			root = "."
		}
		if abs, err := filepath.Abs(root); err == nil {
			root = abs
		}
		if rp, err := filepath.EvalSymlinks(root); err == nil {
			root = rp
		}

		visited := make(map[string]struct{}) // resolved dir paths
		stack := []string{root}

		for len(stack) > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			dir := stack[len(stack)-1]
			stack = stack[:len(stack)-1]

			resolvedDir := dir
			if rp, err := filepath.EvalSymlinks(dir); err == nil {
				resolvedDir = rp
			}
			if _, seen := visited[resolvedDir]; seen {
				continue
			}
			visited[resolvedDir] = struct{}{}

			entries, err := os.ReadDir(dir)
			if err != nil {
				return err
			}
			for _, e := range entries {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				full := filepath.Join(dir, e.Name())
				mode := e.Type()

				if mode&fs.ModeSymlink != 0 {
					fi, err := os.Stat(full)
					if err != nil {
						continue
					}
					if fi.IsDir() {
						if rp, err := filepath.EvalSymlinks(full); err == nil {
							stack = append(stack, rp)
						}
						continue
					}
					if fi.Mode().IsRegular() {
						select {
						case <-ctx.Done():
							return ctx.Err()
						case out <- full:
						}
					}
					continue
				}

				if e.IsDir() {
					next := full
					if rp, err := filepath.EvalSymlinks(full); err == nil {
						next = rp
					}
					stack = append(stack, next)
					continue
				}

				if mode.IsRegular() {
					select {
					case <-ctx.Done():
						return ctx.Err()
					case out <- full:
					}
				}
			}
		}
		return nil
	}

	workers := n.Workers
	if workers <= 0 {
		workers = 1
	}
	// channel of roots to distribute among workers
	roots := make(chan string)
	// start workers
	errCh := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			for r := range roots {
				if err := walkOne(ctx, r); err != nil {
					errCh <- err
					return
				}
			}
			errCh <- nil
		}()
	}
	// feed roots
	go func() {
		for _, d := range n.Dirs {
			select {
			case <-ctx.Done():
				close(roots)
				return
			case roots <- d:
			}
		}
		close(roots)
	}()
	// wait workers
	for i := 0; i < workers; i++ {
		if err := <-errCh; err != nil {
			return err
		}
	}
	return nil
}
