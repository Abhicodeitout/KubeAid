package analyzer

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientkubernetes "k8s.io/client-go/kubernetes"
	"kube-debugger/pkg/kubernetes"
)

// TimelineEntry is a timestamped incident step.
type TimelineEntry struct {
	Time       time.Time `json:"time"`
	Kind       string    `json:"kind"`
	Severity   string    `json:"severity"`
	Summary    string    `json:"summary"`
	Details    string    `json:"details"`
	Source     string    `json:"source"`
	Role       string    `json:"role"`
	PodName    string    `json:"pod_name"`
	Namespace  string    `json:"namespace"`
}

// TimelineReport holds incident timeline output for an app.
type TimelineReport struct {
	AppName      string          `json:"app_name"`
	Namespace    string          `json:"namespace"`
	PodName      string          `json:"pod_name"`
	GeneratedAt  time.Time       `json:"generated_at"`
	HealthScore  int             `json:"health_score"`
	FirstCause   *TimelineEntry  `json:"first_cause,omitempty"`
	Timeline     []TimelineEntry `json:"timeline"`
}

type timelineEvent struct {
	Time     time.Time
	Kind     string
	Severity string
	Reason   string
	Summary  string
	Details  string
	Source   string
}

// AnalyzeTimeline reconstructs an ordered incident timeline for the primary pod.
func AnalyzeTimeline(appName, namespace string) (*TimelineReport, error) {
	if namespace == "" {
		namespace = strings.TrimSpace(os.Getenv("KUBE_NAMESPACE"))
	}
	if namespace == "" {
		namespace = "default"
	}

	baseReport, err := AnalyzeAppReport(appName, namespace)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.GetKubeClient()
	if err != nil {
		return nil, fmt.Errorf("error connecting to cluster: %w", err)
	}

	pod, err := findPodByName(clientset, baseReport.Namespace, baseReport.PodName)
	if err != nil {
		return nil, err
	}

	podEvents, err := kubernetes.ListPodEvents(clientset, baseReport.Namespace, pod.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to read pod events: %w", err)
	}

	logLines, err := kubernetes.GetTimestampedPodLogs(clientset, baseReport.Namespace, pod.Name, 50)
	if err != nil {
		logLines = nil
	}

	entries := buildTimelineEntries(baseReport.Namespace, pod.Name, pod, podEvents, logLines, baseReport.GeneratedAt)
	firstCause := firstCauseEntry(entries)

	return &TimelineReport{
		AppName:     appName,
		Namespace:   baseReport.Namespace,
		PodName:     pod.Name,
		GeneratedAt: baseReport.GeneratedAt,
		HealthScore: baseReport.HealthScore,
		FirstCause:  firstCause,
		Timeline:    entries,
	}, nil
}

func findPodByName(clientset *clientkubernetes.Clientset, namespace, podName string) (*corev1.Pod, error) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod %q in namespace %q: %w", podName, namespace, err)
	}
	return pod, nil
}

func buildTimelineEntries(namespace, podName string, pod *corev1.Pod, podEvents []kubernetes.PodEvent, logLines []kubernetes.LogLine, generatedAt time.Time) []TimelineEntry {
	raw := make([]timelineEvent, 0, len(podEvents)+len(logLines)+2)

	for _, event := range podEvents {
		severity := eventSeverity(event.Type, event.Reason, event.Message)
		kind := "event"
		if strings.EqualFold(event.Reason, "Unhealthy") {
			kind = "probe"
		}
		raw = append(raw, timelineEvent{
			Time:     event.Time,
			Kind:     kind,
			Severity: severity,
			Reason:   event.Reason,
			Summary:  event.Reason,
			Details:  event.Message,
			Source:   "k8s-event",
		})
	}

	for _, logLine := range logLines {
		if !isImportantLogLine(logLine.Message) {
			continue
		}
		raw = append(raw, timelineEvent{
			Time:     logLine.Time,
			Kind:     "log",
			Severity: logSeverity(logLine.Message),
			Reason:   "ApplicationLog",
			Summary:  trimSummary(logLine.Message),
			Details:  logLine.Message,
			Source:   "pod-log",
		})
	}

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.LastTerminationState.Terminated == nil {
			continue
		}
		terminated := cs.LastTerminationState.Terminated
		when := terminated.FinishedAt.Time
		if when.IsZero() {
			when = generatedAt
		}
		summary := terminated.Reason
		if summary == "" {
			summary = fmt.Sprintf("Container %s restarted", cs.Name)
		}
		details := terminated.Message
		if details == "" {
			details = fmt.Sprintf("Exit code %d after %s", terminated.ExitCode, cs.Name)
		}
		raw = append(raw, timelineEvent{
			Time:     when,
			Kind:     "restart",
			Severity: terminationSeverity(terminated.Reason, terminated.ExitCode, details),
			Reason:   terminated.Reason,
			Summary:  summary,
			Details:  details,
			Source:   "container-status",
		})
	}

	sort.SliceStable(raw, func(i, j int) bool {
		if raw[i].Time.Equal(raw[j].Time) {
			return timelineRank(raw[i]) < timelineRank(raw[j])
		}
		return raw[i].Time.Before(raw[j].Time)
	})

	firstCauseIdx := firstCauseIndex(raw)
	entries := make([]TimelineEntry, 0, len(raw))
	for idx, event := range raw {
		role := "context"
		if idx == firstCauseIdx {
			role = "first-cause"
		} else if firstCauseIdx >= 0 && idx > firstCauseIdx && event.Severity != "info" {
			role = "impact"
		}
		entries = append(entries, TimelineEntry{
			Time:      event.Time,
			Kind:      event.Kind,
			Severity:  event.Severity,
			Summary:   event.Summary,
			Details:   event.Details,
			Source:    event.Source,
			Role:      role,
			PodName:   podName,
			Namespace: namespace,
		})
	}

	return entries
}

