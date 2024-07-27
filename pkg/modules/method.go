package modules

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

func SingleMethodCheck(result *common.ServerResult) {

	logger.Debugln("SingleMethodCheck module running")
	methods := []string{"POST", "FOO"}
	headers := []string{"X-Forwarded-For", "X-Forward-For", "X-Remote-IP", "X-Originating-IP", "X-Remote-Addr", "X-Client-IP"}
	for _, method := range methods {
		if ok, tcode, errTreat := testAccessControl(result.Url, method); ok {
			msg := fmt.Sprintf("[Method] Access control Bypassed for target %s using method %s: %d vs. %d \n", result.Url, method, result.StatusCode, tcode)
			color.Red(msg)
			notify.SendMessage(msg)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			common.AddToCrawlMap(result.Url, "method", result.StatusCode)
		} else if errTreat != nil {
			logger.Debugf("Error testing access control: treament - %v\n", errTreat)
			continue
		}
	}
	for _, header := range headers {
		if ok, payload, sc := checkHeaderOverwrite(result.Url, header); ok {
			msg := fmt.Sprintf("[Method] Access Control Bypassed for target %s using header %s and payload %s (%d vs. %d)\n", result.Url, header, payload, result.StatusCode, sc)
			color.Red(msg)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
			break //to avoid flood
		}
	}

}

func testAccessControl(urlStr, verb string) (bool, int, error) {

	treatmentReq, err := http.NewRequest(verb, urlStr, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return false, 0, err
	}

	treatmentResp, errTreat := checker.CheckServerCustom(treatmentReq, clients.NoRedirectClient)
	if errTreat != nil {
		logger.Debugf("Error getting response: treament - %v\n", errTreat)
		return false, 0, errTreat
	}
	if checker.CheckAccess(treatmentResp) {
		//to fix equifax false positive
		if !strings.Contains(treatmentResp.Body, "Something went wrong") || !strings.Contains(treatmentResp.Body, "Equifax") {
			//to fix other false positive
			if !strings.Contains(treatmentResp.Body, "request blocked") {
				return true, treatmentResp.StatusCode, nil
			}
		}
	}

	return false, 0, nil

}

func checkHeaderOverwrite(urlStr, header string) (bool, string, int) {
	treatmentReq, err := http.NewRequest("GET", urlStr, nil)
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
