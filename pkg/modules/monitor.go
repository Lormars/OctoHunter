package modules

import (
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/pkg/modules/takeover"
)

func Monitor(opts *common.Opts) {
	if opts.Cname {
		go func() {
			for {
				//takeover.MonitorPreprocess()
				takeover.CNAMETakeover(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	if opts.Dork {
		go func() {
			for {
				GoogleDork(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	if opts.Hopper {
		go func() {
			for {
				CheckHop(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	if opts.Redirect {
		go func() {
			for {
				CheckRedirect(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

	if opts.Method {
		go func() {
			for {
				CheckMethod(opts)
				time.Sleep(15 * time.Minute)
			}
		}()
	}

}
