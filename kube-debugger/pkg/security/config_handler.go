package security

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigHandler provides secure configuration management
type ConfigHandler struct {
	configDir string
	auditLog  *AuditLogger
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(configDir string) *ConfigHandler {
	return &ConfigHandler{
		configDir: configDir,
		auditLog:  GetAuditLogger(),
	}
}

// ValidateKubeconfig validates the kubeconfig file
func (ch *ConfigHandler) ValidateKubeconfig() error {
	kubeconfigPath := getKubeconfigPath()

	if kubeconfigPath == "" {
		return fmt.Errorf("KUBECONFIG environment variable not set and ~/.kube/config not found")
	}

	// Check if file exists
	if _, err := os.Stat(kubeconfigPath); err != nil {
		return fmt.Errorf("kubeconfig file not found: %s", kubeconfigPath)
	}

	// Check file permissions (should not be world-readable)
	fileInfo, err := os.Stat(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to stat kubeconfig: %w", err)
	}

	mode := fileInfo.Mode()
	if mode&0077 != 0 { // Check if group or others have any permissions
		return fmt.Errorf("kubeconfig file has insecure permissions (%o); should be 0600 or more restrictive", mode)
	}

	return nil
}

// SecureKubeconfig ensures kubeconfig has secure permissions
func (ch *ConfigHandler) SecureKubeconfig() error {
	kubeconfigPath := getKubeconfigPath()

	if kubeconfigPath == "" {
		return fmt.Errorf("kubeconfig path cannot be determined")
	}

	if _, err := os.Stat(kubeconfigPath); err != nil {
		return fmt.Errorf("kubeconfig file not found: %s", kubeconfigPath)
	}

	// Set permissions to 0600 (owner read/write only)
	if err := os.Chmod(kubeconfigPath, 0600); err != nil {
		return fmt.Errorf("failed to secure kubeconfig permissions: %w", err)
	}

	ch.auditLog.LogSecurityEvent("kubeconfig_secured", "info", map[string]interface{}{
		"path": kubeconfigPath,
	})

	return nil
}

// ValidateEnvironmentVariables checks for secure environment configuration
func (ch *ConfigHandler) ValidateEnvironmentVariables() []string {
	var warnings []string

	// Check insecure skip verify
	if os.Getenv("KUBECONFIG_INSECURE_SKIP_VERIFY") == "true" {
		warnings = append(warnings, "KUBECONFIG_INSECURE_SKIP_VERIFY is enabled - certificate verification disabled")
	}

	// Check for potentially sensitive env vars
	sensitiveEnvVars := []string{
		"KUBECONFIG_PASSWORD",
		"KUBERNETES_PASSWORD",
		"API_KEY",
		"API_SECRET",
		"TOKEN",
		"SECRET",
		"AWS_SECRET_ACCESS_KEY",
	}

	for _, envVar := range sensitiveEnvVars {
		if val := os.Getenv(envVar); val != "" {
			if len(val) > 0 {
				warnings = append(warnings, fmt.Sprintf("Found sensitive environment variable: %s", envVar))
			}
		}
	}

	return warnings
}

// CreateSecureConfigDir creates a secure configuration directory
func (ch *ConfigHandler) CreateSecureConfigDir() error {
	if ch.configDir == "" {
		return fmt.Errorf("config directory path not set")
	}

	// Create directory with secure permissions (0700 - owner only)
	if err := os.MkdirAll(ch.configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Verify permissions
	fileInfo, err := os.Stat(ch.configDir)
	if err != nil {
		return fmt.Errorf("failed to stat config directory: %w", err)
	}

	mode := fileInfo.Mode()
	if mode&0077 != 0 {
		if err := os.Chmod(ch.configDir, 0700); err != nil {
			return fmt.Errorf("failed to secure config directory permissions: %w", err)
		}
	}

	return nil
}

// LoadSecureConfig loads configuration from a file with validation
func (ch *ConfigHandler) LoadSecureConfig(filename string) (string, error) {
	configPath := filepath.Join(ch.configDir, filename)

	// Prevent directory traversal attacks
	if !strings.HasPrefix(filepath.Clean(configPath), ch.configDir) {
		return "", fmt.Errorf("invalid config file path (potential directory traversal)")
	}

	// Check file existence
	if _, err := os.Stat(configPath); err != nil {
		return "", fmt.Errorf("config file not found: %s", configPath)
	}

	// Check file permissions
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to stat config file: %w", err)
	}

	mode := fileInfo.Mode()
	if mode&0077 != 0 {
		return "", fmt.Errorf("config file has insecure permissions (%o); should be 0600 or more restrictive", mode)
	}

	// Read config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	// Redact secrets before logging
	redactedData := RedactSecrets(string(data))
	ch.auditLog.LogSecurityEvent("config_loaded", "info", map[string]interface{}{
		"path": configPath,
		"data": redactedData,
	})

	return string(data), nil
}

// SaveSecureConfig saves configuration to a file with secure permissions
func (ch *ConfigHandler) SaveSecureConfig(filename, content string) error {
	// Create config directory if needed
	if err := ch.CreateSecureConfigDir(); err != nil {
		return err
	}

	configPath := filepath.Join(ch.configDir, filename)

	// Prevent directory traversal attacks
	if !strings.HasPrefix(filepath.Clean(configPath), ch.configDir) {
		return fmt.Errorf("invalid config file path (potential directory traversal)")
	}

	// Write config with secure permissions
	if err := os.WriteFile(configPath, []byte(content), 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	ch.auditLog.LogSecurityEvent("config_saved", "info", map[string]interface{}{
		"path": configPath,
	})

	return nil
}

// getKubeconfigPath returns the kubeconfig path
func getKubeconfigPath() string {
	if fromEnv := strings.TrimSpace(os.Getenv("KUBECONFIG")); fromEnv != "" {
		// KUBECONFIG may contain multiple paths; use the first as the primary source.
		return strings.Split(fromEnv, string(os.PathListSeparator))[0]
	}

	home := os.Getenv("HOME")
	if home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}

// IsRunningInCluster checks if code is running in a Kubernetes cluster
func IsRunningInCluster() bool {
	// Check for in-cluster token
	_, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token")
	return err == nil
}

// CheckSecurityContext validates security context for CLI execution
func CheckSecurityContext() map[string]interface{} {
	context := make(map[string]interface{})

	// Check if running as root
	context["running_as_root"] = os.Geteuid() == 0

	// Check if in cluster
	context["running_in_cluster"] = IsRunningInCluster()

	// Check kubeconfig security
	kubeconfigPath := getKubeconfigPath()
	if kubeconfigPath != "" {
		if fileInfo, err := os.Stat(kubeconfigPath); err == nil {
			mode := fileInfo.Mode()
			context["kubeconfig_perms_secure"] = (mode&0077 == 0)
		}
	}

	// Check environment variable warnings
	warnings := validateEnvironmentVariables()
	if len(warnings) > 0 {
		context["env_warnings"] = warnings
	}

	return context
}

// validateEnvironmentVariables checks for insecure environment variable usage
func validateEnvironmentVariables() []string {
	var warnings []string

	// Check insecure skip verify
	if os.Getenv("KUBECONFIG_INSECURE_SKIP_VERIFY") == "true" {
		warnings = append(warnings, "certificate verification disabled")
	}

	return warnings
}
