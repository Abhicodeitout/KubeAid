package security

import (
	"regexp"
	"strings"
)

// OutputFilter provides methods to filter sensitive information from output
type OutputFilter struct {
	redactSecrets bool
	redactPaths   bool
}

// NewOutputFilter creates a new output filter
func NewOutputFilter(redactSecrets, redactPaths bool) *OutputFilter {
	return &OutputFilter{
		redactSecrets: redactSecrets,
		redactPaths:   redactPaths,
	}
}

// FilterOutput applies all filters to output
func (of *OutputFilter) FilterOutput(output string) string {
	result := output

	if of.redactSecrets {
		result = RedactSecrets(result)
	}

	if of.redactPaths {
		result = of.redactFilePaths(result)
	}

	return result
}

// redactFilePaths removes or masks file paths from output
func (of *OutputFilter) redactFilePaths(output string) string {
	result := output

	// Redact home directory paths
	homePattern := regexp.MustCompile(`/home/[^/\s:;,'"]+`)
	result = homePattern.ReplaceAllString(result, "[HOME]")

	// Redact root paths that might contain sensitive info
	sensitivePattern := regexp.MustCompile(`/var/(log|run|cache|lock)/[^\s:;,'"]+`)
	result = sensitivePattern.ReplaceAllString(result, "[REDACTED_PATH]")

	return result
}

// FilterLogLine filters a single log line
func (of *OutputFilter) FilterLogLine(line string) string {
	return of.FilterOutput(line)
}

// FilterJSON filters a JSON string (basic approach without full parsing)
func (of *OutputFilter) FilterJSON(jsonStr string) string {
	result := jsonStr

	if of.redactSecrets {
		// Redact JSON values for sensitive keys
		sensitiveKeys := []string{"password", "secret", "token", "api_key", "authorization", "auth"}
		for _, key := range sensitiveKeys {
			pattern := regexp.MustCompile(`"` + key + `":\s*"[^"]*"`)
			result = pattern.ReplaceAllString(result, `"`+key+`":"[REDACTED]"`)
		}
	}

	if of.redactPaths {
		result = of.redactFilePaths(result)
	}

	return result
}

// FilterLines applies filter to each line of output
func (of *OutputFilter) FilterLines(lines []string) []string {
	var filtered []string
	for _, line := range lines {
		filtered = append(filtered, of.FilterOutput(line))
	}
	return filtered
}

// SanitizeOutput removes potentially dangerous characters from output
func (of *OutputFilter) SanitizeOutput(output string) string {
	// Remove null bytes
	result := strings.ReplaceAll(output, "\x00", "")

	// Remove other control characters (except newline, tab, carriage return)
	controlCharPattern := regexp.MustCompile(`[\x00-\x08\x0B-\x0C\x0E-\x1F]`)
	result = controlCharPattern.ReplaceAllString(result, "")

	return result
}

// TruncateOutput truncates output to a maximum length
func (of *OutputFilter) TruncateOutput(output string, maxLength int) string {
	if len(output) <= maxLength {
		return output
	}
	return output[:maxLength] + "\n... (output truncated)"
}

// MaskEmailAddresses masks email addresses in output
func (of *OutputFilter) MaskEmailAddresses(output string) string {
	emailPattern := regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	return emailPattern.ReplaceAllString(output, "[EMAIL_REDACTED]")
}

// MaskIPAddresses masks IP addresses in output (handles both IPv4 and IPv6)
func (of *OutputFilter) MaskIPAddresses(output string) string {
	// IPv4 pattern
	ipv4Pattern := regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	result := ipv4Pattern.ReplaceAllString(output, "[IP_REDACTED]")

	// IPv6 pattern (simplified)
	ipv6Pattern := regexp.MustCompile(`\b(?:[0-9a-fA-F]{0,4}:){2,7}[0-9a-fA-F]{0,4}\b`)
	result = ipv6Pattern.ReplaceAllString(result, "[IP_REDACTED]")

	return result
}

// FilterSensitiveFields filters output by removing lines containing sensitive field names
func FilterSensitiveFields(output string) string {
	lines := strings.Split(output, "\n")
	var filtered []string

	for _, line := range lines {
		isSensitive := false
		for _, sensitive := range sensitiveFields {
			if strings.Contains(strings.ToLower(line), strings.ToLower(sensitive)) {
				isSensitive = true
				break
			}
		}
		if !isSensitive {
			filtered = append(filtered, line)
		}
	}

	return strings.Join(filtered, "\n")
}
