package crawler

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
	"github.com/lormars/octohunter/internal/parser"
)

// Crawler that does not follow redirect
// but what is it crawling....? like other modules? crawl from a file? or crawl from the result of other modules?
func Crawl(response *common.ServerResult) {
	// Crawl the web
	if !cacher.CheckCache(response.Url, "crawl") {
		return
	}

	logger.Debugf("Crawler running on %s\n", response.Url)
	urls := parser.ExtractUrls(response.Url, response.Body)
	var wg sync.WaitGroup

	pattern := `window\.location\.href\s*=\s*|window\.location\s*=\s*|location\s*=\s*|location\.href\s*=\s*`
	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Warnf("Error compiling regex: %v\n", err)
	}

	for _, url := range urls {
		if !cacher.CheckCache(url, "crawl") {
			continue
		}

		wg.Add(1)
		go func(url string) {
			defer wg.Done()

			if strings.HasSuffix(url, ".svg") || strings.HasSuffix(url, ".png") || strings.HasSuffix(url, ".jpg") || strings.HasSuffix(url, ".gif") {
				common.Cl0P.PublishMessage(url)
			} else {
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					logger.Debugf("Error creating request: %v", err)
					return
				}
				resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
				if err != nil {
					logger.Debugf("Error getting response from %s: %v\n", url, err)
					return
				}
				match := re.MatchString(resp.Body)
				if match {
					msg := fmt.Sprintf("[OR Suspect] %s might have a DOM-OR (window.location match) on %s", response.Url, url)
					common.OutputP.PublishMessage(msg)
					notify.SendMessage(msg)
				}
				resp.Depth = response.Depth + 1
				common.DividerP.PublishMessage(resp)
			}
		}(url)
	}

	wg.Wait()
}
