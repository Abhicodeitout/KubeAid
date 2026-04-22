package security

import (
	"testing"
)

func TestSecretRedaction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "API Key Redaction",
			input:    "api_key=sk-12345abcdef",
			expected: "[REDACTED]",
		},
		{
			name:     "Password Redaction",
			input:    "password=super-secret-123",
			expected: "[REDACTED]",
		},
		{
			name:     "Token Redaction",
			input:    "token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "[REDACTED]",
		},
		{
			name:     "Bearer Token",
			input:    "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expected: "[REDACTED_JWT]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactSecrets(tt.input)
			if result == tt.input {
				t.Errorf("RedactSecrets() failed to redact secrets in: %s", tt.input)
			}
			if len(result) > 0 && result != tt.input {
				t.Logf("✓ %s: %s -> %s", tt.name, tt.input, result)
			}
		})
	}
}

func TestValidateAppName(t *testing.T) {
	tests := []struct {
		name    string
		appName string
		valid   bool
	}{
		{"valid-name", "my-app", true},
		{"valid-name-with-numbers", "app-123", true},
		{"invalid-semicolon", "app;test", false},
		{"invalid-pipe", "app|test", false},
		{"invalid-ampersand", "app&test", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAppName(tt.appName)
			if (err == nil) != tt.valid {
				if tt.valid {
					t.Errorf("ValidateAppName(%s) should be valid", tt.appName)
				} else {
					t.Errorf("ValidateAppName(%s) should be invalid", tt.appName)
				}
			} else if !tt.valid {
				t.Logf("✓ Correctly rejected invalid app name: %s", tt.appName)
			} else {
				t.Logf("✓ Correctly validated app name: %s", tt.appName)
			}
		})
	}
}

func TestValidateNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		valid     bool
	}{
		{"valid-ns", "default", true},
		{"valid-with-hyphen", "kube-system", true},
		{"invalid-underscore", "invalid_ns", false},
		{"invalid-uppercase", "MyNamespace", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNamespace(tt.namespace)
			if (err == nil) != tt.valid {
				if tt.valid {
					t.Errorf("ValidateNamespace(%s) should be valid", tt.namespace)
				} else {
					t.Errorf("ValidateNamespace(%s) should be invalid", tt.namespace)
				}
			} else if !tt.valid {
				t.Logf("✓ Correctly rejected invalid namespace: %s", tt.namespace)
			} else {
				t.Logf("✓ Correctly validated namespace: %s", tt.namespace)
			}
		})
	}
}

func TestOutputFilter(t *testing.T) {
	filter := NewOutputFilter(true, false)

	tests := []struct {
		name     string
		input    string
		filtered bool
	}{
		{
			"API Key in output",
			"connection using api_key=secret123",
			true,
		},
		{
			"Password in output",
			"logged in with password=mypassword",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.FilterOutput(tt.input)
			if tt.filtered && result == tt.input {
				t.Errorf("FilterOutput() should have filtered %s", tt.input)
			} else if tt.filtered {
				t.Logf("✓ Successfully filtered output: %s -> %s", tt.input, result)
			}
		})
	}
}
