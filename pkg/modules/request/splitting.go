package request

import (
	"fmt"
	"net/url"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/cacher"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/matcher"
)

func RequestSplitting(result *common.ServerResult) {
	if !cacher.CheckCache(result.Url, "split") {
		return
	}

	var params []string
	var ok bool

	if ok, params = matcher.HeadercontainsQueryParamValue(result); !ok {
		return
	}

	parsedURL, err := url.Parse(result.Url)
	if err != nil {
		logger.Debugf("Error parsing URL: %v\n", err)
	}

	queryParams := parsedURL.Query()
	fmt.Println(queryParams)
	fmt.Print(params)

}
