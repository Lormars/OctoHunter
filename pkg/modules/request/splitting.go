package request

import (
	"net/http"
	"net/url"

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

	var params []string
	var ok bool

	if ok, params = matcher.HeadercontainsQueryParamValue(result, ""); !ok {
		return
	}

	parsedURL, err := url.Parse(result.Url)
	if err != nil {
		logger.Debugf("Error parsing URL: %v\n", err)
	}

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
		resp, err := checker.CheckServerCustom(req, clients.NoRedirectClient)
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
	//TODO: continue from here
	logger.Infof("Request splitting found on %s: %v\n", result.Url, controllable)

}
