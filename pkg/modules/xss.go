package modules

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
)

var prefix = "wpqbuzi"
var suffix = "pqpbqza"

// Xss checkes for possible xss vulnerabilities in the given url
func Xss(xssInput *common.XssInput) {

	for_cache := xssInput.Url + xssInput.Param
	if !cacher.CheckCache(for_cache, "xss") {
		return
	}

	common.AddToCrawlMap(xssInput.Url, "xss", 200) //TODO: can be accurate

	// Do the xss check here
	if xssInput.Location == "attribute" {
		checkXssInAttribute(xssInput.Url, xssInput.Param)
	} else if xssInput.Location == "tag" {
		checkXssInTag(xssInput.Url, xssInput.Param)
	} else if xssInput.Location == "both" {
		checkXssInAttribute(xssInput.Url, xssInput.Param)
		checkXssInTag(xssInput.Url, xssInput.Param)
	}

}

// Check for inappropiate encoding of double quotes and &quot; in attributes
func checkXssInAttribute(urlStr string, param string) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	queries := parsedURL.Query()

	toCheck := prefix + `"` + suffix

	payloads := []string{
		prefix + `"` + suffix,
		prefix + "&quot;" + suffix,
	}

	for _, payload := range payloads {
		queries.Set(param, payload)
		parsedURL.RawQuery = queries.Encode()
		// Do the request here
		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			logger.Warnf("Error creating request: %v", err)
			continue
		}
		resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
		if err != nil {
			logger.Debugf("Error getting response: %v", err)
			continue
		}
		if strings.Contains(resp.Body, toCheck) {
			msg := "[XSS] Possible XSS in attribute found: " + urlStr + " with parameter " + param
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
		}
	}
}

func checkXssInTag(urlStr string, param string) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return
	}
	queries := parsedURL.Query()

	toCheck := prefix + `<>` + suffix

	payloads := []string{
		prefix + `<>` + suffix,
		prefix + `&lt;&gt;` + suffix,
	}

	for _, payload := range payloads {
		queries.Set(param, payload)
		parsedURL.RawQuery = queries.Encode()
		// Do the request here
		req, err := http.NewRequest("GET", parsedURL.String(), nil)
		if err != nil {
			logger.Warnf("Error creating request: %v", err)
			continue
		}
		resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
		if err != nil {
			logger.Debugf("Error getting response: %v", err)
			continue
		}
		if strings.Contains(resp.Body, toCheck) {
			msg := "[XSS] Possible XSS in tag found: " + urlStr + " with parameter " + param
			if common.SendOutput {
				common.OutputP.PublishMessage(msg)
			}
			notify.SendMessage(msg)
		}
	}
}
