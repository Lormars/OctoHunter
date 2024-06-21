package dispatcher

import (
	"bufio"
	"os"
	"strings"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

func Input(opts *common.Opts) {
	Init(opts)
	file, err := os.Open(opts.DispatcherFile)
	if err != nil {
		logger.Errorln("Error opening file: ", err)
		return
	}
	defer file.Close()
	lineCh := make(chan string)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for line := range lineCh {
				divider(line)
			}
		}()
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineCh <- line
	}
}

func divider(domainString string) {
	if !checker.ResolveDNS(domainString) {
		logger.Debugln("DNS resolution failed for: ", domainString)
		common.CnameP.PublishMessage(domainString)
		return
	}

	common.CnameP.PublishMessage(domainString)

	httpStatus, httpsStatus, errhttp, errhttps := checker.CheckHTTPAndHTTPSServers(domainString)
	if errhttps != nil {
		logger.Debugf("Error checking https server: %v\n", errhttps)
	} else if httpsStatus.Online {
		if checker.CheckRedirect(httpsStatus.StatusCode) {
			common.RedirectP.PublishMessage(httpsStatus.Url)
			common.MethodP.PublishMessage(httpsStatus.Url)
		} else if checker.CheckRequestError(httpsStatus.StatusCode) {
			common.HopP.PublishMessage(httpsStatus.Url)
		}
	}
	if errhttp != nil {
		logger.Debugf("Error checking http server: %v\n", errhttp)
	} else if httpStatus.Online {
		if checker.CheckRequestError(httpStatus.StatusCode) {
			common.HopP.PublishMessage(httpStatus.Url)
		}
		return
	}

}
