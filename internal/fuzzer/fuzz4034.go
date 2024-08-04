package fuzzer

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/getter"
	"github.com/lormars/octohunter/internal/notify"
)

func Fuzz4034(inputStr string) {
	if strings.HasPrefix(inputStr, "http") {
		fuzzAllPath(inputStr)
	} else {
		fuzzNewPath(inputStr)
	}
}

// a new 403/404 endpoint is found, fuzz all sibling path to find possible non-404 endpoints
func fuzzAllPath(urlStr string) {
	// logger.Warnf("Debug AllPath input %s", urlStr)
	fuzzPathInput := &common.ServerResult{
		Url:        urlStr,
		StatusCode: 404, //403/404 does not matter, only used to send alert
	}

	go common.FuzzPathP.PublishMessage(fuzzPathInput)

	rootDomain, err := getter.GetDomain(urlStr)
	if err != nil {
		return
	}
	pathMaps, ok := common.Paths.Load(rootDomain)
	if !ok {
		return
	}
	pathMap := pathMaps.(*sync.Map)

	resultMap := make(map[string]*common.ServerResult)
	var mu sync.Mutex

	pathMap.Range(func(original, _ interface{}) bool {
		originalStr := original.(string)
		fuzzPath := strings.TrimRight(urlStr, "/") + originalStr
		// logger.Warnf("Debug AllPath: %s", fuzzPath)
		req, err := http.NewRequest("GET", fuzzPath, nil)
		if err != nil {
			return true
		}
		resp, err := checker.CheckServerCustom(req, clients.Clients.GetRandomClient("h0", false, true))
		if err != nil {
			return true
		}
		if resp.StatusCode != 404 && resp.StatusCode != 403 && resp.StatusCode != 429 && resp.StatusCode != 500 {
			mu.Lock()
			hashed := common.Hash(resp.Body)
			if _, exists := resultMap[hashed]; !exists {
				resultMap[hashed] = resp
				common.DividerP.PublishMessage(resp)
				common.AddToCrawlMap(resp.Url, "fuzz", resp.StatusCode)
				// logger.Warnf("found new endpoint: %s", fuzzPath)
				msg := fmt.Sprintf("[Fuzz Path(SDomain)] Found new endpoint: %s with SC %d", resp.Url, resp.StatusCode)
				if common.SendOutput {
					common.OutputP.PublishMessage(msg)
				}
				notify.SendMessage(msg)
			}
			mu.Unlock()

		}
		time.Sleep(100 * time.Millisecond)

		return true
	})

}

// a new sibling path is found, fuzz all sibling subdomains to find possible non-404 endpoints
func fuzzNewPath(domainWithPath string) {
	// logger.Warnf("Debug NewPath input %s", domainWithPath)
	splited := strings.Split(domainWithPath, "/")
	domain := splited[0]
	subdomainMaps, ok := common.Domains.Load(domain)
	if !ok {
		return
	}
	subdomainMap := subdomainMaps.(*sync.Map)
	subdomainMap.Range(func(original, _ interface{}) bool {
		originalStr := original.(string)
		fuzzPath := strings.TrimRight(strings.ReplaceAll(originalStr, domain, domainWithPath), "/")
		// logger.Warnf("Debug NewPath %s", fuzzPath)
		req, err := http.NewRequest("GET", fuzzPath, nil)
		if err != nil {
			return true
		}
		resp, err := checker.CheckServerCustom(req, clients.Clients.GetRandomClient("h0", false, true))
		if err != nil {
			return true
		}
		if resp.StatusCode != 404 && resp.StatusCode != 403 && resp.StatusCode != 429 && resp.StatusCode != 500 {
			common.AddToCrawlMap(resp.Url, "fuzz", resp.StatusCode)
			common.DividerP.PublishMessage(resp)
			// logger.Warnf("found new endpoint: %s", fuzzPath)
			msg := fmt.Sprintf("[Fuzz Path(SPath)] Found new endpoint: %s with SC %d", resp.Url, resp.StatusCode)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
		}
		return true
	})
}
