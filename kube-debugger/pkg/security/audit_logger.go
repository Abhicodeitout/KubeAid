package security

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// AuditEvent represents a security audit log entry
type AuditEvent struct {
	Timestamp   time.Time              `json:"timestamp"`
	EventType   string                 `json:"event_type"`
	Username    string                 `json:"username"`
	Command     string                 `json:"command"`
	Arguments   []string               `json:"arguments"`
	AppName     string                 `json:"app_name,omitempty"`
	Namespace   string                 `json:"namespace,omitempty"`
	Status      string                 `json:"status"`
	ErrorMsg    string                 `json:"error,omitempty"`
	SourceIP    string                 `json:"source_ip,omitempty"`
	ClusterName string                 `json:"cluster_name,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// AuditLogger handles audit logging operations
type AuditLogger struct {
	mu        sync.Mutex
	logFile   string
	enabled   bool
	maxSize   int64 // Maximum log file size in bytes
	maxBackup int   // Maximum number of backup files
}

var (
	globalAuditLogger *AuditLogger
	once              sync.Once
)

// InitAuditLogger initializes the global audit logger
func InitAuditLogger(logDir string, enabled bool) error {
	var err error
	once.Do(func() {
		if !enabled {
			globalAuditLogger = &AuditLogger{enabled: false}
			return
		}

		// Create log directory if it doesn't exist
		if err = os.MkdirAll(logDir, 0700); err != nil {
			return
		}

		logPath := filepath.Join(logDir, "kube-debugger-audit.log")
		globalAuditLogger = &AuditLogger{
			logFile:   logPath,
			enabled:   true,
			maxSize:   10485760, // 10MB
			maxBackup: 10,
		}
	})
	return err
}

// GetAuditLogger returns the global audit logger
func GetAuditLogger() *AuditLogger {
	if globalAuditLogger == nil {
		globalAuditLogger = &AuditLogger{enabled: false}
	}
	return globalAuditLogger
}

// LogEvent logs an audit event
func (al *AuditLogger) LogEvent(event *AuditEvent) error {
	if !al.enabled {
		return nil
	}

	al.mu.Lock()
	defer al.mu.Unlock()

	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	if event.Username == "" {
		event.Username = os.Getenv("USER")
	}

	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal audit event: %w", err)
	}

	// Check if log file needs rotation
	if err := al.rotateIfNeeded(); err != nil {
		return err
	}

	// Append to log file
	f, err := os.OpenFile(al.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open audit log: %w", err)
	}
	defer f.Close()

	_, err = f.WriteString(string(data) + "\n")
	return err
}

// rotateIfNeeded rotates the log file if it exceeds max size
func (al *AuditLogger) rotateIfNeeded() error {
	fileInfo, err := os.Stat(al.logFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	if fileInfo.Size() < al.maxSize {
		return nil
	}

	// Rotate old backups
	for i := al.maxBackup - 1; i >= 1; i-- {
		oldName := fmt.Sprintf("%s.%d", al.logFile, i)
		newName := fmt.Sprintf("%s.%d", al.logFile, i+1)
		if _, err := os.Stat(oldName); err == nil {
			os.Rename(oldName, newName)
		}
	}

	// Rename current file to .1
	return os.Rename(al.logFile, al.logFile+".1")
}

// LogCommand logs a command execution
func (al *AuditLogger) LogCommand(cmd string, args []string, appName, namespace string, err error) error {
	event := &AuditEvent{
		EventType: "command_execution",
		Command:   cmd,
		Arguments: args,
		AppName:   appName,
		Namespace: namespace,
		Status:    "success",
	}

	if err != nil {
		event.Status = "failure"
		event.ErrorMsg = err.Error()
	}

	return al.LogEvent(event)
}

// LogKubeAction logs Kubernetes API actions
func (al *AuditLogger) LogKubeAction(action, resource, appName, namespace string, err error) error {
	event := &AuditEvent{
		EventType: "kubernetes_action",
		Command:   action,
		AppName:   appName,
		Namespace: namespace,
		Metadata: map[string]interface{}{
			"resource": resource,
		},
		Status: "success",
	}

	if err != nil {
		event.Status = "failure"
		event.ErrorMsg = err.Error()
	}

	return al.LogEvent(event)
}

// LogSecurityEvent logs security-related events
func (al *AuditLogger) LogSecurityEvent(eventType, severity string, details map[string]interface{}) error {
	event := &AuditEvent{
		EventType: eventType,
		Status:    severity,
		Metadata:  details,
	}

	return al.LogEvent(event)
}

// ReadAuditLogs returns recent audit logs
func (al *AuditLogger) ReadAuditLogs(count int) ([]AuditEvent, error) {
	if !al.enabled {
		return nil, fmt.Errorf("audit logging is not enabled")
	}

	data, err := os.ReadFile(al.logFile)
	if err != nil {
		return nil, err
	}

	var events []AuditEvent
	lines := string(data)
	// Parse last `count` lines
	lineArray := strings.Split(strings.TrimSpace(lines), "\n")

	start := len(lineArray) - count
	if start < 0 {
		start = 0
	}

	for i := start; i < len(lineArray); i++ {
		var event AuditEvent
		if err := json.Unmarshal([]byte(lineArray[i]), &event); err == nil {
			events = append(events, event)
		}
	}

	return events, nil
}
