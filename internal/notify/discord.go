package notify

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/lormars/octohunter/internal/logger"
)

type WebhookMessage struct {
	Content string `json:"content"`
}

var client = &http.Client{}

func SendMessage(message string) error {
	webhookURL := os.Getenv("DISCORD")
	webhookMessage := WebhookMessage{
		Content: message,
	}

	payload, err := json.Marshal(webhookMessage)
	if err != nil {
		logger.Debugf("Error marshalling JSON: %v\n", err)
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(payload))
	if err != nil {
		logger.Debugf("Error creating request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	//use separate client for this
	resp, err := client.Do(req)

	if err != nil {
		logger.Debugf("Error sending message: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		logger.Debugf("Error sending message: %v\n", err)
		return err
	}

	logger.Infof("Message sent: %s\n", message)

	return nil

}
