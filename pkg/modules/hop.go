package modules

import (
	"context"
	"fmt"
	"sync"

	"github.com/fatih/color"
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/comparer"
	"github.com/lormars/octohunter/internal/multiplex"
	"github.com/lormars/requester/pkg/runner"
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
	control_config, err1 := runner.NewConfig(options.Target)
	treatment_config, err2 := runner.NewConfig(options.Target)
	if err1 != nil || err2 != nil {
		return
	}
	control_config.Header_input = "Connection: close"
	treatment_config.Header_input = "Connection: close, X-Forwarded-For"
	control_resp, err1 := runner.Run(control_config)
	treatment_resp, err2 := runner.Run(treatment_config)
	if err1 != nil || err2 != nil {
		return
	}
	result, place := comparer.CompareResponse(control_resp, treatment_resp)
	if !result && place == "status" {
		if treatment_resp.Status < 400 {
			msg := fmt.Sprintf("[Hop] The responses are different for %s: %d vs %d\n", options.Target, control_resp.Status, treatment_resp.Status)
			color.Red(msg)
			if options.Module.Contains("broker") {
				common.PublishMessage(msg)
			}
		}
	}
}
