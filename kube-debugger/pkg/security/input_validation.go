package security

import (
	"fmt"
	"regexp"
	"strings"
)

// ValidateAppName validates Kubernetes app/pod name format
func ValidateAppName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("app name cannot be empty")
	}
	if len(name) > 253 {
		return fmt.Errorf("app name exceeds maximum length of 253 characters")
	}

	// RFC 1123 subdomain validation for Kubernetes resources
	pattern := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	if !regexp.MustCompile(pattern).MatchString(name) {
		return fmt.Errorf("app name '%s' contains invalid characters; must match DNS subdomain rules", name)
	}
	return nil
}

// ValidateNamespace validates Kubernetes namespace name format
func ValidateNamespace(namespace string) error {
	if len(namespace) == 0 {
		return fmt.Errorf("namespace cannot be empty")
	}
	if len(namespace) > 63 {
		return fmt.Errorf("namespace exceeds maximum length of 63 characters")
	}

	pattern := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	if !regexp.MustCompile(pattern).MatchString(namespace) {
		return fmt.Errorf("namespace '%s' contains invalid characters", namespace)
	}
	return nil
}

// SanitizeInput removes potentially dangerous characters from user input
func SanitizeInput(input string) string {
	// Remove shell metacharacters and control characters
	dangerous := []string{";", "|", "&", "`", "$", "(", ")", "{", "}", "[", "]", "<", ">", "\000"}
	result := input
	for _, char := range dangerous {
		result = strings.ReplaceAll(result, char, "")
	}
	return strings.TrimSpace(result)
}

// ValidateInterval validates the watch interval value
func ValidateInterval(interval int) error {
	if interval < 1 {
		return fmt.Errorf("interval must be at least 1 second")
	}
	if interval > 3600 {
		return fmt.Errorf("interval cannot exceed 1 hour (3600 seconds)")
	}
	return nil
}

// ValidateThreshold validates health score threshold
func ValidateThreshold(threshold int) error {
	if threshold < 0 || threshold > 100 {
		return fmt.Errorf("threshold must be between 0 and 100")
	}
	return nil
}

// ValidateWebhookURL validates webhook URL format
func ValidateWebhookURL(url string) error {
	if len(url) == 0 {
		return fmt.Errorf("webhook URL cannot be empty")
	}
	if len(url) > 2048 {
		return fmt.Errorf("webhook URL exceeds maximum length")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("webhook URL must start with http:// or https://")
	}
	return nil
}
