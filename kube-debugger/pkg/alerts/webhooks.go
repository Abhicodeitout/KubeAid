package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// WebhookChannel sends alerts via generic webhooks
type WebhookChannel struct {
	url string
}

// SlackChannel sends alerts to Slack
type SlackChannel struct {
	webhook string
}

// EmailChannel sends alerts via email (stub for now)
type EmailChannel struct {
	sender   string
	password string
	smtpHost string
	smtpPort string
}

// NewWebhookChannel creates a webhook channel
func NewWebhookChannel(url string) *WebhookChannel {
	return &WebhookChannel{url: url}
}

// NewSlackChannel creates a Slack channel
func NewSlackChannel() *SlackChannel {
	return &SlackChannel{
		webhook: os.Getenv("KUBAID_SLACK_WEBHOOK"),
	}
}

// NewEmailChannel creates an email channel
func NewEmailChannel() *EmailChannel {
	return &EmailChannel{
		sender:   os.Getenv("KUBAID_EMAIL_SENDER"),
		password: os.Getenv("KUBAID_EMAIL_PASSWORD"),
		smtpHost: os.Getenv("KUBAID_SMTP_HOST"),
		smtpPort: os.Getenv("KUBAID_SMTP_PORT"),
	}
}

// Name returns the channel name
func (w *WebhookChannel) Name() string {
	return "webhook"
}

// Send sends alert via webhook
func (w *WebhookChannel) Send(alert Alert) error {
	payload := map[string]interface{}{
		"title":     alert.Title,
		"severity":  alert.Severity,
		"message":   alert.Message,
		"app":       alert.AppName,
		"namespace": alert.Namespace,
		"timestamp": alert.Timestamp,
		"details":   alert.Details,
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(w.url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return nil
}

// IsConfigured returns if webhook is configured
func (w *WebhookChannel) IsConfigured() bool {
	return w.url != ""
}

// Name returns the channel name
func (s *SlackChannel) Name() string {
	return "slack"
}

// Send sends alert to Slack
func (s *SlackChannel) Send(alert Alert) error {
	color := "#36a64f" // green
	switch alert.Severity {
	case SeverityCritical:
		color = "#ff0000" // red
	case SeverityWarning:
		color = "#ffaa00" // orange
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color":      color,
				"title":      alert.Title,
				"text":       alert.Message,
				"fields": []map[string]interface{}{
					{"title": "App", "value": alert.AppName, "short": true},
					{"title": "Namespace", "value": alert.Namespace, "short": true},
					{"title": "Severity", "value": alert.Severity, "short": true},
					{"title": "Time", "value": alert.Timestamp.Format("2006-01-02 15:04:05"), "short": true},
				},
				"ts": alert.Timestamp.Unix(),
			},
		},
	}

	data, _ := json.Marshal(payload)
	resp, err := http.Post(s.webhook, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	return nil
}

// IsConfigured returns if Slack is configured
func (s *SlackChannel) IsConfigured() bool {
	return s.webhook != ""
}

// Name returns the channel name
func (e *EmailChannel) Name() string {
	return "email"
}

// Send sends alert via email
func (e *EmailChannel) Send(alert Alert) error {
	// Email implementation - placeholder for brevity
	fmt.Printf("Email Alert: %s - %s\n", alert.Title, alert.Message)
	return nil
}

// IsConfigured returns if email is configured
func (e *EmailChannel) IsConfigured() bool {
	return e.sender != "" && e.password != "" && e.smtpHost != ""
}
