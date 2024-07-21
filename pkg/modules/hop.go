package modules

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/comparer"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

func SingleHopCheck(result *common.ServerResult) {
	if !cacher.CheckCache(result.Url, "hop") {
		return
	}
	logger.Debugln("SingleHopCheck module running")
	controlReq, err := http.NewRequest("GET", result.Url, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return
	}
	treatmentReq, err := http.NewRequest("GET", result.Url, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return
	}

	controlReq.Header.Set("Connection", "close")
	treatmentReq.Header.Set("Connection", "close, X-Forwarded-For")
	controlResp, errCtrl := checker.CheckServerCustom(controlReq, clients.NoRedirecth1Client)
	treatmentResp, errTreat := checker.CheckServerCustom(treatmentReq, clients.NoRedirecth1Client)
	if errCtrl != nil || errTreat != nil {
		logger.Debugf("Error getting response: control - %v | treament - %v\n", errCtrl, errTreat)
		return
	}

	common.AddToCrawlMap(result.Url, "hop", controlResp.StatusCode)

	compareResult, place := comparer.CompareResponse(controlResp, treatmentResp)
	if !compareResult && place == "status" {
		if treatmentResp.StatusCode < 400 && controlResp.StatusCode != 429 {
			if checker.CheckAccess(treatmentResp) {
				common.CrawlP.PublishMessage(treatmentResp)
			}
			if strings.Contains(treatmentResp.Body, "Request Rejected") {
				return
			}
			msg := fmt.Sprintf("[Hop] The responses are different for %s: %d vs %d\n", result.Url, controlResp.StatusCode, treatmentResp.StatusCode)
			color.Red(msg)
			notify.SendMessage(msg)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
		}
	}
}
