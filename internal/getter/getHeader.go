package getter

import (
	"fmt"
	"net/http"

	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

func GetHeader(urlStr, header string) (string, error) {
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
		logger.Debugf("Error creating request: %v", err)
		return "", err
	}
	resp, err := checker.CheckServerCustom(req, http.DefaultClient)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", urlStr, err)
		return "", err
	}

	headerValue := resp.Headers.Get(header)
	if headerValue == "" {
		logger.Debugf("Header %s not found\n", header)
		return "", fmt.Errorf("Header not found")
	}
	return headerValue, nil
}
