package modules

import (
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/pkg/modules/takeover"
)

func Monitor(opts *common.Opts) {
	if opts.Cname {
		takeover.CNAMETakeover(opts)
		//fsnotify on cnames and gunames...
	}

}
