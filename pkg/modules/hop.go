package modules

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/comparer"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/multiplex"
)

func CheckHop(ctx context.Context, wg *sync.WaitGroup, options *common.Opts) {
	defer wg.Done()
	if options.Target != "none" {
		singleCheck(options)
	} else {
		multiplex.Conscan(ctx, singleCheck, options, options.HopperFile, "hop", 10)
	}
}

func singleCheck(options *common.Opts) {
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
	controlResp, errCtrl := checker.CheckServerCustom(controlReq, common.NoRedirectHTTP1Client)
	treatmentResp, errTreat := checker.CheckServerCustom(treatmentReq, common.NoRedirectHTTP1Client)
	if errCtrl != nil || errTreat != nil {
		logger.Debugf("Error getting response: control - %v | treament - %v\n", errCtrl, errTreat)
		return
	}
	result, place := comparer.CompareResponse(controlResp, treatmentResp)
	if !result && place == "status" {
		if treatmentResp.StatusCode < 400 {
			msg := fmt.Sprintf("[Hop] The responses are different for %s: %d vs %d\n", options.Target, controlResp.StatusCode, treatmentResp.StatusCode)
			color.Red(msg)
			if options.Module.Contains("broker") {
				common.OutputP.PublishMessage(msg)
			}
		}
	}
}
