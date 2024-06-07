package modules

import (
	"fmt"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/comparer"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/requester/pkg/runner"
)

func CheckHop(options *common.Opts) {
	if options.Target != "none" {
		singleCheck(options)
	} else {
		multiCheck(options)
	}
}

func multiCheck(options *common.Opts) {
	multiplex.Conscan(singleCheck, options)
}

func singleCheck(options *common.Opts) {
	control_config := runner.NewConfig(options.Target)
	treatment_config := runner.NewConfig(options.Target)
	treatment_config.Header_input = "Connection: close, X-Forwarded-For"
	control_resp := runner.Run(control_config)
	treatment_resp := runner.Run(treatment_config)
	if !comparer.CompareResponse(control_resp, treatment_resp) {
		fmt.Println("The responses are different")
	}
}
