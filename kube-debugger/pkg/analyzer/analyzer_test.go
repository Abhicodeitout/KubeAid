//go:build integration

package analyzer

import "testing"

// TestAnalyzeApp requires a live Kubernetes cluster.
// Run with: go test -tags=integration ./pkg/analyzer/...
func TestAnalyzeApp(t *testing.T) {
	result := AnalyzeApp("test-app", "default")
	if result == "" {
		t.Error("Expected non-empty result")
	}
}
