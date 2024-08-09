package salesforce

import (
	"bytes"
	"net/http"
	"strings"

	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
)

// Fingerprint checks if the target is running salesforce, taking url as input
func Fingerprint(target string) (bool, string) {
	path := []string{"/aura", "/s/sfsites/aura", "/sfsites/aura"}

	for _, p := range path {
		newUrl := strings.TrimSuffix(target, "/") + p
		jsonStr := []byte(`{}`)

		req, err := http.NewRequest("POST", newUrl, bytes.NewBuffer(jsonStr))
		if err != nil {
			logger.Debugf("Error creating request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		octoReq := &clients.OctoRequest{
			Request:  req,
			Producer: clients.Salesforce,
		}
		resp, err := checker.CheckServerCustom(octoReq, clients.Clients.GetRandomClient("h0", true, true))
		if err != nil {
			logger.Debugf("Error getting response from %s: %v\n", newUrl, err)
			continue
		}

		if bytes.Contains([]byte(resp.Body), []byte("aura:invalidSession")) {
			return true, newUrl
		}
	}
	return false, ""
}
