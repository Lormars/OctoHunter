package dispatcher

import (
	"sync"
	"time"

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

type numChan struct {
	num   int
	chans []chan struct{}
}

func Init(opts *common.Opts) {

	go waybackConsumer(opts) //only one due to rate limit

	var nameFuncMap = map[string]func(*common.Opts) chan struct{}{
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
	}

	semaphore := make(chan struct{}, 1000)

	go func() {
		mu := sync.Mutex{}
		var numMap = make(map[string]numChan)
		for _, function := range nameFuncMap {
			go function(opts)
		}
		for {
			common.GlobalMu.Lock()
			for name, waitingNum := range common.WaitingQueue {
				if name == "wayback" {
					continue
				}
				if waitingNum >= 2 {
					semaphore <- struct{}{}
					go func(name string) {
						closeChan := nameFuncMap[name](opts)
						mu.Lock()
						if _, ok := numMap[name]; !ok {
							numMap[name] = numChan{num: 1, chans: []chan struct{}{closeChan}}
						} else {
							numchan := numMap[name]
							numchan.num++
							numchan.chans = append(numchan.chans, closeChan)
							numMap[name] = numchan
						}
						// logger.Infof("Starting %s, have %d running", name, numMap[name].num+1)
						mu.Unlock()
						<-closeChan
						mu.Lock()
						newNum := numMap[name].num - 1
						if newNum == 0 {
							delete(numMap, name)
						} else {
							numMap[name] = numChan{num: newNum, chans: numMap[name].chans[1:]}
						}
						// logger.Infof("Stopping %s, have %d running", name, newNum+1)
						mu.Unlock()
						<-semaphore
					}(name)

				} else if waitingNum <= -5 {
					mu.Lock()
					if numChan, ok := numMap[name]; ok {
						closeChan := numChan.chans[0]
						close(closeChan)
					}
					mu.Unlock()
				}
			}

			common.GlobalMu.Unlock()
			time.Sleep(1 * time.Second)
		}
	}()

}

func waybackConsumer(opts *common.Opts) {
	common.WaybackP.ConsumeMessage(wayback.GetWaybackURLs, opts)
}

func sstiConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.SstiP.ConsumeMessage(modules.CheckSSTI, opts)
	return closeChan
}

func xssConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.XssP.ConsumeMessage(modules.Xss, opts)
	return closeChan
}

func fuzzUnkeyedConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.FuzzUnkeyedP.ConsumeMessage(fuzzer.FuzzUnkeyed, opts)
	return closeChan
}

func fuzzAPIConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.FuzzAPIP.ConsumeMessage(fuzzer.FuzzAPI, opts)
	return closeChan

}

func pathTraversalConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.PathTraversalP.ConsumeMessage(modules.CheckPathTraversal, opts)
	return closeChan
}

func fuzz404Consumer(opts *common.Opts) chan struct{} {
	closeChan := common.Fuzz4034P.ConsumeMessage(fuzzer.Fuzz4034, opts)
	return closeChan
}

func cnameConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.CnameP.ConsumeMessage(takeover.Takeover, opts)
	return closeChan
}

func redirectConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.RedirectP.ConsumeMessage(modules.SingleRedirectCheck, opts)
	return closeChan
}

func methodConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.MethodP.ConsumeMessage(modules.SingleMethodCheck, opts)
	return closeChan
}

func hopperConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.HopP.ConsumeMessage(modules.SingleHopCheck, opts)
	return closeChan
}

func dividerConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.DividerP.ConsumeMessage(Divider, opts)
	return closeChan
}

func crawlerConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.CrawlP.ConsumeMessage(crawler.Crawl, opts)
	return closeChan
}

func salesforceConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.SalesforceP.ConsumeMessage(salesforce.SalesforceScan, opts)
	return closeChan
}

func splittingConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.SplittingP.ConsumeMessage(request.RequestSplitting, opts)
	return closeChan
}

func cl0Consumer(opts *common.Opts) chan struct{} {
	closeChan := common.Cl0P.ConsumeMessage(smuggle.CheckCl0, opts)
	return closeChan
}

func quirksConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.QuirksP.ConsumeMessage(quirks.CheckQuirks, opts)
	return closeChan
}

func raceConditionConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.RCP.ConsumeMessage(racecondition.RaceCondition, opts)
	return closeChan
}

func corsConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.CorsP.ConsumeMessage(modules.CheckCors, opts)
	return closeChan
}

func pathConfuse(opts *common.Opts) chan struct{} {
	closeChan := common.PathConfuseP.ConsumeMessage(pathconfusion.CheckPathConfusion, opts)
	return closeChan
}
