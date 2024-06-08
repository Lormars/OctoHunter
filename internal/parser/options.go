package parser

import (
	"flag"

	"github.com/lormars/octohunter/common"
)

func Parse_Options() *common.Opts {
	var (
		hopper = flag.Bool("hopper", false, "Enable the hopper")
		dork   = flag.Bool("dork", false, "Enable the dorker")
		target = flag.String("target", "none", "The target to scan")
		file   = flag.String("file", "none", "The file to scan")
	)
	flag.Parse()
	return &common.Opts{
		Hopper: *hopper,
		Dork:   *dork,
		Target: *target,
		File:   *file,
	}
}
