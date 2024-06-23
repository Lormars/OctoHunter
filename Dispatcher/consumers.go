package dispatcher

import (
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/crawler"
	"github.com/lormars/octohunter/pkg/modules"
	"github.com/lormars/octohunter/pkg/modules/request"
	"github.com/lormars/octohunter/pkg/modules/salesforce"
	"github.com/lormars/octohunter/pkg/modules/takeover"
)

func Init(opts *common.Opts) {
	for i := 0; i < opts.Concurrency/100; i++ {
		go cnameConsumer(opts)
		go redirectConsumer(opts)
		go methodConsumer(opts)
		go hopperConsumer(opts)
		go crawlerConsumer(opts)
		go dividerConsumer(opts)
		go salesforceConsumer(opts)
		go splittingConsumer(opts)
	}
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

func salesforceConsumer(opts *common.Opts) {
	common.SalesforceP.ConsumeMessage(salesforce.SalesforceScan, opts)
}

func splittingConsumer(opts *common.Opts) {
	common.SplittingP.ConsumeMessage(request.RequestSplitting, opts)
}
