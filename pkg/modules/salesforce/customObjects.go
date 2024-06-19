package salesforce

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/lormars/octohunter/common"
)

func PullCustomObjects(urlString string) (string, error) {
	auraContext := os.Getenv("AURACONTEXT")
	auraToken := os.Getenv("AURATOKEN")
	message := `{"actions":[{"id":"123;a","descriptor":"serviceComponent://ui.force.components.controllers.hostConfig.HostConfigController/ACTION$getConfigData","callingDescriptor":"UNKNOWN","params":{}}]}`
	jsonStr := []byte(`message=` + message + `&aura.context=` + auraContext + `&aura.token=` + auraToken)

	req, err := http.NewRequest("POST", urlString, bytes.NewBuffer(jsonStr))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := common.NormalClient.Do(req)
	if err != nil {
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	resp.Body.Close()

	fmt.Println(string(body))
	return string(body), nil
}
