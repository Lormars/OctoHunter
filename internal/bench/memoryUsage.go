package bench

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
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
			msg := "[MU] Alloc = " + bToMb(m.Alloc) + " MiB." + "\tSys = " + bToMb(m.Sys) + " MiB. "
			msg += fmt.Sprintf("Data: %.6f GB. ", clients.GetTotalDataTransferred())
			msg += fmt.Sprintf("Concurrent: %d. ", clients.GetConcurrentRequests())
			msg += clients.PrintResStats()
			diversity, rate := common.Sliding.GetHostDiversityScore()
			msg += fmt.Sprintf(" Diversity: %.2f. ", diversity)
			msg += fmt.Sprintf(" Rate: %.2f. ", rate)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			sys, err := strconv.Atoi(bToMb(m.Sys))
			if err != nil {
				logger.Debugf("Error converting Alloc to int: %v\n", err)
			}
			if sys > 5000 {
				notify.SendMessage(msg)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func bToMb(b uint64) string {
	return strconv.FormatUint(b/1024/1024, 10)
}
