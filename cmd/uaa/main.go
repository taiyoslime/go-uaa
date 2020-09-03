package main

import (
	"golang.org/x/tools/go/analysis/unitchecker"
	"uaa"
)

func main() { unitchecker.Main(uaa.Analyzer) }
