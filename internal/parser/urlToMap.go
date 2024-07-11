package parser

import (
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/getter"
	"github.com/lormars/octohunter/internal/logger"
)

// It stores path and domain to map.
func UrlToMap(urlStr string) {
	domain, path, err := getter.GetDomainAndFirstPath(urlStr)
	if err != nil {
		logger.Debugf("Error getting domain and path: %v", err)
		return
	} else if domain == "" || path == "" {
		logger.Debugf("Empty domain or path")
		return
	}
	existingPaths, _ := common.Paths.LoadOrStore(domain, new(sync.Map))
	pathMap := existingPaths.(*sync.Map)

	pathMap.Store(path, true)
	domainWithPath := domain + path

	if !cacher.CheckCache(domainWithPath, "fuzz404") {
		return
	}

	common.Fuzz4034P.PublishMessage(domainWithPath)

}
