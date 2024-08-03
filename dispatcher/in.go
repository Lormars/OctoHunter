package dispatcher

import (
	"bufio"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/common/score"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

var scanned int64
var allerred int

func Input(opts *common.Opts) {
	Init(opts)

	go func() {
		for {
			time.Sleep(60 * time.Second)
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
		scanned = 0
		lineCh := make(chan string, opts.Concurrency)
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				for domainString := range lineCh {
					logger.Debugf("Processing domain: %s", domainString)
					if !strings.HasPrefix(domainString, "http") {
						logger.Debugf("Domain does not have http/https prefix: %s", domainString)
						go common.CnameP.PublishMessage(domainString)
					}

					_, err := clients.DnsCache.LookupIP(domainString)
					if err != nil {
						continue
					}
					atomic.AddInt64(&scanned, 1)
					domainString = strings.TrimPrefix(domainString, "http://")
					domainString = strings.TrimPrefix(domainString, "https://")
					httpStatus, httpsStatus, errhttp, errhttps := checker.CheckHTTPAndHTTPSServers(domainString)
					//why? to make sure the statuscode is right.
					//using the default client may be blocked due to various bot checks.
					//So need to use our client to request again to make sure the status code is right.
					if errhttp != nil && errhttps != nil {
						allerred++
					}
					go func() {
						if errhttp == nil && httpStatus.Online {
							req, err := http.NewRequest("GET", httpStatus.Url, nil)
							if err == nil {
								resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
								if err == nil {
									go common.DividerP.PublishMessage(resp)
								} else {
									logger.Debugf("Error checking http server: %v", err)
								}
							} else {
								logger.Debugf("Error checking http server: %v", err)
							}
						} else {
							logger.Debugf("Error checking http server: %v", errhttp)
						}
					}()
					go func() {
						if (errhttps == nil && httpsStatus.Online) || (errhttps != nil && strings.Contains(errhttps.Error(), "browser check")) {
							if httpsStatus == nil {
								logger.Warnf("shouldnt happen: %v", errhttps)
								return
							}
							req, err := http.NewRequest("GET", httpsStatus.Url, nil)
							if err == nil {
								resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)

								// resp, err := checker.CheckServerCustom(req, http.DefaultClient)
								if err == nil {
									go common.DividerP.PublishMessage(resp)
								} else {
									logger.Debugf("Error checking https server: %v", err)
								}
							} else {
								logger.Debugf("Error checking https server: %v", err)
							}
						} else {
							logger.Debugf("Error checking https server: %v", errhttps)
						}
					}()
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
		if common.SendOutput {
			common.OutputP.PublishMessage("Finished processing all domains")
		}
		notify.SendMessage("Finished processing all domains")
		time.Sleep(1 * time.Hour)
	}
}

func GetScanned() int64 {
	totalScanned := scanned
	return totalScanned
}

func GetAllerred() int {
	totalAllerred := allerred
	return totalAllerred
}
