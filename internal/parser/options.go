package parser

import (
	"flag"

	"github.com/lormars/octohunter/common"
)

func Parse_Options() *common.Opts {
	var (
		hopper = flag.Bool("hopper", false, "Enable the hopper")
		dork   = flag.Bool("dork", false, "Enable the dorker")
		broker = flag.Bool("broker", false, "Enable the broker")
		method = flag.Bool("method", false, "Enable the HTTP method checker")
		cname  = flag.Bool("cname", false, "Enable the CNAME takeover checker")
		target = flag.String("target", "none", "The target to scan")
		file   = flag.String("file", "none", "The file to scan")
	)
	flag.Parse()
	return &common.Opts{
		Hopper: *hopper,
		Dork:   *dork,
		Broker: *broker,
		Method: *method,
		Cname:  *cname,
		Target: *target,
		File:   *file,
	}
}
