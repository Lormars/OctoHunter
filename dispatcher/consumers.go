package dispatcher

import (
	"sync"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/crawler"
	"github.com/lormars/octohunter/internal/fuzzer"
	"github.com/lormars/octohunter/internal/logger"
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
		"graphql":       graphqlConsumer,
		"mime":          mimeConsumer,
		"fuzzpath":      fuzzPathConsumer,
	}

	var maxConcurrent = map[string]int{
		"cname":         50,
		"redirect":      70,
		"method":        50,
		"hopper":        20,
		"divider":       10,
		"crawl":         50,
		"salesforce":    10,
		"splitting":     100,
		"cl0":           70,
		"quirks":        70,
		"rc":            70,
		"cors":          20,
		"pathconfuse":   20,
		"fuzz4034":      20,
		"pathtraversal": 70,
		"fuzzapi":       50,
		"fuzzunkeyed":   90,
		"xss":           70,
		"ssti":          60,
		"graphql":       50,
		"mime":          50,
		"fuzzpath":      50,
	}

	go func() {
		mu := sync.Mutex{}
		var numMap = make(map[string]numChan)
		var borrowMap = make(map[string]int)
		for _, function := range nameFuncMap {
			for i := 0; i < opts.Concurrency; i++ {
				go function(opts)
			}
		}
		for {
			common.GlobalMu.Lock()
			for name, waitingNum := range common.WaitingQueue {
				if name == "wayback" {
					continue
				}
				mu.Lock()
				startConsumer := waitingNum >= 2
				stopConsumer := waitingNum <= -3

				if stopConsumer {
					if numChan, ok := numMap[name]; ok {
						closeChan := numChan.chans[0]
						msg := "closing " + name + " due to lack of work"
						safeClose(closeChan, msg)
						if borrowNum, ok := borrowMap[name]; ok && borrowNum > 0 {
							borrowMap[name]--
						}
					}
				}

				if startConsumer {
					if numMap[name].num >= maxConcurrent[name] { //check if we are at max concurrency
						if len(common.ConsumerSemaphore) < cap(common.ConsumerSemaphore) { //check if we have space in the semaphore
							borrowMap[name]++ //borrow a consumer
						}
					} else if len(common.ConsumerSemaphore) == cap(common.ConsumerSemaphore) { //if we are not at max concurrency but no space in the semaphore
						for borrowFun, borrowNum := range borrowMap { //check if there are borrowed consumers
							if borrowNum > 0 { //if there are borrowed consumers, make them return
								borrowMap[borrowFun]--
								if numChan, ok := numMap[borrowFun]; ok {
									closeChan := numChan.chans[0]
									msg := "closing " + borrowFun + " due to borrowing"
									safeClose(closeChan, msg)
								} else {
									logger.Errorf("Shouldn't happen. Borrowed consumer not found: %s", borrowFun)
								}
								break
							}
						}
					}
				}
				mu.Unlock()

				if startConsumer {
					go func(name string) {
						common.ConsumerSemaphore <- struct{}{}
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
						<-common.ConsumerSemaphore
					}(name)

				}

				time.Sleep(1 * time.Second)
			}

			common.GlobalMu.Unlock()
			if len(common.ConsumerSemaphore) > 1100 {
				logger.Warnf("sepamore running: %d", len(common.ConsumerSemaphore))
			}
			time.Sleep(1 * time.Second)
		}
	}()

}

func recovery(msg string) {
	if r := recover(); r != nil {
		logger.Errorf("Recovered in consumer: %v (%s)", r, msg)
	}
}

func safeClose(ch chan struct{}, msg string) {
	defer recovery(msg)
	close(ch)
}

func fuzzPathConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.FuzzPathP.ConsumeMessage(fuzzer.FuzzPath, opts)
	return closeChan
}

func mimeConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.MimeP.ConsumeMessage(mime.CheckMime, opts)
	return closeChan
}

func graphqlConsumer(opts *common.Opts) chan struct{} {
	closeChan := common.GraphqlP.ConsumeMessage(api.CheckGraphql, opts)
	return closeChan
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
