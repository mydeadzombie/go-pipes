package nodes

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"sync"

	"go-pipes/pkg/pipe"
)

type MD5Result struct {
	Path string
	Sum  [16]byte
	Err  error
}

type MD5Hasher struct {
	pipe.BaseNode
	Workers int
}

func NewMD5Hasher(id string, workers int) *MD5Hasher {
	if workers <= 0 {
		workers = 10
	}
	return &MD5Hasher{BaseNode: pipe.BaseNode{IDValue: id}, Workers: workers}
}

func (n *MD5Hasher) Start(ctx context.Context) error {
	defer n.CloseOutputs()
	in, _ := n.GetInput("paths")
	out, _ := n.GetOutput("results")
	if in == nil || out == nil {
		return nil
	}

	var wg sync.WaitGroup
	wg.Add(n.Workers)
	for i := 0; i < n.Workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case p, ok := <-in:
					if !ok {
						return
					}
					path, _ := p.(string)
					res := MD5Result{Path: path}
					f, err := os.Open(path)
					if err != nil {
						res.Err = err
					} else {
						h := md5.New()
						if _, err := io.Copy(h, f); err != nil {
							res.Err = err
						} else {
							copy(res.Sum[:], h.Sum(nil))
						}
						_ = f.Close()
					}
					select {
					case <-ctx.Done():
						return
					case out <- res:
					}
				}
			}
		}()
	}
	wg.Wait()
	return nil
}

func (r MD5Result) String() string {
	return fmt.Sprintf("%x  %s", r.Sum, r.Path)
}
