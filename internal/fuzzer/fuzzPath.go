package fuzzer

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

func init() {
	file, err := os.Open("asset/onelistforallmicro.txt")
	if err != nil {
		panic("Error opening file")
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		path := scanner.Text()
		if strings.TrimSpace(path) != "" {
			PathList = append(PathList, path)
		}
	}
	logger.Infof("Fuzz Wordlist loaded")
}

func FuzzPath(result *common.ServerResult) {
	urlStr := result.Url
	if !cacher.CheckCache(urlStr, "fuzzPath") {
		return
	}

	if strings.HasPrefix(urlStr, "http://") && checker.CheckHttpRedirectToHttps(urlStr) {
		return
	}

	logger.Debugf("FuzzPath: %s", urlStr)
	var wg sync.WaitGroup
	var mu sync.Mutex
	resultMap := make(map[string]*common.ServerResult)
	semaphore := make(chan struct{}, 10)
	for _, path := range PathList {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(path string) {
			defer wg.Done()
			defer func() {
				<-semaphore
			}()
			fuzzedURL := fmt.Sprintf("%s/%s", strings.TrimRight(urlStr, "/"), path)
			req, err := clients.NewRequest("GET", fuzzedURL, nil, clients.Fuzzpath)
			if err != nil {
				return
			}
			resp, err := checker.CheckServerCustom(req, clients.Clients.GetRandomClient("h0", false, true))
			if err != nil {
				return
			}
			if resp.StatusCode != 404 && resp.StatusCode != 403 && resp.StatusCode != 429 && resp.StatusCode != 500 {
				mu.Lock()
				hashed := common.Hash(resp.Body)
				if _, exists := resultMap[hashed]; !exists {
					resultMap[hashed] = resp
					common.AddToCrawlMap(resp.Url, "fuzz", resp.StatusCode)
					common.DividerP.PublishMessage(resp)
					// logger.Warnf("found new endpoint: %s", fuzzPath)
					if result.StatusCode == 404 { // if the original endpoint is 404 or 403, then it's a new endpoint that is worthy of notification
						msg := fmt.Sprintf("[Fuzz Path] Found new endpoint: %s with SC %d", resp.Url, resp.StatusCode)
						if common.SendOutput {
							common.OutputP.PublishMessage(msg)
						}
						notify.SendMessage(msg)
					}
				}
				mu.Unlock()
			}

		}(path)
	}
	wg.Wait()

}
