package api

import (
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

var payloads = []string{"/graphql/v1", "/graphql", "/api", "/api/graphql", "/graphql/api", "/graphql/graphql"}

var introspect = `{"query": "{__schema{queryType{name}}}"}`

func CheckGraphql(urlStr string) {
	if !strings.Contains(urlStr, "graphql") {
		target := strings.TrimRight(urlStr, "/")
		for _, payload := range payloads {
			testURL := target + payload
			req, err := http.NewRequest("GET", testURL, nil)
			if err != nil {
				logger.Warnf("Error creating request: %v", err)
				continue
			}
			_, err = checker.CheckServerCustom(req, clients.NoRedirectClient)
			if err != nil {
				logger.Debugf("Error checking server: %v", err)
				continue
			}
			go checkIntrospect(testURL)
		}
	} else { // if the URL already contains "graphql" (sent from parsejs.go)
		checkIntrospect(urlStr)
	}
}

func checkIntrospect(urlStr string) {
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(introspect))
	if err != nil {
		logger.Warnf("Error creating request: %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
	if err != nil {
		logger.Debugf("Error checking server: %v", err)
		return
	}
	if strings.Contains(resp.Body, "__schema") {
		msg := "[GQL Introspection] Found introspection query in " + urlStr
		if common.SendOutput {
			common.OutputP.PublishMessage(msg)
		}
		notify.SendMessage(msg)
		color.Red(msg)
	}
}
