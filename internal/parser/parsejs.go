package parser

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/BishopFox/jsluice"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

func ParseJS(result *common.ServerResult) {
	analyzer := jsluice.NewAnalyzer([]byte(result.Body))

	for _, url := range analyzer.GetURLs() {

		resolvedURL, err := resolveURL(result.Url, url.URL)
		if err != nil {
			continue
		}
		if url.Method == "GET" {

			req, err := http.NewRequest("GET", resolvedURL, nil)
			if err != nil {
				logger.Debugf("Error creating request: %v", err)
				continue
			}
			resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
			if err != nil {
				logger.Debugf("Error getting response from %s: %v\n", url.URL, err)
				continue
			}
			common.AddToCrawlMap(resolvedURL, "jsParse", resp.StatusCode)
			common.CrawlP.PublishMessage(resp)
			if strings.Contains(resolvedURL, "EXPR") {
				msg := fmt.Sprintf("[JS DOM] %s in %s", resolvedURL, result.Url)
				if common.SendOutput {
					common.OutputP.PublishMessage(msg)
				}
				notify.SendMessage(msg)
			}

			if strings.Contains(resolvedURL, "api") && !strings.HasSuffix(resolvedURL, ".js") && !strings.HasSuffix(resolvedURL, ".css") {
				common.PathTraversalP.PublishMessage(resolvedURL)
				common.FuzzAPIP.PublishMessage(resolvedURL)
			}
		} else {
			if strings.Contains(resolvedURL, "graphql") {
				common.GraphqlP.PublishMessage(resolvedURL)
			} else {
				//this is used to fuzz and test for API
				if !strings.HasSuffix(resolvedURL, ".js") && !strings.HasSuffix(resolvedURL, ".css") {
					common.PathTraversalP.PublishMessage(resolvedURL)
					common.FuzzAPIP.PublishMessage(resolvedURL)
				}
			}
		}
	}
}
