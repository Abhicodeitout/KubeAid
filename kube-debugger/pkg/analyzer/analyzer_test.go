package analyzer

import "testing"

func TestAnalyzeApp(t *testing.T) {
	result := AnalyzeApp("test-app")
	if result == "" {
		t.Error("Expected non-empty result")
	}
}
