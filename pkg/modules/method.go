package modules

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/octohunter/internal/notify"
)

func CheckMethod(ctx context.Context, wg *sync.WaitGroup, options *common.Opts) {
	defer wg.Done()
	if options.Target != "none" {
		SingleMethodCheck(options)
	} else {
		multiplex.Conscan(ctx, SingleMethodCheck, options, options.MethodFile, "method", 10)
	}
}

func SingleMethodCheck(options *common.Opts) {
	if !cacher.CheckCache(options.Target, "method") {
		return
	}
	logger.Debugln("SingleMethodCheck module running")
	methods := []string{"POST", "FOO"}
	headers := []string{"X-Forwarded-For", "X-Forward-For", "X-Remote-IP", "X-Originating-IP", "X-Remote-Addr", "X-Client-IP"}
	for _, method := range methods { //FIXME: not optimal, do not need to send GET request for each method
		if ok, ccode, tcode, errCtrl, errTreat := testAccessControl(options, method); ok {
			if ccode == 429 {
				time.Sleep(5 * time.Second)
				continue
			}
			msg := fmt.Sprintf("[Method] Access control Bypassed for target %s using method %s: %d vs. %d \n", options.Target, method, ccode, tcode)
			color.Red(msg)
			notify.SendMessage(msg)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			common.AddToCrawlMap(options.Target, "method", ccode)
		} else if errCtrl != nil || errTreat != nil {
			logger.Debugf("Error testing access control: control - %v | treament - %v\n", errCtrl, errTreat)
			continue
		}
	}
	for _, header := range headers {
		if ok, payload, sc := checkHeaderOverwrite(options, header); ok {
			msg := fmt.Sprintf("[Method] Access Control Bypassed for target %s using header %s and payload %s (%d)\n", options.Target, header, payload, sc)
			color.Red(msg)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
			break //to avoid flood
		}
	}

}

func testAccessControl(options *common.Opts, verb string) (bool, int, int, error, error) {
	controlReq, err := http.NewRequest("GET", options.Target, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return false, 0, 0, err, nil
	}
	treatmentReq, err := http.NewRequest(verb, options.Target, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return false, 0, 0, nil, err
	}

	controlResp, errCtrl := checker.CheckServerCustom(controlReq, clients.NoRedirectClient)
	treatmentResp, errTreat := checker.CheckServerCustom(treatmentReq, clients.NoRedirectClient)
	if errCtrl != nil || errTreat != nil {
		logger.Debugf("Error getting response: control - %v | treament - %v\n", errCtrl, errTreat)
		return false, 0, 0, errCtrl, errTreat
	}
	if !checker.CheckAccess(controlResp) && checker.CheckAccess(treatmentResp) {
		//to fix equifax false positive
		if !strings.Contains(treatmentResp.Body, "Something went wrong") || !strings.Contains(treatmentResp.Body, "Equifax") {
			//to fix other false positive
			if !strings.Contains(treatmentResp.Body, "request blocked") {
				return true, controlResp.StatusCode, treatmentResp.StatusCode, nil, nil
			}
		}
	}

	return false, 0, 0, nil, nil

}

func checkHeaderOverwrite(options *common.Opts, header string) (bool, string, int) {
	treatmentReq, err := http.NewRequest("GET", options.Target, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return false, "", 0
	}

	payloads := []string{"127.0.0.1", "localhost"}

	for _, payload := range payloads {
		treatmentReq.Header.Set(header, payload)
		treatmentResp, errTreat := checker.CheckServerCustom(treatmentReq, clients.NoRedirectClient)
		if errTreat != nil {
			logger.Debugf("Error getting response: treament - %v\n", errTreat)
			continue
		}
		if treatmentResp.StatusCode < 400 {
			return true, payload, treatmentResp.StatusCode
		}
	}
	return false, "", 0

}
