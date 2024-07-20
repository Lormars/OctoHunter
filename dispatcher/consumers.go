package dispatcher

import (
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/crawler"
	"github.com/lormars/octohunter/internal/fuzzer"
	"github.com/lormars/octohunter/internal/wayback"
	"github.com/lormars/octohunter/pkg/modules"
	pathconfusion "github.com/lormars/octohunter/pkg/modules/pathConfusion"
	"github.com/lormars/octohunter/pkg/modules/quirks"
	racecondition "github.com/lormars/octohunter/pkg/modules/raceCondition"
	"github.com/lormars/octohunter/pkg/modules/request"
	"github.com/lormars/octohunter/pkg/modules/request/smuggle"
	"github.com/lormars/octohunter/pkg/modules/salesforce"
	"github.com/lormars/octohunter/pkg/modules/takeover"
)

func Init(opts *common.Opts) {

	go waybackConsumer(opts) //only one due to rate limit

	for i := 0; i < opts.Concurrency/10; i++ {
		go redirectConsumer(opts)
		go methodConsumer(opts)
		go hopperConsumer(opts)
		go salesforceConsumer(opts)
		go cl0Consumer(opts)
		go dividerConsumer(opts)
		go raceConditionConsumer(opts)
		go corsConsumer(opts)
		go xssConsumer(opts)
	}
	for i := 0; i < opts.Concurrency; i++ {
		go cnameConsumer(opts)
		go crawlerConsumer(opts)
		go quirksConsumer(opts)
		go pathConfuse(opts)
		go fuzz404Consumer(opts)
		go pathTraversalConsumer(opts)
		go fuzzAPIConsumer(opts)
		go fuzzUnkeyedConsumer(opts)
		go sstiConsumer(opts)
		go splittingConsumer(opts)
	}
}

func waybackConsumer(opts *common.Opts) {
	common.WaybackP.ConsumeMessage(wayback.GetWaybackURLs, opts)
}

func sstiConsumer(opts *common.Opts) {
	common.SstiP.ConsumeMessage(modules.CheckSSTI, opts)
}

func xssConsumer(opts *common.Opts) {
	common.XssP.ConsumeMessage(modules.Xss, opts)
}

func fuzzUnkeyedConsumer(opts *common.Opts) {
	common.FuzzUnkeyedP.ConsumeMessage(fuzzer.FuzzUnkeyed, opts)
}

func fuzzAPIConsumer(opts *common.Opts) {
	common.FuzzAPIP.ConsumeMessage(fuzzer.FuzzAPI, opts)

}

func pathTraversalConsumer(opts *common.Opts) {
	common.PathTraversalP.ConsumeMessage(modules.CheckPathTraversal, opts)
}

func fuzz404Consumer(opts *common.Opts) {
	common.Fuzz4034P.ConsumeMessage(fuzzer.Fuzz4034, opts)
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

func cl0Consumer(opts *common.Opts) {
	common.Cl0P.ConsumeMessage(smuggle.CheckCl0, opts)
}

func quirksConsumer(opts *common.Opts) {
	common.QuirksP.ConsumeMessage(quirks.CheckQuirks, opts)
}

func raceConditionConsumer(opts *common.Opts) {
	common.RCP.ConsumeMessage(racecondition.RaceCondition, opts)
}

func corsConsumer(opts *common.Opts) {
	common.CorsP.ConsumeMessage(modules.CheckCors, opts)
}

func pathConfuse(opts *common.Opts) {
	common.PathConfuseP.ConsumeMessage(pathconfusion.CheckPathConfusion, opts)
}