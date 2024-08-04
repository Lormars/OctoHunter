package modules

import (
	"net/http"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/notify"
)

// CheckPathTraversal checks for path traversal vulnerabilities in REST api endpoints.
func CheckPathTraversal(urlStr string) {

	urlStr = strings.TrimRight(urlStr, "/")
	splits := strings.Split(urlStr, "/")
	fileName := splits[len(splits)-1]

	fuzzURL := urlStr + "/%2e%2e%2f" + fileName
	falsePositiveURL := urlStr + "/%2e%2e%2f" + "xubwozi"

	controlReq, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		return
	}
	fuzzReq, err := http.NewRequest("GET", fuzzURL, nil)
	if err != nil {
		return
	}

	falsePositiveReq, err := http.NewRequest("GET", falsePositiveURL, nil)
	if err != nil {
		return
	}

	controlResp, err := checker.CheckServerCustom(controlReq, clients.Clients.GetRandomClient("h0", false, true))
	if err != nil {
		return
	}

	if controlResp.StatusCode == 404 {
		common.AddToCrawlMap(urlStr, "traversal", 404)
		return
	}
	common.AddToCrawlMap(urlStr, "traversal", controlResp.StatusCode)

	//only care about api endpoitns for now
	contentType := controlResp.Headers.Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		return
	}

	fuzzResp, err := checker.CheckServerCustom(fuzzReq, clients.Clients.GetRandomClient("h0", false, true))
	if err != nil {
		return
	}

	if controlResp.StatusCode == fuzzResp.StatusCode && controlResp.Body == fuzzResp.Body {
		falsePositiveResp, err := checker.CheckServerCustom(falsePositiveReq, clients.Clients.GetRandomClient("h0", false, true))
		if err != nil {
			return
		}
		if controlResp.StatusCode == falsePositiveResp.StatusCode && controlResp.Body == falsePositiveResp.Body {
			return
		}
		msg := "[Path Traversal] " + urlStr
		if common.SendOutput {
			common.OutputP.PublishMessage(msg)
		}
		notify.SendMessage(msg)
	}
}
