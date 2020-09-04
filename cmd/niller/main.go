package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"
	"niller"
)

func main() { unitchecker.Main(niller.Analyzer) }
