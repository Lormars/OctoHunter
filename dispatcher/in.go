package dispatcher

import (
	"bufio"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/common/score"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

var scanned int

func Input(opts *common.Opts) {
	Init(opts)

	go func() {
		for {
			time.Sleep(10 * time.Second)
			score.CalculateScore()
		}
	}()

	for {
		time.Sleep(5 * time.Second)
		file, err := os.Open(opts.DispatcherFile)
		if err != nil {
			logger.Errorln("Error opening file: ", err)
			return
		}

		lineCh := make(chan string, opts.Concurrency)
		var wg sync.WaitGroup
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				for domainString := range lineCh {
					if !strings.HasPrefix(domainString, "http") {
						go common.CnameP.PublishMessage(domainString)
					}
					domainString = strings.TrimPrefix(domainString, "http://")
					domainString = strings.TrimPrefix(domainString, "https://")
					httpStatus, httpsStatus, errhttp, errhttps := checker.CheckHTTPAndHTTPSServers(domainString)
					//why? to make sure the statuscode is right.
					//using the default client may be blocked due to various bot checks.
					//So need to use our client to request again to make sure the status code is right.
					if errhttp == nil && httpStatus.Online {
						req, err := http.NewRequest("GET", httpStatus.Url, nil)
						if err == nil {
							resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
							if err == nil {
								go common.DividerP.PublishMessage(resp)
							}
						}
					}
					if errhttps == nil && httpsStatus.Online {
						req, err := http.NewRequest("GET", httpsStatus.Url, nil)
						if err == nil {
							resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
							if err == nil {
								go common.DividerP.PublishMessage(resp)
							}
						}
					}
				}
				wg.Done()
			}()
		}
		scanned = 0
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			lineCh <- line
			scanned++
		}

		if err := scanner.Err(); err != nil {
			logger.Errorln("Error reading file: ", err)
		}
		file.Close()
		close(lineCh)
		wg.Wait()
		if common.SendOutput {
			common.OutputP.PublishMessage("Finished processing all domains")
		}
		notify.SendMessage("Finished processing all domains")
		time.Sleep(1 * time.Hour)
	}
}

func GetScanned() int {
	return scanned
}
