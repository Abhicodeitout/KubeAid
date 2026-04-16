package diagnostics

import (
	"fmt"
	"strings"
)

// AnalyzeWithContext is the main entry point for AI analysis.
// It tries a real LLM first (if KUBEAID_AI_PROVIDER is set), then falls back
// to fast local pattern matching.
func AnalyzeWithContext(appName, namespace, podName, status string, restarts int32, logs, events string) string {
	cfg := ResolveLLMConfig()
	if cfg.Provider != "" {
		prompt := BuildAnalysisPrompt(appName, namespace, podName, status, restarts, logs, events)
		result, err := CallLLM(cfg, prompt)
		if err == nil && result != "" {
			return fmt.Sprintf("🤖 AI (%s/%s): %s", cfg.Provider, cfg.Model, result)
		}
		// fall through to pattern-based on error
		_ = err
	}
	return AnalyzeLogsAI(logs)
}

// AnalyzeLogsAI provides fast pattern-based hints from log content (used as fallback).
func AnalyzeLogsAI(logs string) string {
	l := strings.ToLower(logs)

	switch {
	case strings.Contains(l, "connection refused") || strings.Contains(l, "dial tcp") || strings.Contains(l, "no such host"):
		return "AI Hint: Network connectivity issue detected. Check that dependent services are running, DNS resolves correctly, and network policies allow the connection."

	case strings.Contains(l, "oomkilled") || strings.Contains(l, "out of memory") || strings.Contains(l, "cannot allocate memory"):
		return "AI Hint: Pod was killed due to out-of-memory. Increase resources.limits.memory or investigate memory leaks in the application."

	case strings.Contains(l, "imagepullbackoff") || strings.Contains(l, "errimagepull") || strings.Contains(l, "pull access denied"):
		return "AI Hint: Image pull failed. Verify the image name and tag, ensure the imagePullSecret is present in the namespace, and confirm registry access."

	case strings.Contains(l, "crashloopbackoff"):
		return "AI Hint: Pod is crash-looping. Review previous container logs (--previous), check liveness probe settings, and ensure the container entrypoint exits cleanly."

	case strings.Contains(l, "readinessprobe") || strings.Contains(l, "livenessprobe") || strings.Contains(l, "probe failed"):
		return "AI Hint: A health probe is failing. Check the probe endpoint, port, initial delay, and that the application starts within the configured timeout."

	case strings.Contains(l, "permission denied") || strings.Contains(l, "operation not permitted") || strings.Contains(l, "forbidden"):
		return "AI Hint: Permission issue detected. Verify RBAC roles/bindings, pod security context, and filesystem mount permissions."

	case strings.Contains(l, "configmap") || strings.Contains(l, "secret not found") || strings.Contains(l, "no such file"):
		return "AI Hint: A required ConfigMap or Secret may be missing. Check that all referenced volumes and environment sources exist in the namespace."

	case strings.Contains(l, "timeout") || strings.Contains(l, "context deadline exceeded"):
		return "AI Hint: Request or operation timed out. Check network latency, upstream service responsiveness, and increase timeout settings if appropriate."

	case strings.Contains(l, "evicted"):
		return "AI Hint: Pod was evicted, likely due to node pressure (disk, memory, or PID). Review node conditions and set proper resource requests."

	case strings.Contains(l, "panic") || strings.Contains(l, "fatal error") || strings.Contains(l, "segfault"):
		return "AI Hint: Application crash detected. Review the stack trace in the logs and fix the underlying code or configuration error."

	case strings.Contains(l, "certificate") || strings.Contains(l, "tls") || strings.Contains(l, "x509"):
		return "AI Hint: TLS/certificate error detected. Verify certificate validity, CA bundles, and that secrets holding certs are up to date."

	default:
		return "AI Hint: No obvious pattern detected. Review full logs and pod events for more details."
	}
}
