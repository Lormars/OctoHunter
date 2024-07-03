package modules

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/getter"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

func CheckCors(response *common.ServerResult) {
	req, err := http.NewRequest("GET", response.Url, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v", err)
		return
	}
	baseDomain, err := getter.GetBaseDomain(response.Url)
	if err != nil {
		logger.Debugf("Error getting base domain: %v", err)
		return
	}
	//to bypass startwith and endwith checks
	payload := fmt.Sprintf("https://%s.example%s", baseDomain, baseDomain)

	req.Header.Set("Origin", payload)
	resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
	if err != nil {
		logger.Debugf("Error getting response: %v", err)
		return
	}
	allowAccess := resp.Headers.Get("Access-Control-Allow-Origin")
	if strings.Contains(allowAccess, "example") {
		AllowCredentials := resp.Headers.Get("Access-Control-Allow-Credentials")
		if strings.Contains(AllowCredentials, "true") {
			msg := fmt.Sprintf("[CORS Confirmed] on %s\n", response.Url)
			color.Red(msg)
			common.OutputP.PublishMessage(msg)
			notify.SendMessage(msg)
		}
	}
}
