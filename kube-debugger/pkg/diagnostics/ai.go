package diagnostics

import (
	"strings"
)

// AnalyzeLogsAI provides AI-like hints based on log content
func AnalyzeLogsAI(logs string) string {
	logs = strings.ToLower(logs)
	if strings.Contains(logs, "connection refused") {
		return "AI Hint: This looks like a DB connection issue. Check your database service and network policies."
	}
	if strings.Contains(logs, "oomkilled") {
		return "AI Hint: Pod was killed due to out-of-memory. Consider increasing memory limits."
	}
	if strings.Contains(logs, "imagepullbackoff") {
		return "AI Hint: Image pull failed. Check image name, tag, and registry credentials."
	}
	return "AI Hint: No obvious issues detected. Check full logs for more details."
}
