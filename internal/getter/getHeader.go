package getter

import (
	"fmt"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/internal/logger"
)

func GetHeader(url, header string) (string, error) {
	resp, err := common.NoRedirectClient.Get(url)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", url, err)
		return "", err
	}
	defer resp.Body.Close()

	headerValue := resp.Header.Get(header)
	if headerValue == "" {
		logger.Debugf("Header %s not found\n", header)
		return "", fmt.Errorf("Header not found")
	}
	return headerValue, nil
}
