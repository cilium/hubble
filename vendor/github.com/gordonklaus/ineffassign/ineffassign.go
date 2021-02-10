package main

import (
	"flag"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(ineffassign.Analyzer)
}

func init() {
	flag.Bool("n", false, "don't recursively check paths (deprecated)")
}
