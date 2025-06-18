package notifications

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// lastNotificationTime tracks when the last notification was sent for each probeId
var lastNotificationTime = make(map[string]time.Time)
var lastNotificationMutex sync.Mutex

// SendNotifications sends error notifications to a Slack channel if configured
// It takes a list of error strings and sends them as a single Slack message
// Notifications are throttled based on MIN_DELAY_BETWEEN_NOTIFICATIONS_SECONDS per probeId
func SendNotifications(probeName string, probeId string, errors []string) {
	contextLogger := log.WithFields(log.Fields{
		"name":  probeName,
		"id":    probeId})

	if len(errors) == 0 {
		contextLogger.Debug("No error")
		return
	}

	slackWebhookURL := os.Getenv("SLACK_WEBHOOK_URL")
	if slackWebhookURL == "" {
		contextLogger.Debug("SLACK_WEBHOOK_URL not set, skipping Slack notifications")
		return
	}
	
	// Check if we should throttle this notification
	lastNotificationMutex.Lock()
	lastSent, exists := lastNotificationTime[probeId]
	lastNotificationMutex.Unlock()

	if exists {
		currentTime := time.Now()
		timeSinceLastNotification := currentTime.Sub(lastSent)
		minDelay, _ := strconv.Atoi(os.Getenv("MIN_DELAY_BETWEEN_NOTIFICATIONS_SECONDS"))
		if minDelay == 0 {
			minDelay = 1200
		}
		if timeSinceLastNotification.Seconds() < float64(minDelay) {
			contextLogger.Debug("Notification throttled")
			return
		}
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
		contextLogger.WithError(err).Error("Failed to marshal Slack message")
		return
	}

	resp, err := http.Post(slackWebhookURL, "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		contextLogger.WithError(err).Error("Failed to send message to Slack")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		contextLogger.WithField("status", resp.StatusCode).Error("Received non-OK response from Slack")
	} else {
		contextLogger.Debug("Successfully sent notifications to Slack")
		
		// Update the last notification time for this probeId
		lastNotificationMutex.Lock()
		lastNotificationTime[probeId] = time.Now()
		lastNotificationMutex.Unlock()
	}
}
