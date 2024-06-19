package getter

import (
	"fmt"

	"github.com/lormars/octohunter/common"
)

func GetHeader(url, header string) (string, error) {
	resp, err := common.NoRedirectClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	headerValue := resp.Header.Get(header)
	if headerValue == "" {
		return "", fmt.Errorf("Header not found")
	}
	return headerValue, nil
}
