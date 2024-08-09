package crawler

import (
	"strings"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/filter"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/parser"
)

// Crawler that does not follow redirect
// but what is it crawling....? like other modules? crawl from a file? or crawl from the result of other modules?
func Crawl(response *common.ServerResult) {
	// Crawl the web

	logger.Debugf("Crawler running on %s\n", response.Url)
	rawUrls := parser.ExtractUrls(response.Url, response.Body)
	urls := filter.GroupAndFilterURLs(rawUrls)
	logger.Debugf("Urls reduced from %d to %d\n", len(rawUrls), len(urls))
	var wg sync.WaitGroup
	var semaphore = make(chan struct{}, 10)

	for structure, url := range urls {
		if !cacher.CheckCache(structure, "crawl") { //check using structure to prevent crawling too much
			continue
		}

		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if strings.HasSuffix(url, ".svg") || strings.HasSuffix(url, ".png") || strings.HasSuffix(url, ".jpg") || strings.HasSuffix(url, ".gif") || strings.HasSuffix(url, ".jpeg") {
				common.Cl0P.PublishMessage(url)
			} else {
				req, err := clients.NewRequest("GET", url, nil, clients.Crawl)
				if err != nil {
					logger.Debugf("Error creating request: %v", err)
					return
				}
				resp, err := checker.CheckServerCustom(req, clients.Clients.GetRandomClient("h0", false, true))
				if err != nil {
					logger.Debugf("Error getting response from %s: %v\n", url, err)
					return
				}

				resp.Depth = response.Depth + 1
				if strings.HasSuffix(resp.Url, ".js") {
					parser.ParseJS(resp)
				}

				common.AddToCrawlMap(resp.Url, "crawl", resp.StatusCode)
				common.DividerP.PublishMessage(resp)
			}
		}(url)
	}

	wg.Wait()
}
