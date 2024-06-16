package crawler

import (
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/parser"
)

// Crawler that does not follow redirect
func Crawl(target string, concurrency int) {
	// Crawl the web
	status, body := checkStatus(target)
	switch status {
	case common.OK:
		var wg sync.WaitGroup
		urlCh := make(chan string, concurrency)
		urls := parser.ExtractUrls(target, body)
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for url := range urlCh {
					Crawl(url, 2)
				}
			}()
		}
		for _, url := range urls {
			urlCh <- url
		}

		wg.Wait()
		//TODO: add other cases
	}

}

func checkStatus(target string) (int, string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(target)
	if err != nil {
		fmt.Printf("Error crawling %s: %v\n", target, err)
		return common.XERROR, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error reading response body: %v\n", err)
			return common.XERROR, ""
		}
		return common.OK, string(body)
	} else if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		return common.REDIRECT, ""
	} else if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		return common.CLIENTERR, ""
	} else if resp.StatusCode >= 500 {
		return common.SERVERERR, ""
	}

	return common.XERROR

}
