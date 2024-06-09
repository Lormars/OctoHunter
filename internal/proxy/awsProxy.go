package proxy

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func AwsProxy(query string) (bool, []byte) {
	endpoint := os.Getenv("AWS_ENDPOINT")
	api_key := os.Getenv("AWS_API_KEY")
	url := endpoint + "?q=" + query
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	req.Header.Set("x-api-key", api_key)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false, nil
	}
	if strings.Contains(string(body), "Our systems have detected unusual traffic from your computer network.") {
		return false, nil
	}
	return true, body
}
