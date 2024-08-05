package dispatcher

import (
	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/crawler"
	"github.com/lormars/octohunter/internal/fuzzer"
	"github.com/lormars/octohunter/internal/wayback"
	"github.com/lormars/octohunter/pkg/modules"
	"github.com/lormars/octohunter/pkg/modules/api"
	pathconfusion "github.com/lormars/octohunter/pkg/modules/pathConfusion"
	"github.com/lormars/octohunter/pkg/modules/quirks"
	racecondition "github.com/lormars/octohunter/pkg/modules/raceCondition"
	"github.com/lormars/octohunter/pkg/modules/request"
	"github.com/lormars/octohunter/pkg/modules/request/smuggle"
	"github.com/lormars/octohunter/pkg/modules/research/mime"
	"github.com/lormars/octohunter/pkg/modules/salesforce"
	"github.com/lormars/octohunter/pkg/modules/takeover"
)

func Init() {

	var nameFuncMap = map[string]func(){
		"cname":         cnameConsumer,
		"redirect":      redirectConsumer,
		"method":        methodConsumer,
		"hopper":        hopperConsumer,
		"divider":       dividerConsumer,
		"crawl":         crawlerConsumer,
		"salesforce":    salesforceConsumer,
		"splitting":     splittingConsumer,
		"cl0":           cl0Consumer,
		"quirks":        quirksConsumer,
		"rc":            raceConditionConsumer,
		"cors":          corsConsumer,
		"pathconfuse":   pathConfuse,
		"fuzz4034":      fuzz404Consumer,
		"pathtraversal": pathTraversalConsumer,
		"fuzzapi":       fuzzAPIConsumer,
		"fuzzunkeyed":   fuzzUnkeyedConsumer,
		"xss":           xssConsumer,
		"ssti":          sstiConsumer,
		"graphql":       graphqlConsumer,
		"mime":          mimeConsumer,
		"fuzzpath":      fuzzPathConsumer,
		"wayback":       waybackConsumer,
	}

	for _, fn := range nameFuncMap {
		fn()
	}

}

func fuzzPathConsumer() {
	common.FuzzPathP.ConsumeMessage(fuzzer.FuzzPath)

}

func mimeConsumer() {
	common.MimeP.ConsumeMessage(mime.CheckMime)

}

func graphqlConsumer() {
	common.GraphqlP.ConsumeMessage(api.CheckGraphql)

}

func waybackConsumer() {
	common.WaybackP.ConsumeMessage(wayback.GetWaybackURLs)
}

func sstiConsumer() {
	common.SstiP.ConsumeMessage(modules.CheckSSTI)

}

func xssConsumer() {
	common.XssP.ConsumeMessage(modules.Xss)

}

func fuzzUnkeyedConsumer() {
	common.FuzzUnkeyedP.ConsumeMessage(fuzzer.FuzzUnkeyed)

}

func fuzzAPIConsumer() {
	common.FuzzAPIP.ConsumeMessage(fuzzer.FuzzAPI)

}

func pathTraversalConsumer() {
	common.PathTraversalP.ConsumeMessage(modules.CheckPathTraversal)

}

func fuzz404Consumer() {
	common.Fuzz4034P.ConsumeMessage(fuzzer.Fuzz4034)

}

func cnameConsumer() {
	common.CnameP.ConsumeMessage(takeover.Takeover)

}

func redirectConsumer() {
	common.RedirectP.ConsumeMessage(modules.SingleRedirectCheck)

}

func methodConsumer() {
	common.MethodP.ConsumeMessage(modules.SingleMethodCheck)

}

func hopperConsumer() {
	common.HopP.ConsumeMessage(modules.SingleHopCheck)

}

func dividerConsumer() {
	common.DividerP.ConsumeMessage(Divider)

}

func crawlerConsumer() {
	common.CrawlP.ConsumeMessage(crawler.Crawl)

}

func salesforceConsumer() {
	common.SalesforceP.ConsumeMessage(salesforce.SalesforceScan)

}

func splittingConsumer() {
	common.SplittingP.ConsumeMessage(request.RequestSplitting)

}

func cl0Consumer() {
	common.Cl0P.ConsumeMessage(smuggle.CheckCl0)

}

func quirksConsumer() {
	common.QuirksP.ConsumeMessage(quirks.CheckQuirks)

}

func raceConditionConsumer() {
	common.RCP.ConsumeMessage(racecondition.RaceCondition)

}

func corsConsumer() {
	common.CorsP.ConsumeMessage(modules.CheckCors)

}

func pathConfuse() {
	common.PathConfuseP.ConsumeMessage(pathconfusion.CheckPathConfusion)

}
