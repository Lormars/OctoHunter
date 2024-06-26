package parser

import (
	"flag"

	"github.com/lormars/octohunter/common"
)

var modules common.ModuleList

func Parse_Options() (*common.Opts, *common.Config) {

	flag.Var(&modules, "modules", "The module to run")
	var (
		target         = flag.String("target", "none", "The target to scan")
		cacheTime      = flag.Int("cache", 60, "The cache time to use")
		concurrency    = flag.Int("concurrency", 100, "The concurrency to use")
		logLevel       = flag.String("loglevel", "info", "The log level to use")
		memoryUsage    = flag.Bool("mu", false, "Print memory usage")
		purgeBroker    = flag.Bool("pb", false, "Purge the broker")
		useProxy       = flag.Bool("px", false, "Use proxy")
		ratelimit      = flag.Int("ratelimit", 4, "The rate limit to use per second")
		cnameFile      = flag.String("cnamefile", "none", "The file to scan for subdomain takeover")
		dorkFile       = flag.String("dorkfile", "none", "The file to scan for Google dork")
		methodFile     = flag.String("methodfile", "none", "The file to scan for HTTP method checker")
		redirectFile   = flag.String("redirectfile", "none", "The file to scan for redirect checker")
		hopperFile     = flag.String("hopperfile", "none", "The file to scan for hopper")
		dispatcherFile = flag.String("dispatcherfile", "list/distest", "The file to scan for dispatcher")
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
		}, &common.Config{
			Loglevel:    *logLevel,
			CacheTime:   *cacheTime,
			MemoryUsage: *memoryUsage,
			RateLimit:   *ratelimit,
			PurgeBroker: *purgeBroker,
			UseProxy:    *useProxy,
		}
}
