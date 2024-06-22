package bench

import (
	"runtime"
	"strconv"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

func PrintMemUsage(opts *common.Opts) {
	time.Sleep(5 * time.Second)
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// For more info, see: https://golang.org/pkg/runtime/#MemStats
		logger.Debugf("Alloc = %v MiB", bToMb(m.Alloc))
		logger.Debugf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
		logger.Debugf("\tSys = %v MiB", bToMb(m.Sys))
		logger.Debugf("\tNumGC = %v\n", m.NumGC)
		if opts.Module.Contains("broker") {
			msg := "[MU] Alloc = " + bToMb(m.Alloc) + " MiB." + "\tTotalAlloc = " + bToMb(m.TotalAlloc) + " MiB." + "\tSys = " + bToMb(m.Sys) + " MiB."
			common.OutputP.PublishMessage(msg)
			alloc, err := strconv.Atoi(bToMb(m.Sys))
			if err != nil {
				logger.Debugf("Error converting Alloc to int: %v\n", err)
			}
			if alloc > 5000 {
				notify.SendMessage(msg)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func bToMb(b uint64) string {
	return strconv.FormatUint(b/1024/1024, 10)
}
