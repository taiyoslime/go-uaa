package main

import (
	"uaa"
	"golang.org/x/tools/go/analysis/unitchecker"
)

func main() { unitchecker.Main(uaa.Analyzer) }

