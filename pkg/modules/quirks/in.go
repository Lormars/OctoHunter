package quirks

import (
	"fmt"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/notify"
)

//Quirks is a general scanner that scan for intersting http responses.
//It does not mean that the responses are vulnerabilities, but they are interesting.

var result *common.ServerResult

func CheckQuirks(res *common.ServerResult) {
	result = res
	doubleHTML()
	jsonwithHTML()
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
