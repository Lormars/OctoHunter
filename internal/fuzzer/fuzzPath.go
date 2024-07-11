package fuzzer

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

func init() {
	file, err := os.Open("list/onelistforallmicro.txt")
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

func FuzzPath(urlStr string) {
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
			req, err := http.NewRequest("GET", fuzzedURL, nil)
			if err != nil {
				return
			}
			resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
			if err != nil {
				return
			}
			if resp.StatusCode != 404 && resp.StatusCode != 403 {
				mu.Lock()
				resultMap[resp.Body] = resp
				mu.Unlock()
			}

		}(path)
	}
	wg.Wait()

	//this is necessary to filter out duplicate false positives
	for _, resp := range resultMap {
		common.DividerP.PublishMessage(resp)
		// logger.Warnf("found new endpoint: %s", fuzzPath)
		msg := fmt.Sprintf("[Fuzz Path(S)] Found new endpoint: %s with SC %d", resp.Url, resp.StatusCode)
		common.OutputP.PublishMessage(msg)
		notify.SendMessage(msg)
	}

}
