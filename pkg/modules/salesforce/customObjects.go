package salesforce

import (
	"bytes"
	"net/http"
	"os"
	"regexp"

	"github.com/lormars/octohunter/common"
	"github.com/lormars/octohunter/common/clients"
	"github.com/lormars/octohunter/internal/checker"
	"github.com/lormars/octohunter/internal/logger"
	"github.com/lormars/octohunter/internal/notify"
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
	resp, err := checker.CheckServerCustom(req, clients.NormalClient)
	if err != nil {
		logger.Debugf("Error getting response from %s: %v\n", urlString, err)
		return err
	}

	re := regexp.MustCompile(`\b\w+__c\b`)
	matches := re.FindAllString(string(resp.Body), -1)
	for _, match := range matches {
		msg := "[Salesforce] Custom Object Found: " + match
		if common.SendOutput {
			common.OutputP.PublishMessage(msg)
		}
		notify.SendMessage(msg)
	}

	return nil
}
