package dispatcher

import (
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

func Divider(domainString string) {

	if !strings.HasPrefix(domainString, "http://") && !strings.HasPrefix(domainString, "https://") {
		if !checker.ResolveDNS(domainString) {
			logger.Debugln("DNS resolution failed for: ", domainString)
			common.CnameP.PublishMessage(domainString)
			return
		}

		common.CnameP.PublishMessage(domainString)
	}
	//ugly, but for now...
	domainString = strings.TrimPrefix(domainString, "http://")
	domainString = strings.TrimPrefix(domainString, "https://")
	if !cacher.CanScan(domainString, "divider") {
		logger.Debugf("Skipping %s\n", domainString)
		return
	}
	cacher.UpdateScanTime(domainString, "divider")
	httpStatus, httpsStatus, errhttp, errhttps := checker.CheckHTTPAndHTTPSServers(domainString)

	//tweaks
	httpsCrawled := false //to avoid duplicate crawl of same endpoint under different protocol

	if errhttps != nil {
		logger.Debugf("Error checking https server: %v\n", errhttps)
	} else if httpsStatus.Online {
		if checker.CheckRedirect(httpsStatus.StatusCode) {
			common.RedirectP.PublishMessage(httpsStatus.Url)
			common.MethodP.PublishMessage(httpsStatus.Url)
		} else if checker.CheckRequestError(httpsStatus.StatusCode) {
			common.HopP.PublishMessage(httpsStatus.Url)
		} else if checker.CheckAccess(httpsStatus) {
			//common.CrawlP.PublishMessage(httpsStatus)
			httpsCrawled = true
		}
	}
	if errhttp != nil {
		logger.Debugf("Error checking http server: %v\n", errhttp)
	} else if httpStatus.Online {
		if checker.CheckRequestError(httpStatus.StatusCode) {
			common.HopP.PublishMessage(httpStatus.Url)
		} else if checker.CheckAccess(httpStatus) && !httpsCrawled {
			//common.CrawlP.PublishMessage(httpStatus)
		}
		return
	}

}
