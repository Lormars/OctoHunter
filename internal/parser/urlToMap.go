package parser

import (
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/getter"
	"github.com/lormars/octohunter/internal/logger"
)

// It stores path and domain to map.
// It returns nil if there is no error and the domain it updates, err otherwise
func UrlToMap(urlStr string) (string, error) {
	domain, path, err := getter.GetDomainAndFirstPath(urlStr)
	if err != nil {
		logger.Debugf("Error getting domain and path: %v", err)
		return "", err
	}
	existingPaths, _ := common.Paths.LoadOrStore(domain, new(sync.Map))
	pathMap := existingPaths.(*sync.Map)

	pathMap.Store(path, true)
	domainWithPath := domain + path

	common.Fuzz404P.PublishMessage(domainWithPath)

	return domainWithPath, nil
}
