package parser

import (
	"flag"

	"github.com/lormars/octohunter/common"
)

var modules common.ModuleList

func Parse_Options() *common.Opts {

	flag.Var(&modules, "modules", "The module to run")
	var (
		target       = flag.String("target", "none", "The target to scan")
		cnameFile    = flag.String("cnamefile", "none", "The file to scan for subdomain takeover")
		dorkFile     = flag.String("dorkfile", "none", "The file to scan for Google dork")
		methodFile   = flag.String("methodfile", "none", "The file to scan for HTTP method checker")
		redirectFile = flag.String("redirectfile", "none", "The file to scan for redirect checker")
		hopFile      = flag.String("hopfile", "none", "The file to scan for hopper")
	)
	flag.Parse()
	return &common.Opts{
		Module:       modules,
		Target:       *target,
		CnameFile:    *cnameFile,
		DorkFile:     *dorkFile,
		MethodFile:   *methodFile,
		RedirectFile: *redirectFile,
		HopFile:      *hopFile,
	}
}