func firstCauseEntry(entries []TimelineEntry) *TimelineEntry {
	for i := range entries {
		if entries[i].Role == "first-cause" {
			entry := entries[i]
			return &entry
		}
	}
	return nil
}

func firstCauseIndex(events []timelineEvent) int {
	for idx, event := range events {
		if isCauseCandidate(event) {
			return idx
		}
	}
	return -1
}

func isCauseCandidate(event timelineEvent) bool {
	if event.Severity == "critical" {
		return true
	}
	reason := strings.ToLower(event.Reason)
	details := strings.ToLower(event.Details)
	for _, token := range []string{"backoff", "failed", "oom", "imagepull", "unhealthy", "panic", "fatal", "error", "killing"} {
		if strings.Contains(reason, token) || strings.Contains(details, token) {
			return true
		}
	}
	return false
}

func timelineRank(event timelineEvent) int {
	switch event.Kind {
	case "event", "probe":
		return 0
	case "restart":
		return 1
	case "log":
		return 2
	default:
		return 3
	}
}

func eventSeverity(eventType, reason, message string) string {
	if strings.EqualFold(eventType, "Warning") {
		return "critical"
	}
	text := strings.ToLower(reason + " " + message)
	for _, token := range []string{"failed", "backoff", "oom", "error", "kill", "unhealthy"} {
		if strings.Contains(text, token) {
			return "critical"
		}
	}
	if strings.Contains(text, "pull") || strings.Contains(text, "probe") {
		return "warning"
	}
	return "info"
}

func logSeverity(message string) string {
	text := strings.ToLower(message)
	for _, token := range []string{"panic", "fatal", "segfault", "out of memory"} {
		if strings.Contains(text, token) {
			return "critical"
		}
	}
	for _, token := range []string{"error", "fail", "refused", "timeout", "warn"} {
		if strings.Contains(text, token) {
			return "warning"
		}
	}
	return "info"
}

func terminationSeverity(reason string, exitCode int32, details string) string {
	text := strings.ToLower(reason + " " + details)
	if exitCode != 0 || strings.Contains(text, "oom") || strings.Contains(text, "error") {
		return "critical"
	}
	return "warning"
}

func isImportantLogLine(message string) bool {
	return logSeverity(message) != "info"
}

func trimSummary(message string) string {
	trimmed := strings.TrimSpace(message)
	if len(trimmed) <= 72 {
		return trimmed
	}
	return trimmed[:69] + "..."
}

// RenderTimeline renders a timeline report for terminal output.
func RenderTimeline(report *TimelineReport) string {
	var b strings.Builder

	b.WriteString(styleBorder.Render(styleTitle.Render(fmt.Sprintf("  Incident Timeline  ·  %s  ·  ns: %s", report.AppName, report.Namespace))) + "\n")
	b.WriteString(divider("TIMELINE SUMMARY"))
	b.WriteString(kv("Primary Pod:", report.PodName) + "\n")
	b.WriteString(kv("Health Score:", fmt.Sprintf("%d/100", report.HealthScore)) + "\n")
	b.WriteString(kv("Entries:", fmt.Sprintf("%d", len(report.Timeline))) + "\n")

	if report.FirstCause != nil {
		b.WriteString("\n")
		b.WriteString(styleRed.Render("  First-cause candidate: ") + styleValue.Render(report.FirstCause.Summary) + "\n")
		if report.FirstCause.Details != "" {
			b.WriteString(styleDim.Render("  " + report.FirstCause.Details) + "\n")
		}
	}

	b.WriteString(divider("ORDERED INCIDENT FLOW"))
	if len(report.Timeline) == 0 {
		b.WriteString(styleDim.Render("  No timestamped incident signals were available.") + "\n")
		return b.String()
	}

	for _, entry := range report.Timeline {
		marker := "·"
		lineStyle := styleDim
		switch entry.Role {
		case "first-cause":
			marker = "CAUSE"
			lineStyle = styleRed
		case "impact":
			marker = "IMPACT"
			lineStyle = styleYellow
		default:
			switch entry.Severity {
			case "critical":
				lineStyle = styleRed
			case "warning":
				lineStyle = styleYellow
			}
		}

		timestamp := entry.Time.UTC().Format("2006-01-02 15:04:05")
		b.WriteString(lineStyle.Render(fmt.Sprintf("  [%s] %-6s %-7s %s", timestamp, strings.ToUpper(entry.Kind), marker, entry.Summary)) + "\n")
		if entry.Details != "" && entry.Details != entry.Summary {
			b.WriteString(styleDim.Render("      " + entry.Details) + "\n")
		}
	}

	b.WriteString("\n" + styleDim.Render(fmt.Sprintf("  Generated at %s", report.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))) + "\n")
	return b.String()
}