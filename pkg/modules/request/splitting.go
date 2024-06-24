package request

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/generator"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/matcher"
)

func RequestSplitting(result *common.ServerResult) {
	if !cacher.CheckCache(result.Url, "split") {
		return
	}
	paramSplitTest(result)
	pathSplitTest(result)

}

// This function tests for HTTP Request Splitting by injecting a CRLF sequence in the parameters
func paramSplitTest(result *common.ServerResult) {
	var params []string
	var ok bool

	//It first checks if the header contains the query parameter value
	if ok, params = matcher.HeadercontainsQueryParamValue(result, ""); !ok {
		return
	}

	parsedURL, err := url.Parse(result.Url)
	if err != nil {
		logger.Debugf("Error parsing URL: %v\n", err)
	}

	//to filter out the parameters that are not controllable
	var controllable []string
	for _, param := range params {
		queryParams := parsedURL.Query()
		signature, err := generator.GenerateSignature()
		if err != nil {
			logger.Debugf("Error generating signature: %v\n", err)
			return
		}
		queryParams.Set(param, signature)
		parsedURL.RawQuery = queryParams.Encode()
		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			logger.Debugf("Error creating request: %v", err)
			continue
		}
		resp, err := checker.CheckServerCustom(req, clients.NoRedirecth1Client)
		if err != nil {
			logger.Debugf("Error getting response from %s: %v\n", parsedURL.String(), err)
			continue
		}
		if ok, _ := matcher.HeadercontainsQueryParamValue(resp, signature); ok {
			controllable = append(controllable, param)
		}

	}

	if len(controllable) == 0 {
		return
	}

	for _, param := range controllable {
		queryParams := parsedURL.Query()
		if err != nil {
			logger.Debugf("Error generating signature: %v\n", err)
			return
		}
		payload := "whatATest%0d%0aX-Injected:%20whatANiceDay%0d%0a"
		queryParams.Set(param, payload)

		//had to make sure all other parameters are included and properly encoded in the URL
		rawQuery := ""
		for key, values := range queryParams {
			for _, value := range values {
				if rawQuery != "" {
					rawQuery += "&"
				}
				if key != param {
					rawQuery += key + "=" + url.QueryEscape(value)
				} else {
					rawQuery += key + "=" + value
				}
			}
		}
		parsedURL.RawQuery = rawQuery

		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			logger.Debugf("Error creating request: %v", err)
			continue
		}
		resp, err := checker.CheckServerCustom(req, clients.NoRedirecth1Client)
		if err != nil {
			logger.Debugf("Error getting response from %s: %v\n", parsedURL.String(), err)
			continue
		}
		logger.Debugf("[Param Split] Testing for HTTP Request Splitting: %s on param %s\n", result.Url, param)
		if matcher.HeaderKeyContainsSignature(resp, "X-Injected") {
			msg := fmt.Sprintf("[Param Split] Vulnerable to HTTP Request Splitting: %s on param %s\n", result.Url, param)
			logger.Infof(msg)
			common.OutputP.PublishMessage(msg)
		}

	}
}

// This, right now, checks only for Location based path splitting that usually happen in http to https redirect
func pathSplitTest(result *common.ServerResult) {
	//return if there is no Location header
	if !matcher.HeaderKeyContainsSignature(result, "Location") {
		return
	}

	location := result.Headers.Get("Location")
	parsedUrl, err := url.Parse(result.Url)
	if err != nil {
		logger.Debugf("Error parsing URL: %v\n", err)
		return
	}
	//return if the URL is not http
	if parsedUrl.Scheme != "http" {
		logger.Debugf("Not a http URL: %s\n", result.Url)
		return
	}
	https_url := fmt.Sprintf("https://%s", parsedUrl.Host)
	//return if the Location header does not redirect to https
	if !strings.Contains(location, https_url) {
		logger.Debugf("Location header does not redirect to https: %s\n", location)
		return
	}

	path := "%0d%0aX-Injected:%20whatANiceDay%0d%0a"
	payloadUrl := fmt.Sprintf("http://%s/%s", parsedUrl.Host, path)
	req, err := http.NewRequest("GET", payloadUrl, nil)
	if err != nil {
		logger.Errorf("Error creating request: %v", err)
		return
	}
	resp, err := checker.CheckServerCustom(req, clients.NoRedirecth1Client)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", payloadUrl, err)
		return
	}
	logger.Debugf("[Path Split] Testing for HTTP Request Splitting: %s\n", result.Url)
	if matcher.HeaderKeyContainsSignature(resp, "X-Injected") {
		msg := fmt.Sprintf("[Path Split] Vulnerable to HTTP Request Splitting: %s\n", result.Url)
		logger.Infof(msg)
		common.OutputP.PublishMessage(msg)
	}

}
