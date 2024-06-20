package checker

import (
	"fmt"
	"io"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
)

func CheckHTTPAndHTTPSServers(domain string) (*common.ServerResult, *common.ServerResult) {
	httpURL := fmt.Sprintf("http://%s", domain)
	httpsURL := fmt.Sprintf("https://%s", domain)

	resultChan := make(chan common.ServerResult, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go checkServer(httpURL, resultChan, &wg)
	go checkServer(httpsURL, resultChan, &wg)

	wg.Wait()
	close(resultChan)

	var httpResult, httpsResult common.ServerResult
	for result := range resultChan {
		if result.Online {
			if result.Url == httpURL {
				httpResult = result
			} else {
				httpsResult = result
			}
		}
	}
	return &httpResult, &httpsResult
}

func checkServer(url string, resultChan chan<- common.ServerResult, wg *sync.WaitGroup) {
	defer wg.Done()
	resp, err := common.NoRedirectClient.Get(url)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", url, err)
		resultChan <- common.ServerResult{
			Url:        url,
			Online:     false,
			StatusCode: 0,
			Headers:    nil,
			Body:       "",
		}
		return
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	body := ""
	if err == nil {
		body = string(bodyBytes)
	}

	resultChan <- common.ServerResult{
		Url:        url,
		Online:     resp.StatusCode >= 100 && resp.StatusCode < 600,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}

}
