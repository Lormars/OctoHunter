package parser

import (
	"flag"

	"github.com/lormars/octohunter/common"
)

func Parse_Options() *common.Opts {
	var (
		hopper = flag.Bool("hopper", false, "Enable the hopper")
		target = flag.String("target", "none", "The target to scan")
		file   = flag.String("file", "none", "The file to scan")
	)
	return &common.Opts{
		Hopper: *hopper,
		Target: *target,
		File:   *file,
	}
}
