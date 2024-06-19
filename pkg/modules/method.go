package modules

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/requester/pkg/runner"
)

func CheckMethod(ctx context.Context, wg *sync.WaitGroup, options *common.Opts) {
	defer wg.Done()
	if options.Target != "none" {
		singleMethodCheck(options)
	} else {
		multiplex.Conscan(ctx, singleMethodCheck, options, options.MethodFile, "method", 10)
	}
}

func singleMethodCheck(options *common.Opts) {
	methods := []string{"POST", "FOO"}
	headers := []string{"X-HTTP-Method-Override", "X-HTTP-Method", "X-Method-Override", "X-Method"}
	for _, method := range methods {
		if testAccessControl(options, method) {
			msg := fmt.Sprintf("[Method] Access control Bypassed for target %s using method %s\n", options.Target, method)
			color.Red(msg)
			if options.Module.Contains("broker") {
				common.PublishMessage(msg)
			}
		}
		time.Sleep(1 * time.Second)
	}
	for _, header := range headers {
		if checkMethodOverwrite(options, header) {
			msg := fmt.Sprintf("[Method] Method Overwrite Bypassed for target %s using header %s\n", options.Target, header)
			color.Red(msg)
			if options.Module.Contains("broker") {
				common.PublishMessage(msg)
			}
		}
		time.Sleep(1 * time.Second)
	}

}

func testAccessControl(options *common.Opts, verb string) bool {
	control_config, err1 := runner.NewConfig(options.Target)
	treatment_config, err2 := runner.NewConfig(options.Target)
	if err1 != nil || err2 != nil {
		return false
	}
	treatment_config.Method = verb
	control_resp, err1 := runner.Run(control_config)
	treatment_resp, err2 := runner.Run(treatment_config)
	if err1 != nil || err2 != nil {
		return false
	}
	if !checker.CheckAccess(control_resp) && checker.CheckAccess(treatment_resp) {
		//to fix equifax false positive
		if !strings.Contains(treatment_resp.Body, "Something went wrong") || !strings.Contains(treatment_resp.Body, "Equifax") {
			return true
		}
	}

	return false

}

func checkMethodOverwrite(options *common.Opts, header string) bool {
	control_config, err1 := runner.NewConfig(options.Target)
	treatment_config, err2 := runner.NewConfig(options.Target)
	if err1 != nil || err2 != nil {
		return false
	}
	control_config.Method = "DELETE"
	treatment_config.Method = "DELETE"
	treatment_config.Header_input = fmt.Sprintf("%s: GET", header)
	control_resp, err1 := runner.Run(control_config)
	treatment_resp, err2 := runner.Run(treatment_config)
	if err1 != nil || err2 != nil {
		return false
	}
	if checker.Check405(control_resp) && !checker.Check405(treatment_resp) && !checker.Check429(treatment_resp) {
		return true
	}

	return false

}
