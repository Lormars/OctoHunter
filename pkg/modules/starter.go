package modules

import (
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/pkg/modules/takeover"
	"github.com/lormars/octohunter/tools/controller"
)

func Startup(moduleManager *controller.ModuleManager, options *common.Opts) {

	if options.Module.Contains("monitor") {
		if options.Module.Contains("broker") {
			common.Init()
			defer common.Close()
		}
		Monitor(options)

	} else {
		if options.Module.Contains("hopper") {
			moduleManager.StartModule("hopper", CheckHop, options)
		}

		if options.Module.Contains("dork") {
			moduleManager.StartModule("dork", GoogleDork, options)
		}

		if options.Module.Contains("method") {
			moduleManager.StartModule("method", CheckMethod, options)
		}

		if options.Module.Contains("redirect") {
			moduleManager.StartModule("redirect", CheckRedirect, options)
		}

		if options.Module.Contains("cname") {
			moduleManager.StartModule("cname", takeover.CNAMETakeover, options)
		}
	}
}
