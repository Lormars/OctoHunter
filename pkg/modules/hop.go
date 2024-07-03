package modules

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/comparer"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/octohunter/internal/notify"
)

func CheckHop(ctx context.Context, wg *sync.WaitGroup, options *common.Opts) {
	defer wg.Done()
	if options.Target != "none" {
		SingleHopCheck(options)
	} else {
		multiplex.Conscan(ctx, SingleHopCheck, options, options.HopperFile, "hop", 10)
	}
}

func SingleHopCheck(options *common.Opts) {
	if !cacher.CheckCache(options.Target, "hop") {
		return
	}
	logger.Debugln("SingleHopCheck module running")
	controlReq, err := http.NewRequest("GET", options.Target, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return
	}
	treatmentReq, err := http.NewRequest("GET", options.Target, nil)
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
	result, place := comparer.CompareResponse(controlResp, treatmentResp)
	if !result && place == "status" {
		if treatmentResp.StatusCode < 400 && controlResp.StatusCode != 429 {
			if checker.CheckAccess(treatmentResp) {
				common.CrawlP.PublishMessage(treatmentResp)
			}
			msg := fmt.Sprintf("[Hop] The responses are different for %s: %d vs %d\n", options.Target, controlResp.StatusCode, treatmentResp.StatusCode)
			color.Red(msg)
			if options.Module.Contains("broker") {
				notify.SendMessage(msg)
				common.OutputP.PublishMessage(msg)
			}
		}
	}
}
