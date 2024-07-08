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
			common.CrawlP.PublishMessage(resp)
			if strings.Contains(resolvedURL, "EXPR") {
				msg := fmt.Sprintf("[JS DOM] %s in %s", resolvedURL, result.Url)
				common.OutputP.PublishMessage(msg)
				notify.SendMessage(msg)
			}
		} else {
			if strings.Contains(resolvedURL, "graphql") {
				msg := fmt.Sprintf("[GQL Suspect] %s in %s", resolvedURL, result.Url)
				common.OutputP.PublishMessage(msg)
				notify.SendMessage(msg)
			} else {
				msg := fmt.Sprintf("[JS API] %s in %s", resolvedURL, result.Url)
				common.OutputP.PublishMessage(msg)
				notify.SendMessage(msg)
			}
		}
	}
}
