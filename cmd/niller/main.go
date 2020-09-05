package main

import (
	"github.com/taiyoslime/niller"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(niller.Analyzer) }
