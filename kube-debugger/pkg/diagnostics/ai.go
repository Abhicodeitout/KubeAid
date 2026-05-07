package diagnostics

import (
	"fmt"
	"strings"
)

// AnalyzeWithContext is the main entry point for AI analysis.
// It tries a real LLM first (if KUBEAID_AI_PROVIDER is set), then falls back
// to fast local pattern matching.
func AnalyzeWithContext(appName, namespace, podName, status string, restarts int32, logs, events string) string {
	hint := AnalyzeWithContextDetailed(appName, namespace, podName, status, restarts, logs, events)
	return hint.Message
}

// AnalyzeWithContextDetailed returns the diagnosis with normalized confidence metadata.
func AnalyzeWithContextDetailed(appName, namespace, podName, status string, restarts int32, logs, events string) AIHint {
	cfg := ResolveLLMConfig()
	if cfg.Provider != "" {
		prompt := BuildAnalysisPrompt(appName, namespace, podName, status, restarts, logs, events)
		result, err := CallLLM(cfg, prompt)
		if err == nil && result != "" {
			confidence, rationale := NormalizeLLMConfidence(result)
			return AIHint{
				Message:    fmt.Sprintf("🤖 AI (%s/%s): %s", cfg.Provider, cfg.Model, result),
				Confidence: confidence,
				Rationale:  rationale,
				Source:     "llm",
			}
		}
		// fall through to pattern-based on error
		_ = err
	}
	return AnalyzeSignalsAIWithConfidence(status, logs, events)
}

// AnalyzeLogsAI provides fast pattern-based hints from log content (used as fallback).
func AnalyzeLogsAI(logs string) string {
	return AnalyzeLogsAIWithConfidence(logs).Message
}

// AnalyzeLogsAIWithConfidence is deterministic pattern matching with confidence metadata.
func AnalyzeLogsAIWithConfidence(logs string) AIHint {
	return AnalyzeSignalsAIWithConfidence("", logs, "")
}

// AnalyzeSignalsAIWithConfidence scores deterministic fallback confidence using status, events, and logs.
// This avoids mismatches where logs are sparse but pod status/events are explicit (for example ImagePullBackOff).
func AnalyzeSignalsAIWithConfidence(status, logs, events string) AIHint {
	s := strings.ToLower(status)
	l := strings.ToLower(logs)
	e := strings.ToLower(events)
	combined := s + "\n" + e + "\n" + l

	switch {
	case s == "imagepullbackoff" || s == "errimagepull" || strings.Contains(combined, "imagepullbackoff") || strings.Contains(combined, "errimagepull") || strings.Contains(combined, "pull access denied"):
		return AIHint{Message: "AI Hint: Image pull failed. Verify the image name and tag, ensure the imagePullSecret is present in the namespace, and confirm registry access.", Confidence: ConfidenceHigh, Rationale: "Image pull failure state/event is explicit and deterministic", Source: "pattern"}

	case s == "oomkilled" || strings.Contains(combined, "oomkilled") || strings.Contains(combined, "out of memory") || strings.Contains(combined, "cannot allocate memory"):
		return AIHint{Message: "AI Hint: Pod was killed due to out-of-memory. Increase resources.limits.memory or investigate memory leaks in the application.", Confidence: ConfidenceHigh, Rationale: "OOM signals from status/events/logs are explicit Kubernetes termination indicators", Source: "pattern"}

	case s == "evicted" || strings.Contains(combined, "evicted"):
		return AIHint{Message: "AI Hint: Pod was evicted, likely due to node pressure (disk, memory, or PID). Review node conditions and set proper resource requests.", Confidence: ConfidenceHigh, Rationale: "Eviction state/event is explicit scheduler/node pressure evidence", Source: "pattern"}

	case s == "crashloopbackoff" || strings.Contains(combined, "crashloopbackoff"):
		return AIHint{Message: "AI Hint: Pod is crash-looping. Review previous container logs (--previous), check liveness probe settings, and ensure the container entrypoint exits cleanly.", Confidence: ConfidenceMedium, Rationale: "CrashLoopBackOff is clear but root cause still requires further inspection", Source: "pattern"}

	case strings.Contains(combined, "connection refused") || strings.Contains(combined, "dial tcp") || strings.Contains(combined, "no such host"):
		return AIHint{Message: "AI Hint: Network connectivity issue detected. Check that dependent services are running, DNS resolves correctly, and network policies allow the connection.", Confidence: ConfidenceHigh, Rationale: "Direct network error signatures were found in logs", Source: "pattern"}

	case strings.Contains(combined, "readinessprobe") || strings.Contains(combined, "livenessprobe") || strings.Contains(combined, "probe failed"):
		return AIHint{Message: "AI Hint: A health probe is failing. Check the probe endpoint, port, initial delay, and that the application starts within the configured timeout.", Confidence: ConfidenceMedium, Rationale: "Probe errors are useful indicators but often secondary symptoms", Source: "pattern"}

	case strings.Contains(combined, "permission denied") || strings.Contains(combined, "operation not permitted") || strings.Contains(combined, "forbidden"):
		return AIHint{Message: "AI Hint: Permission issue detected. Verify RBAC roles/bindings, pod security context, and filesystem mount permissions.", Confidence: ConfidenceMedium, Rationale: "Permission signals are strong but can map to multiple policy layers", Source: "pattern"}

	case strings.Contains(combined, "configmap") || strings.Contains(combined, "secret not found") || strings.Contains(combined, "no such file"):
		return AIHint{Message: "AI Hint: A required ConfigMap or Secret may be missing. Check that all referenced volumes and environment sources exist in the namespace.", Confidence: ConfidenceMedium, Rationale: "Missing config signals are specific but may need manifest/event confirmation", Source: "pattern"}

	case strings.Contains(combined, "timeout") || strings.Contains(combined, "context deadline exceeded"):
		return AIHint{Message: "AI Hint: Request or operation timed out. Check network latency, upstream service responsiveness, and increase timeout settings if appropriate.", Confidence: ConfidenceMedium, Rationale: "Timeouts indicate degradation but not a single deterministic cause", Source: "pattern"}

	case strings.Contains(combined, "panic") || strings.Contains(combined, "fatal error") || strings.Contains(combined, "segfault"):
		return AIHint{Message: "AI Hint: Application crash detected. Review the stack trace in the logs and fix the underlying code or configuration error.", Confidence: ConfidenceHigh, Rationale: "Crash signatures provide direct evidence of application failure", Source: "pattern"}

	case strings.Contains(combined, "certificate") || strings.Contains(combined, "tls") || strings.Contains(combined, "x509"):
		return AIHint{Message: "AI Hint: TLS/certificate error detected. Verify certificate validity, CA bundles, and that secrets holding certs are up to date.", Confidence: ConfidenceHigh, Rationale: "TLS/x509 error text is highly specific to certificate configuration", Source: "pattern"}

	default:
		return AIHint{Message: "AI Hint: No obvious pattern detected. Review full logs and pod events for more details.", Confidence: ConfidenceLow, Rationale: "No deterministic signature matched in status/events/logs", Source: "pattern"}
	}
}
