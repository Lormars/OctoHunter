package parser

import (
	"flag"

	"github.com/lormars/octohunter/common"
)

var modules common.ModuleList

func Parse_Options() (*common.Opts, string, int, bool) {

	flag.Var(&modules, "modules", "The module to run")
	var (
		target         = flag.String("target", "none", "The target to scan")
		cacheTime      = flag.Int("cache", 60, "The cache time to use")
		concurrency    = flag.Int("concurrency", 500, "The concurrency to use")
		logLevel       = flag.String("loglevel", "info", "The log level to use")
		memoryUsage    = flag.Bool("mu", false, "Print memory usage")
		cnameFile      = flag.String("cnamefile", "none", "The file to scan for subdomain takeover")
		dorkFile       = flag.String("dorkfile", "none", "The file to scan for Google dork")
		methodFile     = flag.String("methodfile", "none", "The file to scan for HTTP method checker")
		redirectFile   = flag.String("redirectfile", "none", "The file to scan for redirect checker")
		hopperFile     = flag.String("hopperfile", "none", "The file to scan for hopper")
		dispatcherFile = flag.String("dispatcherfile", "none", "The file to scan for dispatcher")
	)
	flag.Parse()
	return &common.Opts{
		Module:         modules,
		Concurrency:    *concurrency,
		Target:         *target,
		CnameFile:      *cnameFile,
		DorkFile:       *dorkFile,
		MethodFile:     *methodFile,
		RedirectFile:   *redirectFile,
		HopperFile:     *hopperFile,
		DispatcherFile: *dispatcherFile,
	}, *logLevel, *cacheTime, *memoryUsage
}
