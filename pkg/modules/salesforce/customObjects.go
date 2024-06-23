package salesforce

import (
	"bytes"
	"io"
	"net/http"
	"os"
	"regexp"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/logger"
)

func PullCustomObjects(urlString string) error {
	auraContext := os.Getenv("AURACONTEXT")
	auraToken := os.Getenv("AURATOKEN")
	message := `{"actions":[{"id":"123;a","descriptor":"serviceComponent://ui.force.components.controllers.hostConfig.HostConfigController/ACTION$getConfigData","callingDescriptor":"UNKNOWN","params":{}}]}`
	bodyStr := []byte(`message=` + message + `&aura.context=` + auraContext + `&aura.token=` + auraToken)

	req, err := http.NewRequest("POST", urlString, bytes.NewBuffer(bodyStr))
	if err != nil {
		logger.Debugf("Error creating request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := clients.NormalClient.Do(req)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", urlString, err)
		return err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debugf("Error reading response body: %v\n", err)
		return err
	}

	resp.Body.Close()
	re := regexp.MustCompile(`\b\w+__c\b`)
	matches := re.FindAllString(string(body), -1)
	for _, match := range matches {
		msg := "[Salesforce] Custom Object Found: " + match
		common.OutputP.PublishMessage(msg)
	}

	return nil
}
