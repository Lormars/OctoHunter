package quirks

import (
	"fmt"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/notify"
)

//Quirks is a general scanner that scan for intersting http responses.
//It does not mean that the responses are vulnerabilities, but they are interesting.

var result *common.ServerResult

func CheckQuirks(res *common.ServerResult) {
	//there are just so many websites with the same quirks on all the endpoints under a path,
	//so need to cache a little more agressively to cache the first path as well
	firstPath, err := cacher.GetFirstPath(res.Url)
	if err != nil {
		if !cacher.CheckCache(res.Url, "quirks") {
			return
		}
	} else {
		if !cacher.CheckCache(firstPath, "quirks") {
			return
		}
	}

	result = res
	doubleHTML()
	jsonwithHTML()
	leakenv()
}

func doubleHTML() {
	contentType := result.Headers.Get("Content-Type")
	if contentType == "" {
		return
	}
	if !checker.CheckMimeType(contentType, "text/html") {
		return
	}
	if strings.Count(result.Body, "</html>") > 1 {
		//if result.Depth > 0, it means this url is the result of a crawl
		//then it is worthy to crawl it to get further endpoint
		//if result.Depth = 0, then it must already be crawled by crawler, so no need to crawl it again
		if result.Depth > 0 {
			common.CrawlP.PublishMessage(result)
		}

		msg := fmt.Sprintf("[Quirks] Double HTML in %s", result.Url)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}
}

func jsonwithHTML() {
	contentType := result.Headers.Get("Content-Type")
	if contentType == "" {
		return
	}
	if !checker.CheckMimeType(contentType, "text/html") {
		return
	}
	if strings.HasPrefix(result.Body, "{") || strings.HasPrefix(result.Body, "[") {
		msg := fmt.Sprintf("[Quirks] JSON with HTML mime in %s", result.Url)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}
}

func leakenv() {
	if strings.Count(result.Body, "HTTP_") > 2 {
		msg := fmt.Sprintf("[Quirks] HTTP_ ENV leak in %s", result.Url)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}
}
