package getter

import (
	"fmt"
	"net/http"
)

func GetHeader(url, header string) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	header_value := resp.Header.Get(header)
	if header_value == "" {
		return "", fmt.Errorf("Header not found")
	}
	return header_value, nil

}
