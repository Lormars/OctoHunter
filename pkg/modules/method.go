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
	//headers := []string{"X-HTTP-Method-Override", "X-HTTP-Method", "X-Method-Override", "X-Method"}
	for _, method := range methods { //FIXME: not optimal, do not need to send GET request for each method
		if ok, ccode, tcode, errCtrl, errTreat := testAccessControl(options, method); ok {
			if ccode == 429 {
				time.Sleep(5 * time.Second)
				continue
			}
			msg := fmt.Sprintf("[Method] Access control Bypassed for target %s using method %s: %d vs. %d \n", options.Target, method, ccode, tcode)
			color.Red(msg)
			if options.Module.Contains("broker") {
				notify.SendMessage(msg)
				common.OutputP.PublishMessage(msg)
			}
		} else if errCtrl != nil || errTreat != nil {
			logger.Debugf("Error testing access control: control - %v | treament - %v\n", errCtrl, errTreat)
			break
		}
	}
	//temporary disable due to high false positive
	// for _, header := range headers {
	// 	if ok, errCtrl, errTreat := checkMethodOverwrite(options, header); ok {
	// 		msg := fmt.Sprintf("[Method] Method Overwrite Bypassed for target %s using header %s\n", options.Target, header)
	// 		color.Red(msg)
	// 		if options.Module.Contains("broker") {
	// 			common.OutputP.PublishMessage(msg)
	// 			notify.SendMessage(msg)
	// 		}
	// 	} else if errCtrl != nil || errTreat != nil {
	// 		logger.Debugf("Error testing method overwrite: control - %v | treament - %v\n", errCtrl, errTreat)
	// 		break
	// 	}
	// }

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
			return true, controlResp.StatusCode, treatmentResp.StatusCode, nil, nil
		}
	}

	return false, 0, 0, nil, nil

}

func checkMethodOverwrite(options *common.Opts, header string) (bool, error, error) {
	controlReq, err := http.NewRequest("DELETE", options.Target, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return false, err, nil
	}
	treatmentReq, err := http.NewRequest("DELETE", options.Target, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return false, err, nil
	}
	treatmentReq.Header.Set(header, "GET")
	controlResp, errCtrl := checker.CheckServerCustom(controlReq, clients.NoRedirectClient)
	treatmentResp, errTreat := checker.CheckServerCustom(treatmentReq, clients.NoRedirectClient)
	if errCtrl != nil || errTreat != nil {
		logger.Debugf("Error getting response: control - %v | treament - %v\n", errCtrl, errTreat)
		return false, errCtrl, errTreat
	}
	if checker.Check405(controlResp) && !checker.Check405(treatmentResp) && !checker.Check429(treatmentResp) {
		return true, nil, nil
	}

	return false, nil, nil

}
