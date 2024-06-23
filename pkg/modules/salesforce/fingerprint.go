package salesforce

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/logger"
)

// Fingerprint checks if the target is running salesforce, taking url as input
func Fingerprint(target string) (bool, string) {
	path := []string{"/aura", "/s/sfsites/aura", "/sfsites/aura"}

	for _, p := range path {
		newUrl := strings.TrimSuffix(target, "/") + p
		fmt.Println(newUrl)
		jsonStr := []byte(`{}`)

		req, err := http.NewRequest("POST", newUrl, bytes.NewBuffer(jsonStr))
		if err != nil {
			logger.Debugf("Error creating request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		resp, err := clients.NormalClient.Do(req)
		if err != nil {
			logger.Debugf("Error getting response from %s: %v\n", newUrl, err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Debugf("Error reading response body: %v\n", err)
			continue
		}

		resp.Body.Close()
		if bytes.Contains(body, []byte("aura:invalidSession")) {
			return true, newUrl
		}
	}
	return false, ""
}
