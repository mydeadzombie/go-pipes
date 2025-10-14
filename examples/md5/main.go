package main

import (
	"context"
	"flag"
	"log"
	"os"

	"go-pipes/pkg/pipe"
	"go-pipes/pkg/pipe/loader"
)

func main() {
	var (
		yamlPath string
		dir      string
		workers  int
		quiet    bool
	)
	flag.StringVar(&yamlPath, "pipeline", "examples/md5/pipeline.yml", "Path to pipeline YAML")
	flag.StringVar(&dir, "dir", ".", "Directory to walk as default")
	flag.IntVar(&workers, "parallelism", 10, "MD5 hashing parallelism")
	flag.BoolVar(&quiet, "quiet", false, "Suppress output")
	flag.Parse()

	// Build registry with builtins and CLI overrides as defaults
	reg := loader.Builtins(dir, workers, quiet)

	g, err := loader.LoadFromFile(yamlPath, reg)
	if err != nil {
		log.Println("failed loading pipeline:", err)
		os.Exit(1)
	}
	if err := pipe.NewRunner(g).Run(context.Background()); err != nil {
		log.Println("pipeline error:", err)
		os.Exit(1)
	}
}
