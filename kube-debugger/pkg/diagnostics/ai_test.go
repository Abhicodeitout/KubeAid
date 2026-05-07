package diagnostics

import (
	"strings"
	"testing"
)

func TestAnalyzeLogsAI(t *testing.T) {
	cases := []struct {
		log    string
		expect string
	}{
		{"connection refused on port 5432", "Network connectivity issue"},
		{"OOMKilled", "out-of-memory"},
		{"ImagePullBackOff", "Image pull failed"},
		{"all good", "No obvious pattern detected"},
	}
	for _, c := range cases {
		result := AnalyzeLogsAI(c.log)
		if c.expect != "" && !strings.Contains(result, c.expect) {
			t.Errorf("Expected '%s' in result for log '%s', got '%s'", c.expect, c.log, result)
		}
	}
}

func TestAnalyzeWithContextDetailedUsesStatusAndEventsSignals(t *testing.T) {
	hint := AnalyzeWithContextDetailed(
		"demo",
		"default",
		"pod-a",
		"ImagePullBackOff",
		0,
		"",
		"Failed: Error: ImagePullBackOff",
	)

	if hint.Confidence != ConfidenceHigh {
		t.Fatalf("expected High confidence from ImagePullBackOff status/events, got %s", hint.Confidence)
	}
	if !strings.Contains(hint.Message, "Image pull failed") {
		t.Fatalf("expected image pull AI hint message, got: %s", hint.Message)
	}
}

func TestAnalyzeSignalsAIWithConfidenceDefaultLow(t *testing.T) {
	hint := AnalyzeSignalsAIWithConfidence("Running", "all good", "Normal: Started")
	if hint.Confidence != ConfidenceLow {
		t.Fatalf("expected Low confidence for non-matching signals, got %s", hint.Confidence)
	}
}
