package checker

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/common/queue"
	"github.com/lormars/octohunter/common/score"
	"github.com/lormars/octohunter/internal/logger"
)

// for input scan, just use a normal client is fine
var inClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Timeout: 60 * time.Second,
}

// for puppeteer
// var puppClient = &http.Client{}
// var limiter = rate.NewLimiter(2, 4)

// Usage: check if the server is online, using NoRedirectClient
func CheckHTTPAndHTTPSServers(domain string) (*common.ServerResult, *common.ServerResult, error, error) {
	httpURL := fmt.Sprintf("http://%s", domain)
	httpsURL := fmt.Sprintf("https://%s", domain)

	httpResult, errhttp := checkServer(httpURL)
	httpsResult, errhttps := checkServer(httpsURL)

	return httpResult, httpsResult, errhttp, errhttps
}

func checkServer(url string) (*common.ServerResult, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v", err)
		return nil, err
	}

	resp, err := inClient.Do(req)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debugf("Error reading response body: %v", err)
		bodyBytes = []byte{}
	}

	return &common.ServerResult{
		Url:        url,
		Online:     resp.StatusCode >= 100 && resp.StatusCode < 600,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       string(bodyBytes),
	}, nil
}

// The ultra-important requester for (nearly) all request...
func CheckServerCustom(req *http.Request, client *clients.OctoClient) (*common.ServerResult, error) {

	// Check if the request is in the lowscoredomain
	currentHostName := req.URL.Hostname()
	score.ScoreMu.Lock()
	for _, lowscore := range score.LowScoreDomains {
		if strings.Contains(currentHostName, lowscore) {
			logger.Warnf("Low score domain filtered: %s\n", currentHostName)
			return nil, fmt.Errorf("low score domain")
		}
	}
	score.ScoreMu.Unlock()

	respCh := queue.AddToQueue(req.URL.Hostname(), []*http.Request{req}, client)
	resps := <-respCh
	resp := resps[0].Resp
	err := resps[0].Err
	url := req.URL.String()

	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", url, err)
		return &common.ServerResult{
			Url:        url,
			FinalUrl:   req.URL,
			Online:     false,
			StatusCode: 0,
			Headers:    nil,
			Body:       "",
		}, err
	}
	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	body := ""
	if err == nil {
		body = string(bodyBytes)
	}

	return &common.ServerResult{
		Url:        url,
		FinalUrl:   resp.Request.URL,
		Online:     resp.StatusCode >= 100 && resp.StatusCode < 600,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}
