package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// SendNotifications sends error notifications to a Slack channel if configured
// It takes a list of error strings and sends them as a single Slack message
func SendNotifications(probeName string, probeId string, errors []string) {
	if len(errors) == 0 {
		log.Debug("No error")
		return
	}

	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackWebhookURL == "" {
		log.Debug("SLACK_WEBHOOK_URL not set, skipping Slack notifications")
		return
	}

	messageBody := fmt.Sprintf("Probe failure for %s", probeId)
	if probeName != "" {
		messageBody += fmt.Sprintf(" (%s)", probeName)
	}
	messageBody += "\n"
	
	for _, errorMsg := range errors {
		messageBody += fmt.Sprintf("%s\n", errorMsg)
	}

	jsonBody, err := json.Marshal(map[string]string{
		"text": messageBody,
	})
	if err != nil {
		log.WithError(err).Error("Failed to marshal Slack message")
		return
	}

	resp, err := http.Post(slackWebhookURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		log.WithError(err).Error("Failed to send message to Slack")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.WithField("status", resp.StatusCode).Error("Received non-OK response from Slack")
	} else {
		log.Debug("Successfully sent notifications to Slack")
	}
}
