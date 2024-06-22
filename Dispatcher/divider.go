package dispatcher

import (
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

// The central wharehouse where you send your input in.
// It will check its status and send it to the appropriate queue.
func Divider(domainString string) {

	if !strings.HasPrefix(domainString, "http://") && !strings.HasPrefix(domainString, "https://") {
		if !checker.ResolveDNS(domainString) {
			logger.Debugln("DNS resolution failed for: ", domainString)
			go common.CnameP.PublishMessage(domainString)
			return
		}

		go common.CnameP.PublishMessage(domainString)
	}
	//ugly, but for now...
	domainString = strings.TrimPrefix(domainString, "http://")
	domainString = strings.TrimPrefix(domainString, "https://")
	if !cacher.CheckCache(domainString, "divider") {
		return
	}
	httpStatus, httpsStatus, errhttp, errhttps := checker.CheckHTTPAndHTTPSServers(domainString)

	//tweaks
	httpsCrawled := false //to avoid duplicate crawl of same endpoint under different protocol

	if errhttps != nil {
		logger.Debugf("Error checking https server: %v\n", errhttps)
	} else if httpsStatus.Online {
		if checker.CheckRedirect(httpsStatus.StatusCode) {
			go common.RedirectP.PublishMessage(httpsStatus.Url)
		} else if checker.CheckRequestError(httpsStatus.StatusCode) {
			go common.MethodP.PublishMessage(httpsStatus.Url)
			go common.HopP.PublishMessage(httpsStatus.Url)
		} else if checker.CheckAccess(httpsStatus) {
			go common.CrawlP.PublishMessage(httpsStatus)
			httpsCrawled = true
		}

		//module-specific checks irrelevant to the current status
		if strings.Contains(httpsStatus.Url, "/aura") || strings.Contains(httpsStatus.Url, "/s/") || strings.Contains(httpsStatus.Url, "/sfsites/") {
			go common.SalesforceP.PublishMessage(httpsStatus.Url)
		}
	}
	if errhttp != nil {
		logger.Debugf("Error checking http server: %v\n", errhttp)
	} else if httpStatus.Online {
		if checker.CheckRequestError(httpStatus.StatusCode) {
			go common.MethodP.PublishMessage(httpStatus.Url)
			go common.HopP.PublishMessage(httpStatus.Url)
		} else if checker.CheckAccess(httpStatus) && !httpsCrawled {
			go common.CrawlP.PublishMessage(httpStatus)
		}
		return
	}

}
