package dispatcher

import (
	"bufio"
	"os"

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
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		divider(line)
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
		}
	}
	if errhttp != nil {
		logger.Debugf("Error checking http server: %v\n", errhttp)
	} else if httpStatus.Online {
	}

}
