package fuzzer

import (
	"net/http"
	"strings"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/getter"
)

func Fuzz404(inputStr string) {
	if strings.HasPrefix(inputStr, "http") {
		fuzzAllPath(inputStr)
	} else {
		fuzzNewPath(inputStr)
	}
}

// a new 404 endpoint is found, fuzz all sibling path to find possible non-404 endpoints
func fuzzAllPath(urlStr string) {
	// logger.Warnf("Debug AllPath input %s", urlStr)
	rootDomain, err := getter.GetDomain(urlStr)
	if err != nil {
		return
	}
	pathMaps, ok := common.Paths.Load(rootDomain)
	if !ok {
		return
	}
	pathMap := pathMaps.(*sync.Map)
	pathMap.Range(func(original, _ interface{}) bool {
		originalStr := original.(string)
		fuzzPath := strings.TrimRight(urlStr, "/") + originalStr
		// logger.Warnf("Debug AllPath: %s", fuzzPath)
		req, err := http.NewRequest("GET", fuzzPath, nil)
		if err != nil {
			return true
		}
		resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
		if err != nil {
			return true
		}
		if resp.StatusCode != 404 {
			common.DividerP.PublishMessage(resp)
			// logger.Warnf("found new endpoint: %s", fuzzPath)
		}

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
		resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
		if err != nil {
			return true
		}
		if resp.StatusCode != 404 {
			common.DividerP.PublishMessage(resp)
			// logger.Warnf("found new endpoint: %s", fuzzPath)
		}
		return true
	})
}
