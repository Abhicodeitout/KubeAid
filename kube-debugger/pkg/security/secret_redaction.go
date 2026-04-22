package security

import (
	"regexp"
	"strings"
)

// Common patterns for secrets that should be redacted
var secretPatterns = map[string]*regexp.Regexp{
	"api_key":       regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*([^\s'"]+)`),
	"token":         regexp.MustCompile(`(?i)(token|bearer)\s*[:=]\s*([^\s'"]+)`),
	"password":      regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*([^\s'"]+)`),
	"secret":        regexp.MustCompile(`(?i)(secret|api[_-]?secret)\s*[:=]\s*([^\s'"]+)`),
	"key":           regexp.MustCompile(`(?i)(private[_-]?key|private_key)\s*[:=]\s*([^\s'"]+)`),
	"auth":          regexp.MustCompile(`(?i)(authorization|auth)\s*[:=]\s*([^\s'"]+)`),
	"aws":           regexp.MustCompile(`(?i)(aws[_-]?secret|aws[_-]?access[_-]?key)\s*[:=]\s*([^\s'"]+)`),
	"docker_auth":   regexp.MustCompile(`(?i)(docker[_-]?auth|auths)\s*[:=]\s*([^\s'"]+)`),
	"jwt":           regexp.MustCompile(`(eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+)`),
	"base64_auth":   regexp.MustCompile(`(?i)(authorization:\s*basic\s+)([A-Za-z0-9+/=]+)`),
	"kubeconfig":    regexp.MustCompile(`(?i)(certificate-authority-data|client-certificate-data|client-key-data):\s*([^\n]+)`),
}

// SensitiveFields are field names that often contain secrets
var sensitiveFields = []string{
	"password", "passwd", "pwd",
	"secret", "token", "bearer",
	"api_key", "apikey", "api-key",
	"auth", "authorization",
	"credential", "credentials",
	"key", "private_key",
	"cert", "certificate",
	"tls_key", "ssl_key",
	"aws_secret_access_key",
	"docker_auth",
	"kubeconfig",
	"env_password",
	"oauth_token",
}

// RedactSecrets removes or masks sensitive information from a string
func RedactSecrets(input string) string {
	output := input

	// Redact JWT tokens
	output = secretPatterns["jwt"].ReplaceAllString(output, "[REDACTED_JWT]")

	// Redact basic auth
	output = secretPatterns["base64_auth"].ReplaceAllString(output, "$1[REDACTED_BASE64]")

	// Redact kubeconfig credentials
	for key, pattern := range secretPatterns {
		if key == "jwt" || key == "base64_auth" {
			continue
		}
		output = pattern.ReplaceAllString(output, "${1}[REDACTED]")
	}

	// Redact kubeconfig specific fields
	output = secretPatterns["kubeconfig"].ReplaceAllString(output, "$1[REDACTED_CERT_DATA]")

	return output
}

// MaskInLogs provides a safe representation of potentially sensitive data
func MaskInLogs(value string, fieldName string) string {
	fieldNameLower := strings.ToLower(fieldName)

	// Check if field name suggests sensitive content
	for _, sensitive := range sensitiveFields {
		if strings.Contains(fieldNameLower, sensitive) {
			if len(value) <= 4 {
				return "[REDACTED]"
			}
			// Show first 2 and last 2 characters for non-empty values
			return value[:2] + "[REDACTED]" + value[len(value)-2:]
		}
	}
	return value
}

// RedactStruct redacts common fields in a struct represented as a map
func RedactStruct(data map[string]interface{}) map[string]interface{} {
	redacted := make(map[string]interface{})

	for key, value := range data {
		keyLower := strings.ToLower(key)
		isSensitive := false

		for _, sensitive := range sensitiveFields {
			if strings.Contains(keyLower, sensitive) {
				isSensitive = true
				break
			}
		}

		if isSensitive {
			redacted[key] = "[REDACTED]"
		} else if str, ok := value.(string); ok {
			redacted[key] = RedactSecrets(str)
		} else {
			redacted[key] = value
		}
	}

	return redacted
}

// RedactEnvironmentVariables returns environment variables with secrets removed
func RedactEnvironmentVariables(env []string) []string {
	var redacted []string
	for _, e := range env {
		redacted = append(redacted, RedactSecrets(e))
	}
	return redacted
}
