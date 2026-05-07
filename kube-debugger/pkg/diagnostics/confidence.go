package diagnostics

import "strings"

// ConfidenceLevel is the normalized confidence bucket shown in CLI/report output.
type ConfidenceLevel string

const (
	ConfidenceHigh   ConfidenceLevel = "High"
	ConfidenceMedium ConfidenceLevel = "Medium"
	ConfidenceLow    ConfidenceLevel = "Low"
)

// RemediationSuggestion is a remediation action with confidence metadata.
type RemediationSuggestion struct {
	Text       string          `json:"text"`
	Confidence ConfidenceLevel `json:"confidence"`
	Rationale  string          `json:"rationale"`
	Source     string          `json:"source,omitempty"`
}

// AIHint contains the AI/pattern diagnosis message plus confidence metadata.
type AIHint struct {
	Message    string          `json:"message"`
	Confidence ConfidenceLevel `json:"confidence"`
	Rationale  string          `json:"rationale"`
	Source     string          `json:"source"`
}

// NormalizeLLMConfidence maps free-form LLM output to deterministic confidence buckets.
func NormalizeLLMConfidence(text string) (ConfidenceLevel, string) {
	t := strings.ToLower(text)

	score := 50
	reasons := make([]string, 0, 3)

	if strings.Contains(t, "confidence: high") || strings.Contains(t, "high confidence") {
		score += 35
		reasons = append(reasons, "LLM reported high confidence")
	}
	if strings.Contains(t, "confidence: medium") || strings.Contains(t, "medium confidence") {
		reasons = append(reasons, "LLM reported medium confidence")
	}
	if strings.Contains(t, "confidence: low") || strings.Contains(t, "low confidence") {
		score -= 35
		reasons = append(reasons, "LLM reported low confidence")
	}

	if strings.Contains(t, "root cause") || strings.Contains(t, "confirmed") || strings.Contains(t, "clearly") || strings.Contains(t, "definitive") {
		score += 15
		reasons = append(reasons, "response used strong evidence language")
	}

	if strings.Contains(t, "maybe") || strings.Contains(t, "might") || strings.Contains(t, "possibly") || strings.Contains(t, "uncertain") {
		score -= 20
		reasons = append(reasons, "response used uncertain language")
	}

	if strings.Contains(t, "could be") || strings.Contains(t, "one possibility") || strings.Contains(t, "several possible causes") {
		score -= 10
		reasons = append(reasons, "response presented multiple possible causes")
	}

	if score >= 70 {
		return ConfidenceHigh, joinReasons(reasons, "Strong evidence and low ambiguity in LLM output")
	}
	if score >= 40 {
		return ConfidenceMedium, joinReasons(reasons, "Partially supported diagnosis with moderate uncertainty")
	}
	return ConfidenceLow, joinReasons(reasons, "Ambiguous or hedged diagnosis from LLM output")
}

func patternConfidenceForStatus(status, lastError string) (ConfidenceLevel, string) {
	s := strings.ToLower(strings.TrimSpace(status))
	e := strings.ToLower(lastError)

	switch {
	case s == "crashloopbackoff":
		if strings.Contains(e, "oom") || strings.Contains(e, "memory") || strings.Contains(e, "connection refused") || strings.Contains(e, "dial") {
			return ConfidenceHigh, "CrashLoopBackOff status with specific last-error signal provides strong evidence"
		}
		return ConfidenceMedium, "CrashLoopBackOff status is reliable but root cause still needs log confirmation"
	case s == "oomkilled":
		return ConfidenceHigh, "OOMKilled is an explicit Kubernetes termination reason"
	case s == "imagepullbackoff" || s == "errimagepull":
		return ConfidenceHigh, "Image pull backoff states explicitly identify registry/image access failures"
	case s == "pending":
		return ConfidenceMedium, "Pending state indicates scheduling/storage constraints but exact blocker varies"
	case s == "evicted":
		return ConfidenceHigh, "Evicted state is explicit and usually tied to node pressure"
	case s == "terminating":
		return ConfidenceMedium, "Terminating state is clear but stuck termination root cause varies"
	case strings.Contains(s, "containercreating"):
		return ConfidenceMedium, "ContainerCreating suggests startup dependency issues but needs event correlation"
	case s == "runcontainererror":
		return ConfidenceHigh, "RunContainerError directly indicates container startup/runtime failure"
	case strings.Contains(s, "probe") || strings.Contains(e, "liveness") || strings.Contains(e, "readiness"):
		return ConfidenceMedium, "Probe failure signal is useful but can be secondary to deeper app problems"
	default:
		return ConfidenceLow, "Generic fallback due to weak or non-specific failure signals"
	}
}

func joinReasons(reasons []string, fallback string) string {
	if len(reasons) == 0 {
		return fallback
	}
	return strings.Join(reasons, "; ")
}
