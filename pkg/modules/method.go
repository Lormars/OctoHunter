package modules

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/requester/pkg/runner"
)

func CheckMethod(options *common.Opts) {
	if options.Target != "none" {
		singleMethodCheck(options)
	} else {
		multiMethodCheck(options)
	}
}

func singleMethodCheck(options *common.Opts) {
	methods := []string{"HEAD", "POST", "FOO"}
	headers := []string{"X-HTTP-Method-Override", "X-HTTP-Method", "X-Method-Override", "X-Method"}
	for _, method := range methods {
		if testAccessControl(options, method) {
			msg := fmt.Sprintf("[Method] Access control Bypassed for target %s using method %s\n", options.Target, method)
			color.Red(msg)
			if options.Broker {
				common.PublishMessage(msg)
			}
		}
	}
	for _, header := range headers {
		if checkMethodOverwrite(options, header) {
			msg := fmt.Sprintf("[Method] Method Overwrite Bypassed for target %s using header %s\n", options.Target, header)
			color.Red(msg)
			if options.Broker {
				common.PublishMessage(msg)
			}
		}
	}

}

func multiMethodCheck(options *common.Opts) {
	multiplex.Conscan(singleMethodCheck, options, 10)
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
		return true
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
	if checker.Check405(control_resp) && !checker.Check405(treatment_resp) {
		return true
	}

	return false

}
