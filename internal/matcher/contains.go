package matcher

import (
	"net/url"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
)

// HeadercontainsQueryParamValue checks if the header contains the query parameter value
// If signature is empty, it checks if the header contains the query parameter value
// If signature is not empty, it checks if the header contains the signature
func HeadercontainsQueryParamValue(result *common.ServerResult, signature string) (bool, []string) {
	contains := false
	var found []string
	urlStr := result.Url
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		logger.Debugf("Error parsing URL: %v\n", err)
		return contains, found
	}

	queryParams := parsedURL.Query()
outer:
	for param, values := range queryParams {
		for _, queryValue := range values {
			for _, headerValues := range result.Headers {
				for _, headerValue := range headerValues {
					if signature == "" {
						if strings.Contains(headerValue, queryValue) {
							found = append(found, param)
							contains = true
							continue outer
						}
					} else {
						if strings.Contains(headerValue, signature) {
							found = append(found, param)
							contains = true
							continue
						}
					}
				}
			}
		}
	}
	return contains, found
}

// HeaderValueContainsSignature checks if the header key contains the signature
func HeaderKeyContainsSignature(result *common.ServerResult, signature string) bool {
	for headerKey := range result.Headers {
		if strings.Contains(strings.ToLower(headerKey), strings.ToLower(signature)) {
			return true
		}
	}

	return false
}

// HeaderValueContainsSignature checks if any header value contains the signature
func HeaderValueContainsSignature(result *common.ServerResult, signature string) bool {
	lowerSignature := strings.ToLower(signature)
	for _, headerValues := range result.Headers {
		for _, headerValue := range headerValues {
			if strings.Contains(strings.ToLower(headerValue), lowerSignature) {
				return true
			}
		}
	}
	return false
}
