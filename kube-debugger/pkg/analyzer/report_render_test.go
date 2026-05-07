package analyzer

import (
	"strings"
	"testing"
	"time"

	"kube-debugger/pkg/diagnostics"
)

func TestRenderReportDisplaysConfidence(t *testing.T) {
	r := &Report{
		AppName:      "demo-app",
		Namespace:    "default",
		PodCount:     1,
		PodName:      "demo-app-123",
		Status:       "CrashLoopBackOff",
		Ready:        "0/1",
		RestartCount: 7,
		Age:          "5m",
		HealthScore:  20,
		AIHint:       "AI Hint: Pod is crash-looping",
		AIConfidence: "Medium",
		AIRationale:  "CrashLoopBackOff is clear but root cause still needs logs",
		Suggestions:  []string{"kubectl describe pod demo-app-123 -n default"},
		SuggestionDetails: []diagnostics.RemediationSuggestion{
			{
				Text:       "kubectl describe pod demo-app-123 -n default",
				Confidence: diagnostics.ConfidenceHigh,
				Rationale:  "CrashLoopBackOff with connection errors is a strong signal",
				Source:     "pattern",
			},
		},
		Events:      "Warning BackOff",
		Resources:   "cpu: 10m",
		Logs:        "error connection refused",
		GeneratedAt: time.Now().UTC(),
	}

	out := RenderReport(r)
	if !strings.Contains(out, "Confidence:") || !strings.Contains(out, "Medium") {
		t.Fatalf("expected AI confidence in render output, got: %s", out)
	}
	if !strings.Contains(out, "[High] kubectl describe pod demo-app-123 -n default") {
		t.Fatalf("expected suggestion confidence label in render output, got: %s", out)
	}
	if !strings.Contains(out, "rationale:") {
		t.Fatalf("expected suggestion rationale in render output, got: %s", out)
	}
}
