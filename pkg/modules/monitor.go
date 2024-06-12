package modules

import (
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/pkg/modules/takeover"
)

func Monitor(opts *common.Opts) {
	if opts.Cname {
		for {
			//takeover.MonitorPreprocess()
			takeover.CNAMETakeover(opts)
			time.Sleep(15 * time.Minute)
		}

	}

}
