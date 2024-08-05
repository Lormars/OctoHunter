package quirks

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"encoding/json"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

type Issue struct {
	Type     string `json:"type"`
	Location struct {
		Line   int `json:"line"`
		Column int `json:"column"`
	} `json:"location"`
}

type ParseResponse struct {
	Issues  []Issue `json:"issues,omitempty"` // Use omitempty to handle absence gracefully
	Message string  `json:"message,omitempty"`
}

func CheckJSQuirks(result *common.ServerResult) {
	//skip common libraries
	if strings.Contains(result.Url, "jquery") || strings.Contains(result.Url, "bootstrap") || strings.Contains(result.Url, "angular") || strings.Contains(result.Url, "react") || strings.Contains(result.Url, "vue") {
		return
	}

	//logger.Warnf("Checking JS quirks for %s", result.Url)
	jsCode := result.Body
	parseServerURL := "http://localhost:9999/parse"
	parseReq, err := http.NewRequest("POST", parseServerURL, bytes.NewBufferString(jsCode))
	if err != nil {
		return
	}
	parseReq.Header.Set("Content-Type", "text/plain")
	resp, err := checker.CheckServerCustom(parseReq, clients.Clients.GetRandomClient("h1NA", false, false))
	if err != nil {
		logger.Warnf("Error getting response from %s: %v\n", parseServerURL, err)
		return
	}
	var response ParseResponse

	if err := json.Unmarshal([]byte(resp.Body), &response); err != nil {
		logger.Warnf("Error unmarshalling response: %v\nBody: %s", err, resp.Body)
		return
	}
	if len(response.Issues) > 0 {
		for _, issue := range response.Issues {
			msg := fmt.Sprintf("[Quirks JS] insecure postmessage listener in %s on line %d column %d", result.Url, issue.Location.Line, issue.Location.Column)
			color.Red(msg)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
		}
	}
	// logger.Warnf("JS quirks check for %s done", result.Url)
}
