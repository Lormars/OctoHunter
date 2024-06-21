package dispatcher

import (
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/crawler"
	"github.com/lormars/octohunter/pkg/modules"
	"github.com/lormars/octohunter/pkg/modules/takeover"
)

func Init(opts *common.Opts) {
	go cnameConsumer(opts)
	go redirectConsumer(opts)
	go methodConsumer(opts)
	go hopperConsumer(opts)
	go dividerConsumer(opts)
	go crawlerConsumer(opts)
}

func cnameConsumer(opts *common.Opts) {
	common.CnameP.ConsumeMessage(takeover.Takeover, opts)
}

func redirectConsumer(opts *common.Opts) {
	common.RedirectP.ConsumeMessage(modules.SingleRedirectCheck, opts)
}

func methodConsumer(opts *common.Opts) {
	common.MethodP.ConsumeMessage(modules.SingleMethodCheck, opts)
}

func hopperConsumer(opts *common.Opts) {
	common.HopP.ConsumeMessage(modules.SingleHopCheck, opts)
}

func dividerConsumer(opts *common.Opts) {
	common.DividerP.ConsumeMessage(Divider, opts)
}

func crawlerConsumer(opts *common.Opts) {
	common.CrawlP.ConsumeMessage(crawler.Crawl, opts)
}
