package salesforce

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func Fingerprint(target string) (bool, string) {
	path := []string{"/aura", "/s/sfsites/aura", "/sfsites/aura"}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	for _, p := range path {
		newUrl := strings.TrimSuffix(target, "/") + p
		fmt.Println(newUrl)
		jsonStr := []byte(`{}`)

		req, err := http.NewRequest("POST", newUrl, bytes.NewBuffer(jsonStr))
		if err != nil {
			return false, ""
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			return false, ""
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}

		resp.Body.Close()
		if bytes.Contains(body, []byte("aura:invalidSession")) {
			return true, newUrl
		}
	}
	return false, ""
}
