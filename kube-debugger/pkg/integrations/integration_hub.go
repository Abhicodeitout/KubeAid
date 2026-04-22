package integrations

import (
	"fmt"
	"os"
)

// Integration interface for different channels
type Integration interface {
	Name() string
	Configure(config map[string]string) error
	SendAlert(title, message, severity string) error
	IsConfigured() bool
}

// IntegrationHub manages all integrations
type IntegrationHub struct {
	integrations map[string]Integration
}

// New creates a new integration hub
func New() *IntegrationHub {
	return &IntegrationHub{
		integrations: make(map[string]Integration),
	}
}

// RegisterIntegration registers an integration
func (ih *IntegrationHub) RegisterIntegration(integration Integration) {
	if integration.IsConfigured() {
		ih.integrations[integration.Name()] = integration
	}
}

// SendAlertToAll sends alert to all configured integrations
func (ih *IntegrationHub) SendAlertToAll(title, message, severity string) error {
	var lastErr error

	for _, integration := range ih.integrations {
		if err := integration.SendAlert(title, message, severity); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// SlackIntegration sends alerts to Slack
type SlackIntegration struct {
	webhook string
}

// NewSlackIntegration creates a Slack integration
func NewSlackIntegration() *SlackIntegration {
	return &SlackIntegration{
		webhook: os.Getenv("KUBAID_SLACK_WEBHOOK"),
	}
}

// Name returns integration name
func (s *SlackIntegration) Name() string {
	return "slack"
}

// Configure configures Slack integration
func (s *SlackIntegration) Configure(config map[string]string) error {
	if webhook, ok := config["webhook"]; ok {
		s.webhook = webhook
	}
	return nil
}

// SendAlert sends alert to Slack
func (s *SlackIntegration) SendAlert(title, message, severity string) error {
	fmt.Printf("[Slack] Alert: %s - %s (severity: %s)\n", title, message, severity)
	// Real implementation would use Slack SDK
	return nil
}

// IsConfigured checks if Slack is configured
func (s *SlackIntegration) IsConfigured() bool {
	return s.webhook != ""
}

// PagerDutyIntegration sends alerts to PagerDuty
type PagerDutyIntegration struct {
	apiKey string
}

// NewPagerDutyIntegration creates a PagerDuty integration
func NewPagerDutyIntegration() *PagerDutyIntegration {
	return &PagerDutyIntegration{
		apiKey: os.Getenv("KUBAID_PAGERDUTY_KEY"),
	}
}

// Name returns integration name
func (p *PagerDutyIntegration) Name() string {
	return "pagerduty"
}

// Configure configures PagerDuty integration
func (p *PagerDutyIntegration) Configure(config map[string]string) error {
	if apiKey, ok := config["apiKey"]; ok {
		p.apiKey = apiKey
	}
	return nil
}

// SendAlert sends alert to PagerDuty
func (p *PagerDutyIntegration) SendAlert(title, message, severity string) error {
	fmt.Printf("[PagerDuty] Alert: %s - %s (severity: %s)\n", title, message, severity)
	// Real implementation would use PagerDuty API
	return nil
}

// IsConfigured checks if PagerDuty is configured
func (p *PagerDutyIntegration) IsConfigured() bool {
	return p.apiKey != ""
}

// EmailIntegration sends alerts via email
type EmailIntegration struct {
	sender   string
	password string
	smtpHost string
	smtpPort string
}

// NewEmailIntegration creates an email integration
func NewEmailIntegration() *EmailIntegration {
	return &EmailIntegration{
		sender:   os.Getenv("KUBAID_EMAIL_SENDER"),
		password: os.Getenv("KUBAID_EMAIL_PASSWORD"),
		smtpHost: os.Getenv("KUBAID_SMTP_HOST"),
		smtpPort: os.Getenv("KUBAID_SMTP_PORT"),
	}
}

// Name returns integration name
func (e *EmailIntegration) Name() string {
	return "email"
}

// Configure configures email integration
func (e *EmailIntegration) Configure(config map[string]string) error {
	if sender, ok := config["sender"]; ok {
		e.sender = sender
	}
	if host, ok := config["host"]; ok {
		e.smtpHost = host
	}
	return nil
}

// SendAlert sends alert via email
func (e *EmailIntegration) SendAlert(title, message, severity string) error {
	fmt.Printf("[Email] Alert: %s - %s (severity: %s)\n", title, message, severity)
	// Real implementation would use SMTP
	return nil
}

// IsConfigured checks if email is configured
func (e *EmailIntegration) IsConfigured() bool {
	return e.sender != "" && e.smtpHost != ""
}

// DatadogIntegration sends metrics to Datadog
type DatadogIntegration struct {
	apiKey string
}

// NewDatadogIntegration creates a Datadog integration
func NewDatadogIntegration() *DatadogIntegration {
	return &DatadogIntegration{
		apiKey: os.Getenv("KUBAID_DATADOG_KEY"),
	}
}

// Name returns integration name
func (d *DatadogIntegration) Name() string {
	return "datadog"
}

// Configure configures Datadog integration
func (d *DatadogIntegration) Configure(config map[string]string) error {
	if apiKey, ok := config["apiKey"]; ok {
		d.apiKey = apiKey
	}
	return nil
}

// SendAlert sends alert to Datadog
func (d *DatadogIntegration) SendAlert(title, message, severity string) error {
	fmt.Printf("[Datadog] Alert: %s - %s (severity: %s)\n", title, message, severity)
	// Real implementation would use Datadog API
	return nil
}

// IsConfigured checks if Datadog is configured
func (d *DatadogIntegration) IsConfigured() bool {
	return d.apiKey != ""
}
