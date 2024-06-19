package getter

import (
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	},
	Timeout: 10 * time.Second,
}

func GetHeader(url, header string) (string, error) {
	resp, err := httpClient.Get(url)
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
