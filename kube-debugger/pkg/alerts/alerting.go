package alerts

import (
	"time"
)

// Alert represents a system alert
type Alert struct {
	ID        string
	Title     string
	Severity  string // critical, warning, info
	Message   string
	Timestamp time.Time
	AppName   string
	Namespace string
	Details   map[string]string
}

// AlertManager manages alerts and sends them through multiple channels
type AlertManager struct {
	channels   map[string]AlertChannel
	dedup      *AlertDeduplicator
	throttler  *AlertThrottler
	history    []Alert
	maxHistory int
}

// AlertChannel interface for different alert channels
type AlertChannel interface {
	Name() string
	Send(alert Alert) error
	IsConfigured() bool
}

// AlertDeduplicator prevents duplicate alerts
type AlertDeduplicator struct {
	seen map[string]time.Time
}

// AlertThrottler prevents alert spam
type AlertThrottler struct {
	lastAlert map[string]time.Time
	interval  time.Duration
}

// New creates a new AlertManager
func New() *AlertManager {
	return &AlertManager{
		channels:   make(map[string]AlertChannel),
		history:    make([]Alert, 0),
		maxHistory: 1000,
		dedup: &AlertDeduplicator{
			seen: make(map[string]time.Time),
		},
		throttler: &AlertThrottler{
			lastAlert: make(map[string]time.Time),
			interval:  5 * time.Minute,
		},
	}
}

// RegisterChannel registers an alert channel
func (am *AlertManager) RegisterChannel(channel AlertChannel) {
	if channel.IsConfigured() {
		am.channels[channel.Name()] = channel
	}
}

// SendAlert sends an alert through configured channels
func (am *AlertManager) SendAlert(alert Alert) error {
	if alert.ID == "" {
		alert.ID = alert.AppName + ":" + alert.Title + ":" + alert.Severity
	}

	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}

	// Check deduplication
	if !am.dedup.ShouldSend(alert.ID) {
		return nil
	}

	// Check throttling
	if !am.throttler.ShouldSend(alert.ID) {
		return nil
	}

	// Send to all configured channels
	for _, channel := range am.channels {
		go func(ch AlertChannel) {
			_ = ch.Send(alert)
		}(channel)
	}

	// Add to history
	am.addToHistory(alert)
	return nil
}

// addToHistory adds alert to history with size limit
func (am *AlertManager) addToHistory(alert Alert) {
	am.history = append(am.history, alert)
	if len(am.history) > am.maxHistory {
		am.history = am.history[1:]
	}
}

// GetHistory returns alert history
func (am *AlertManager) GetHistory(limit int) []Alert {
	if limit > len(am.history) {
		limit = len(am.history)
	}
	return am.history[len(am.history)-limit:]
}

// ShouldSend checks if alert should be sent (deduplication)
func (ad *AlertDeduplicator) ShouldSend(id string) bool {
	if lastTime, exists := ad.seen[id]; exists {
		if time.Since(lastTime) < 10*time.Minute {
			return false
		}
	}
	ad.seen[id] = time.Now()
	return true
}

// ShouldSend checks if alert should be sent (throttling)
func (at *AlertThrottler) ShouldSend(id string) bool {
	if lastTime, exists := at.lastAlert[id]; exists {
		if time.Since(lastTime) < at.interval {
			return false
		}
	}
	at.lastAlert[id] = time.Now()
	return true
}

// AlertSeverity levels
const (
	SeverityCritical = "critical"
	SeverityWarning  = "warning"
	SeverityInfo     = "info"
)
