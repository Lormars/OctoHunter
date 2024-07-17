package racecondition

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"math/rand"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/common/clients/proxyP"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

// target must be a valid URL
func RaceCondition(urlStr string) {
	//check cache
	if !cacher.CheckCache(urlStr, "race") {
		return
	}

	common.AddToCrawlMap(urlStr, "race", 200) //TODO: can be accurate

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	var mu sync.Mutex
	var treatResponses []*common.ServerResult
	// var controlResponses []*common.ServerResult

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Warnf("Failed to parse URL: %v", err)
		return
	}

	pattern := `wxoyvz\d{1,2}`
	re, err := regexp.Compile(pattern)
	if err != nil {
		logger.Warnf("Failed to compile regex: %v", err)
		return
	}
	proxyP.Proxies.Mu.Lock()
	proxy := proxyP.Proxies.Proxies[rand.Intn(len(proxyP.Proxies.Proxies))]
	proxyP.Proxies.Mu.Unlock()

	ctx := context.WithValue(context.Background(), "race", true)
	ctx = context.WithValue(ctx, "proxy", proxy)
	// Function to send a request
	sendRequest := func(id int) {
		defer wg.Done()
		buster := "wxoyvz" + strconv.Itoa(id)
		cachebuster := "boqpz=" + buster
		payloadURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path + "?" + cachebuster

		req, err := http.NewRequestWithContext(ctx, "GET", payloadURL, nil)
		if err != nil {
			logger.Warnf("Goroutine %d: Failed to create request: %v", id, err)
			return
		}

		//req.Header.Set("User-Agent", "")

		// Send the request

		resp, err := clients.NoRedirectRCClient.Do(req)
		if err != nil {
			logger.Debugf("Goroutine %d: Failed to send request: %v", id, err)
			return
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Debugf("Goroutine %d: Failed to read response body: %v", id, err)
		}
		response := &common.ServerResult{
			StatusCode: resp.StatusCode,
			Body:       string(body),
			Online:     resp.StatusCode >= 100 && resp.StatusCode < 600,
			Url:        payloadURL,
			Headers:    resp.Header,
		}

		match := re.FindString(response.Body)

		if match != "" && match != buster {
			common.CrawlP.PublishMessage(response)
			msg := fmt.Sprintf("[RC Confirmed] Race Condition on endpoint %s (match is %s)", response.Url, match)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
		}

		mu.Lock()
		treatResponses = append(treatResponses, response)
		mu.Unlock()

	}

	//control group
	// for i := 1; i <= 20; i++ {
	// 	req, err := http.NewRequest("GET", urlStr, nil)
	// 	if err != nil {
	// 		logger.Warnf("Failed to create request: %v", i, err)
	// 		return
	// 	}
	// 	resp, err := checker.CheckServerCustom(req, clients.NoRedirecth2Client)
	// 	if err != nil {
	// 		logger.Warnf("Failed to send request: %v", i, err)
	// 		return
	// 	}
	// 	controlResponses = append(controlResponses, resp)

	// }

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logger.Warnf("Failed to create request: %v", err)
		return
	}
	resp, err := checker.CheckServerCustom(req, clients.NoRedirecth2Client)
	if err != nil { //first check with a normal http2 request to see if the server accepts HTTP2 protocol
		logger.Debugf("Failed to send request: %v", err)
		return
	}
	controlStatus := resp.StatusCode

	time.Sleep(1 * time.Second)

	// Launch 20 goroutines to send requests concurrently
	for i := 1; i <= 20; i++ {
		wg.Add(1)
		go sendRequest(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	for _, resp := range treatResponses {
		if resp.StatusCode != controlStatus && resp.StatusCode > 400 && resp.StatusCode != 429 && resp.StatusCode != 502 && resp.StatusCode != 503 && resp.StatusCode != 403 {
			msg := fmt.Sprintf("[RC Suspect] Race Condition on %s with status %d", resp.Url, resp.StatusCode)
			common.CrawlP.PublishMessage(resp)
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
			break
		}
	}
}
