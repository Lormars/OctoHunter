package matcher

import (
	"net/url"
	"strings"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
)

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
