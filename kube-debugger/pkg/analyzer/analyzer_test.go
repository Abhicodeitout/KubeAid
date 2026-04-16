package analyzer

import "testing"

func TestAnalyzeApp(t *testing.T) {
	result := AnalyzeApp("test-app", "default")
	if result == "" {
		t.Error("Expected non-empty result")
	}
}
