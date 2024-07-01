package racecondition

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"sync"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/common/clients/proxyP"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
	"golang.org/x/exp/rand"
)

// target must be a valid URL
func RaceCondition(urlStr string) {
	//check cache

	// WaitGroup to wait for all goroutines to finish
	var wg sync.WaitGroup
	var mu sync.Mutex
	var treatResponses []*common.ServerResult
	var controlResponses []*common.ServerResult

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Warnf("Failed to parse URL: %v", err)
		return
	}

	pattern := `wxoyvz\d`
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
		buster := "wxoyvz" + string(id)
		cachebuster := "boqpz=" + buster
		payloadURL := parsedURL.Scheme + "://" + parsedURL.Host + parsedURL.Path + "?" + cachebuster

		req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
		if err != nil {
			logger.Warnf("Goroutine %d: Failed to create request: %v", id, err)
			return
		}

		//req.Header.Set("User-Agent", "")

		// Send the request

		resp, err := clients.NoRedirectRCClient.Do(req)
		if err != nil {
			log.Printf("Goroutine %d: Failed to send request: %v", id, err)
			return
		}
		defer resp.Body.Close()
		fmt.Printf("Goroutine %d: %s\n", id, resp.Status)
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
			msg := fmt.Sprintf("[RC Confirmed] Race Condition on endpoint %s", response.Url)
			common.OutputP.PublishMessage(msg)
			notify.SendMessage(msg)
		}

		mu.Lock()
		treatResponses = append(treatResponses, response)
		mu.Unlock()

	}

	//control group
	for i := 1; i <= 20; i++ {
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			logger.Warnf("Failed to create request: %v", i, err)
			return
		}
		resp, err := checker.CheckServerCustom(req, clients.NoRedirecth2Client)
		if err != nil {
			logger.Warnf("Failed to send request: %v", i, err)
			return
		}
		controlResponses = append(controlResponses, resp)

	}

	// Launch 20 goroutines to send requests concurrently
	for i := 1; i <= 20; i++ {
		wg.Add(1)
		go sendRequest(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	for _, resp := range treatResponses {
		if resp.StatusCode != 429 && resp.StatusCode != 502 && resp.StatusCode != 503 && resp.StatusCode != 403 {
			msg := fmt.Sprintf("[RC Suspect] Race Condition on %s", resp.Url)
			common.OutputP.PublishMessage(msg)
			notify.SendMessage(msg)

		}
	}
}
