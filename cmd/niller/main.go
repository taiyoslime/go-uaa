package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"
	"github.com/taiyoslime/niller"
)

func main() { unitchecker.Main(niller.Analyzer) }
