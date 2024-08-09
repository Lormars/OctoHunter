package modules

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

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

	var wg sync.WaitGroup
	methodResults := make(chan string, len(methods))
	headerResults := make(chan string, len(headers))

	// Method checks
	for _, method := range methods {
		wg.Add(1)
		go func(method string) {
			defer wg.Done()

			if ok, tcode, errTreat := testAccessControl(result.Url, method); ok {
				msg := fmt.Sprintf("[Method] Access control bypassed for target %s using method %s: %d vs. %d\n", result.Url, method, result.StatusCode, tcode)
				methodResults <- msg
			} else if errTreat != nil {
				logger.Debugf("Error testing access control: treatment - %v\n", errTreat)
			}
		}(method)
	}

	// Header checks
	for _, header := range headers {
		wg.Add(1)
		go func(header string) {
			defer wg.Done()

			if ok, payload, sc := checkHeaderOverwrite(result.Url, header); ok {
				msg := fmt.Sprintf("[Method] Access control bypassed for target %s using header %s and payload %s (%d vs. %d)\n", result.Url, header, payload, result.StatusCode, sc)
				headerResults <- msg
			}
		}(header)
	}

	// Process method results
	go func() {
		for msg := range methodResults {
			color.Red(msg)
			notify.SendMessage(msg)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			common.AddToCrawlMap(result.Url, "method", result.StatusCode)
		}
	}()
	// Process header results
	go func() {
		for msg := range headerResults {
			color.Red(msg)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
			break // to avoid flood
		}
	}()
	wg.Wait()
	close(methodResults)
	close(headerResults)
}

func testAccessControl(urlStr, verb string) (bool, int, error) {

	treatmentReq, err := clients.NewRequest(verb, urlStr, nil, clients.Method)
	if err != nil {
		logger.Debugf("Error creating request: %v\n", err)
		return false, 0, err
	}

	treatmentResp, errTreat := checker.CheckServerCustom(treatmentReq, clients.Clients.GetRandomClient("h0", false, true))
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
		octoTreatmentReq := &clients.OctoRequest{
			Request:  treatmentReq,
			Producer: clients.Method,
		}
		treatmentResp, errTreat := checker.CheckServerCustom(octoTreatmentReq, clients.Clients.GetRandomClient("h0", false, true))
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
