package checker

import (
	"fmt"
	"io"
	"net/http"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
)

func CheckHTTPAndHTTPSServers(domain string) (*common.ServerResult, *common.ServerResult, error, error) {
	httpURL := fmt.Sprintf("http://%s", domain)
	httpsURL := fmt.Sprintf("https://%s", domain)

	httpResult, errhttp := checkServer(httpURL)
	httpsResult, errhttps := checkServer(httpsURL)

	return httpResult, httpsResult, errhttp, errhttps
}

func checkServer(url string) (*common.ServerResult, error) {
	resp, err := common.NoRedirectClient.Get(url)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", url, err)
		return &common.ServerResult{
			Url:        url,
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
		Online:     resp.StatusCode >= 100 && resp.StatusCode < 600,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil

}
func CheckServerCustom(req *http.Request, client *http.Client) (*common.ServerResult, error) {
	resp, err := client.Do(req)
	url := req.URL.String()

	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", url, err)
		return &common.ServerResult{
			Url:        url,
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
		Online:     resp.StatusCode >= 100 && resp.StatusCode < 600,
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil

}
