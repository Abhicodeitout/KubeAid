package diagnostics

import (
	"strings"
	"testing"
)

func TestNormalizeLLMConfidenceBuckets(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want ConfidenceLevel
	}{
		{
			name: "high confidence explicit",
			in:   "Root cause confirmed. Confidence: high. Clearly caused by OOMKilled.",
			want: ConfidenceHigh,
		},
		{
			name: "medium confidence mixed language",
			in:   "Likely a probe timing issue. Could be related to startup latency.",
			want: ConfidenceMedium,
		},
		{
			name: "low confidence uncertain language",
			in:   "Maybe network. Possibly DNS or service discovery. One possibility is policy.",
			want: ConfidenceLow,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, rationale := NormalizeLLMConfidence(tc.in)
			if got != tc.want {
				t.Fatalf("NormalizeLLMConfidence(%q)=%s, want %s", tc.in, got, tc.want)
			}
			if strings.TrimSpace(rationale) == "" {
				t.Fatalf("expected non-empty rationale for %s", tc.name)
			}
		})
	}
}

func TestSuggestFixWithConfidenceDeterministicRules(t *testing.T) {
	crash := SuggestFixWithConfidenceForPod("CrashLoopBackOff", "connection refused", "pod-a", "default")
	if len(crash) == 0 {
		t.Fatal("expected crashloop suggestions")
	}
	for _, s := range crash {
		if s.Confidence != ConfidenceHigh {
			t.Fatalf("expected High confidence for crashloop+connection signals, got %s", s.Confidence)
		}
		if strings.TrimSpace(s.Rationale) == "" {
			t.Fatal("expected rationale for crashloop suggestion")
		}
	}

	fallback := SuggestFixWithConfidenceForPod("UnknownStatus", "", "pod-a", "default")
	if len(fallback) == 0 {
		t.Fatal("expected fallback suggestions")
	}
	for _, s := range fallback {
		if s.Confidence != ConfidenceLow {
			t.Fatalf("expected Low confidence for default fallback, got %s", s.Confidence)
		}
	}
}

func TestCopilotFormatIncludesConfidence(t *testing.T) {
	s := &CopilotSuggestion{
		Title:       "Sample",
		Severity:    "warning",
		Confidence:  ConfidenceMedium,
		Rationale:   "Probe failures can be secondary symptoms",
		Description: "Check readiness probe",
	}
	out := s.Format()
	if !strings.Contains(out, "**Confidence:** Medium") {
		t.Fatalf("expected confidence line in formatted output, got: %s", out)
	}
	if !strings.Contains(out, "**Confidence Rationale:** Probe failures can be secondary symptoms") {
		t.Fatalf("expected confidence rationale in formatted output, got: %s", out)
	}
}
