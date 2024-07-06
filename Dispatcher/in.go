package dispatcher

import (
	"bufio"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

func Input(opts *common.Opts) {
	Init(opts)
	for {
		time.Sleep(5 * time.Second)
		file, err := os.Open(opts.DispatcherFile)
		if err != nil {
			logger.Errorln("Error opening file: ", err)
			return
		}

		lineCh := make(chan string, opts.Concurrency)
		var wg sync.WaitGroup
		for i := 0; i < opts.Concurrency; i++ {
			wg.Add(1)
			go func() {
				for domainString := range lineCh {
					if !strings.HasPrefix(domainString, "http") {
						if !checker.ResolveDNS(domainString) {
							logger.Debugln("DNS resolution failed for: ", domainString)
							go common.CnameP.PublishMessage(domainString)
							continue
						}

						go common.CnameP.PublishMessage(domainString)
					}
					domainString = strings.TrimPrefix(domainString, "http://")
					domainString = strings.TrimPrefix(domainString, "https://")
					httpStatus, httpsStatus, errhttp, errhttps := checker.CheckHTTPAndHTTPSServers(domainString)
					if errhttp == nil && httpStatus.Online {
						go common.DividerP.PublishMessage(httpStatus)
					}
					if errhttps == nil && httpsStatus.Online {
						go common.DividerP.PublishMessage(httpsStatus)
					}
				}
				wg.Done()
			}()
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			lineCh <- line
		}

		if err := scanner.Err(); err != nil {
			logger.Errorln("Error reading file: ", err)
		}
		file.Close()
		close(lineCh)
		wg.Wait()
		common.OutputP.PublishMessage("Finished processing all domains")
		notify.SendMessage("Finished processing all domains")
		time.Sleep(1 * time.Hour)
	}
}
