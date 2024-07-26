package bench

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	dispatcher "github.com/lormars/octohunter/dispatcher"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

var notified = false

func PrintMemUsage(opts *common.Opts) {
	start := time.Now()
	time.Sleep(5 * time.Second)
	for {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		// For more info, see: https://golang.org/pkg/runtime/#MemStats
		// logger.Debugf("Alloc = %v MiB", bToMb(m.Alloc))
		// logger.Debugf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
		// logger.Debugf("\tSys = %v MiB", bToMb(m.Sys))
		// logger.Debugf("\tNumGC = %v\n", m.NumGC)
		if opts.Module.Contains("broker") {
			msg := "[MU] Alloc = " + bToMb(m.Alloc) + " MiB." + "\tSys = " + bToMb(m.Sys) + " MiB. "
			msg += fmt.Sprintf("Data: %.6f GB. ", clients.GetTotalDataTransferred())
			msg += fmt.Sprintf("Con: %d. ", clients.GetConcurrentRequests())
			msg += clients.PrintResStats()
			diversity, rate := common.Sliding.GetHostDiversityScore()
			msg += fmt.Sprintf(" Div: %.2f. ", diversity)
			msg += fmt.Sprintf(" Rate: %.2f. ", rate)
			scanned := dispatcher.GetScanned()
			msg += fmt.Sprintf(" Scanned: %d. ", scanned)
			all429 := clients.Get429Count()
			msg += fmt.Sprintf(" Slowed: %d. ", all429)
			consumerUsage := len(common.ConsumerSemaphore)
			msg += fmt.Sprintf(" CUsage: %d. ", consumerUsage)
			browserUsage := len(common.NeedBrowser)
			msg += fmt.Sprintf(" Browser: %d. ", browserUsage)
			elapsed := time.Since(start)
			msg += fmt.Sprintf(" Elapsed: %.2f. ", elapsed.Minutes())
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			sys, err := strconv.Atoi(bToMb(m.Sys))
			if err != nil {
				logger.Debugf("Error converting Alloc to int: %v\n", err)
			}
			if sys > 7000 && !notified {
				notified = true
				notify.SendMessage(msg)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

func bToMb(b uint64) string {
	return strconv.FormatUint(b/1024/1024, 10)
}
