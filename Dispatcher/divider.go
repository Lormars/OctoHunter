package dispatcher

import (
	"strings"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/getter"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/parser"
)

// The central wharehouse where you send your input in.
// It will check its status and send it to the appropriate queue.
// The input must be a serverresult.
func Divider(result *common.ServerResult) {
	urlStr := result.Url
	if !cacher.CheckCache(urlStr, "divider") {
		return
	}

	useHttps := strings.HasPrefix(urlStr, "https")

	//tweaks

	if checker.CheckRedirect(result.StatusCode) {
		//though param splitting does not happen only in redirect, most of it happens here, so...
		go common.SplittingP.PublishMessage(result)
		if useHttps {
			go common.RedirectP.PublishMessage(result.Url)
		}
	} else if checker.CheckRequestError(result.StatusCode) {
		go common.MethodP.PublishMessage(result.Url)
		go common.HopP.PublishMessage(result.Url)
		//if the homepage itself is 404, fuzz for directories
		if result.StatusCode == 404 && checker.CheckHomePage(result.Url) {
			saveDomainToMap(result.Url)

		}
	} else if checker.CheckAccess(result) {
		if result.Depth < 5 { //limit depth
			//crawler should get its input mostly from other modules
			//instead of getting it from crawler itself to avoid reinforcement loop,
			//which would lead to memory explosion no matter what...
			go common.CrawlP.PublishMessage(result)
		}

		if checker.CheckHomePage(result.Url) {
			go common.RCP.PublishMessage(result.Url)
		}

		//go common.PathConfuseP.PublishMessage(result.Url)

		//check path confusion
	}

	//module-specific checks irrelevant to the current status
	if strings.Contains(result.Url, "/aura") || strings.Contains(result.Url, "/s/") || strings.Contains(result.Url, "/sfsites/") {
		go common.SalesforceP.PublishMessage(result.Url)
	}
	contentType := result.Headers.Get("Content-Type")
	if checker.CheckMimeType(contentType, "application/json") {
		go common.CorsP.PublishMessage(result)
	}

	if result.StatusCode != 404 {
		parser.UrlToMap(result.Url)
	}

	//quirks check
	go common.QuirksP.PublishMessage(result)

}

func saveDomainToMap(urlStr string) {

	if !cacher.CheckCache(urlStr, "fuzz404") {
		return
	}

	domain, err := getter.GetDomain(urlStr)
	if err != nil {
		logger.Debugf("Error getting domain: %v", err)
		return
	}
	existingSubdomains, _ := common.Domains.LoadOrStore(domain, new(sync.Map))
	existingSubdomainsMap := existingSubdomains.(*sync.Map)
	existingSubdomainsMap.Store(urlStr, true)
	go common.Fuzz404P.PublishMessage(urlStr)

}
