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
		concurrency    = flag.Int("concurrency", 5, "The concurrency to use")
		logLevel       = flag.String("loglevel", "info", "The log level to use")
		memoryUsage    = flag.Bool("mu", false, "Print memory usage")
		purgeBroker    = flag.Bool("pb", false, "Purge the broker")
		useProxy       = flag.Bool("px", false, "Use proxy")
		headless       = flag.Bool("hl", false, "use headless browser")
		ratelimit      = flag.Int("ratelimit", 4, "The rate limit to use per second")
		dispatcherFile = flag.String("dispatcherfile", "list/distest", "The file to scan for dispatcher")
	)
	flag.Parse()
	return &common.Opts{
			Module:         modules,
			Concurrency:    *concurrency,
			Target:         *target,
			DispatcherFile: *dispatcherFile,
		}, &common.Config{
			Loglevel:    *logLevel,
			CacheTime:   *cacheTime,
			MemoryUsage: *memoryUsage,
			RateLimit:   *ratelimit,
			PurgeBroker: *purgeBroker,
			UseProxy:    *useProxy,
			Headless:    *headless,
		}
}
