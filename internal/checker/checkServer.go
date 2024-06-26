package checker

import (
	"fmt"
	"io"
	"net/http"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/logger"
)

// Usage: check if the server is online, using NoRedirectClient
func CheckHTTPAndHTTPSServers(domain string) (*common.ServerResult, *common.ServerResult, error, error) {
	httpURL := fmt.Sprintf("http://%s", domain)
	httpsURL := fmt.Sprintf("https://%s", domain)

	httpResult, errhttp := checkServer(httpURL)
	httpsResult, errhttps := checkServer(httpsURL)

	return httpResult, httpsResult, errhttp, errhttps
}

func checkServer(urlStr string) (*common.ServerResult, error) {

	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v", err)
		return &common.ServerResult{
			Url:        urlStr,
			FinalUrl:   nil,
			Online:     false,
			StatusCode: 0,
			Headers:    nil,
			Body:       "",
		}, err
	}
	resp, err := CheckServerCustom(req, clients.NoRedirectClient)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", urlStr, err)
		return &common.ServerResult{
			Url:        urlStr,
			FinalUrl:   nil,
			Online:     false,
			StatusCode: 0,
			Headers:    nil,
			Body:       "",
		}, err
	}

	return &common.ServerResult{
		Url:        urlStr,
		Online:     resp.StatusCode >= 100 && resp.StatusCode < 600,
		StatusCode: resp.StatusCode,
		Headers:    resp.Headers,
		Body:       resp.Body,
	}, nil

}

// The ultra-important requester for (nearly) all request...
func CheckServerCustom(req *http.Request, client *http.Client) (*common.ServerResult, error) {
	respCh := clients.AddToQueue(req.Host, []*http.Request{req}, client)
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
