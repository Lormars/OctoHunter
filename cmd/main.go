package main

import (
	"github.com/lormars/octohunter/internal/parser"
	"github.com/lormars/octohunter/pkg/modules"
)

func main() {
	options := parser.Parse_Options()

	if options.Hopper {
		modules.CheckHop(options)
	}
}
