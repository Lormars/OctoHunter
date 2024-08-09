package parser

import (
	"fmt"
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

		if strings.Contains(resolvedURL, "EXPR") {
			msg := fmt.Sprintf("[JS DOM] %s in %s", resolvedURL, result.Url)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
		}

		req, err := clients.NewRequest("GET", resolvedURL, nil, clients.Crawl)
		if err != nil {
			logger.Debugf("Error creating request: %v", err)
			continue
		}

		resp, err := checker.CheckServerCustom(req, clients.Clients.GetRandomClient("h0", false, true))
		if err != nil {
			logger.Debugf("Error getting response from %s: %v\n", url.URL, err)
			continue
		}
		common.AddToCrawlMap(resolvedURL, "jsParse", resp.StatusCode)
		resp.Depth += 1
		common.DividerP.PublishMessage(resp)
		contentType := resp.Headers.Get("Content-Type")
		jsonOrXML := checker.CheckMimeType(contentType, "application/json") || checker.CheckMimeType(contentType, "application/xml") || checker.CheckMimeType(contentType, "text/xml")
		if url.Method == "GET" {

			if jsonOrXML || (strings.Contains(resolvedURL, "api") && !strings.HasSuffix(resolvedURL, ".js") && !strings.HasSuffix(resolvedURL, ".css")) {
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
