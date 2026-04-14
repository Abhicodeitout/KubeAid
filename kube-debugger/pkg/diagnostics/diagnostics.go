package diagnostics

func SuggestFix(status, lastError string) []string {
	// Simple rules for demo; expand for real use
	if status == "CrashLoopBackOff" && lastError != "" {
		return []string{"Check logs for stack trace", "Verify DB service", "Check env variables"}
	}
	if status == "ImagePullBackOff" {
		return []string{"Check image name", "Check image pull secret"}
	}
	if status == "Pending" {
		return []string{"Check node resources", "Check pod scheduling events"}
	}
	return []string{"Check pod events", "Review resource limits"}
}
