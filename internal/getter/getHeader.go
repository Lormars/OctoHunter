package getter

import (
	"fmt"

	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

func GetHeader(urlStr, header string) (string, error) {
	req, err := clients.NewRequest("GET", urlStr, nil, clients.Misc)
	if err != nil {
		logger.Debugf("Error creating request: %v", err)
		return "", err
	}
	resp, err := checker.CheckServerCustom(req, clients.Clients.GetRandomClient("h0", false, true))
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", urlStr, err)
		return "", err
	}

	headerValue := resp.Headers.Get(header)
	if headerValue == "" {
		logger.Debugf("Header %s not found\n", header)
		return "", fmt.Errorf("header not found")
	}
	return headerValue, nil
}
