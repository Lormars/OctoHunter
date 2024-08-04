package checker

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/common/score"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/logger"
)

// for input scan, just use a normal client is fine
var inClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
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

	if strings.HasPrefix(url, "https://") {
		if resp.StatusCode == 403 {
			hostname := req.URL.Hostname()
			if cacher.CheckCache(hostname, "browser") {
				// statusCode := CheckWithRealBrowser(url)
				result, err := common.RequestWithBrowser(req, inClient)
				if err != nil {
					return nil, err
				}
				// logger.Warnf("rc %d", result.StatusCode)
				if result.StatusCode != 403 {
					logger.Warnln("endpoint has browser check")
					common.NeedBrowser[hostname] = true
					return &common.ServerResult{
						Url: url,
					}, fmt.Errorf("browser check")
				}
			}
			// msg := fmt.Sprintf("Endpoint %s DOES NOT HAVE browser check", url)
			// color.Red(msg)
		}
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
		if currentHostName == lowscore {
			logger.Warnf("Low score domain filtered: %s\n", currentHostName)
			return nil, fmt.Errorf("low score domain")
		}
	}
	score.ScoreMu.Unlock()

	respCh := common.AddToQueue(req.URL.Hostname(), []*http.Request{req}, client)
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

// func CheckWithRealBrowser(urlStr string) int {
// 	requestURL := "http://localhost:9999/status"
// 	requestData := map[string]string{"url": urlStr}

// 	jsonData, err := json.Marshal(requestData)
// 	if err != nil {
// 		logger.Warnf("Error marshalling request data: %v", err)
// 		return 403
// 	}

// 	ctx := context.Background()

// 	if err := limiter.Wait(ctx); err != nil {
// 		logger.Warnf("Error waiting for rate limiter: %v", err)
// 		return 403
// 	}

// 	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonData))
// 	if err != nil {
// 		logger.Warnf("Error creating request: %v", err)
// 		return 403
// 	}
// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err := puppClient.Do(req)
// 	if err != nil {
// 		logger.Warnf("Error getting response from %s: %v\n", requestURL, err)
// 		return 403
// 	}
// 	defer resp.Body.Close()

// 	// Read and print the response body
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		logger.Warnf("Error reading response body: %v", err)
// 		return 403
// 	}

// 	// Unmarshal the response
// 	var result map[string]interface{}
// 	if err := json.Unmarshal(body, &result); err != nil {
// 		logger.Warnf("Error unmarshalling response: %v", err)
// 		return 403
// 	}

// 	// Print the statusCode
// 	if statusCode, ok := result["statusCode"].(float64); ok {
// 		return int(statusCode)
// 	}
// 	return 403

// }
