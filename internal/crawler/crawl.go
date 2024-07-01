package crawler

import (
	"net/http"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/parser"
)

// Crawler that does not follow redirect
// but what is it crawling....? like other modules? crawl from a file? or crawl from the result of other modules?
func Crawl(response *common.ServerResult) {
	// Crawl the web
	logger.Debugf("Crawler running on %s\n", response.Url)
	urls := parser.ExtractUrls(response.Url, response.Body)
	for _, url := range urls {
		if !cacher.CheckCache(url, "crawl") {
			continue
		}
		if strings.HasSuffix(url, ".svg") || strings.HasSuffix(url, ".png") || strings.HasSuffix(url, ".jpg") || strings.HasSuffix(url, ".gif") {
			common.Cl0P.PublishMessage(url)
		} else {
			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				logger.Debugf("Error creating request: %v", err)
				continue
			}
			resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
			if err != nil {
				logger.Debugf("Error getting response from %s: %v\n", url, err)
				continue
			}
			common.DividerP.PublishMessage(resp)
		}
	}
}
