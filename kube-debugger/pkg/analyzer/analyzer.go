package analyzer

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-debugger/pkg/diagnostics"
	"kube-debugger/pkg/kubernetes"
)

// PodSummary holds per-pod data within a multi-pod report.
type PodSummary struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	Ready        string `json:"ready"`
	RestartCount int32  `json:"restart_count"`
	Age          string `json:"age"`
}

// Report holds structured analysis data for a Kubernetes app.
type Report struct {
	AppName   string       `json:"app_name"`
	Namespace string       `json:"namespace"`
	PodCount  int          `json:"pod_count"`
	Pods      []PodSummary `json:"pods"`
	// Primary pod (most relevant / most troubled)
	PodName      string    `json:"pod_name"`
	Status       string    `json:"status"`
	Ready        string    `json:"ready"`
	RestartCount int32     `json:"restart_count"`
	Age          string    `json:"age"`
	HealthScore  int       `json:"health_score"`
	Logs         string    `json:"logs"`
	Events       string    `json:"events"`
	Resources    string    `json:"resources"`
	AIHint       string    `json:"ai_hint"`
	Suggestions  []string  `json:"suggestions"`
	CopilotFix   string    `json:"copilot_fix"`   // Copilot-style structured suggestion
	GeneratedAt  time.Time `json:"generated_at"`
}

// HealthScore computes a 0–100 score based on status, readiness, restarts, and events.
func computeHealthScore(status, ready string, restarts int32, events string) int {
	score := 100
	s := strings.ToLower(status)
	r := strings.TrimSpace(strings.ToLower(ready))
	if r != "1/1" {
		// A non-ready primary pod is a strong signal of degraded service.
		score -= 30
	}
	switch s {
	case "crashloopbackoff":
		score -= 60
	case "oomkilled":
		score -= 50
	case "imagepullbackoff", "errimagepull":
		score -= 50
	case "evicted":
		score -= 40
	case "terminating":
		score -= 30
	case "containercreating":
		score -= 20
	case "pending":
		score -= 20
	case "running":
		// healthy base
	}
	// penalise restarts
	if restarts >= 10 {
		score -= 30
	} else if restarts >= 5 {
		score -= 20
	} else if restarts >= 1 {
		score -= 10
	}
	// penalise warning/failed events
	for _, line := range strings.Split(strings.ToLower(events), "\n") {
		if strings.Contains(line, "warning") || strings.Contains(line, "failed") || strings.Contains(line, "kill") {
			score -= 5
		}
	}
	if score < 0 {
		score = 0
	}
	return score
}

func podAge(creationTime metav1.Time) string {
	d := time.Since(creationTime.Time)
	if d.Hours() >= 24 {
		return fmt.Sprintf("%.0fd", d.Hours()/24)
	}
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.0fh", d.Hours())
	}
	return fmt.Sprintf("%.0fm", d.Minutes())
}

func podTroubleScore(status, ready string, restarts int32) int {
	s := strings.ToLower(strings.TrimSpace(status))
	score := 0

	switch s {
	case "crashloopbackoff":
		score += 120
	case "oomkilled", "imagepullbackoff", "errimagepull", "runcontainererror":
		score += 100
	case "evicted", "failed":
		score += 80
	case "pending", "containercreating", "terminating":
		score += 50
	case "running":
		// healthy baseline
	default:
		score += 60
	}

	if strings.TrimSpace(strings.ToLower(ready)) != "1/1" {
		score += 40
	}

	restartPenalty := int(restarts) * 3
	if restartPenalty > 60 {
		restartPenalty = 60
	}
	return score + restartPenalty
}

// AnalyzeAppReport performs analysis and returns a structured Report.
func AnalyzeAppReport(appName, namespace string) (*Report, error) {
	if namespace == "" {
		namespace = os.Getenv("KUBE_NAMESPACE")
	}
	if namespace == "" {
		namespace = "default"
	}
	clientset, err := kubernetes.GetKubeClient()
	if err != nil {
		return nil, fmt.Errorf("error connecting to cluster: %w", err)
	}
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{
		LabelSelector: fmt.Sprintf("app=%s", appName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query pods for app '%s' in namespace '%s': %w", appName, namespace, err)
	}
	if len(podList.Items) == 0 {
		return nil, fmt.Errorf("no pods found for app '%s' in namespace '%s'", appName, namespace)
	}

	// Build pod summaries and pick the "primary" pod — prefer most troubled one
	var summaries []PodSummary
	primaryIdx := 0
	bestScore := -1
	for i, p := range podList.Items {
		st := string(p.Status.Phase)
		if len(p.Status.ContainerStatuses) > 0 && p.Status.ContainerStatuses[0].State.Waiting != nil {
			st = p.Status.ContainerStatuses[0].State.Waiting.Reason
		}
		ready := "0/1"
		var rc int32
		if len(p.Status.ContainerStatuses) > 0 {
			rc = p.Status.ContainerStatuses[0].RestartCount
			if p.Status.ContainerStatuses[0].Ready {
				ready = "1/1"
			}
		}
		summaries = append(summaries, PodSummary{
			Name:         p.Name,
			Status:       st,
			Ready:        ready,
			RestartCount: rc,
			Age:          podAge(p.CreationTimestamp),
		})

		troubleScore := podTroubleScore(st, ready, rc)
		if troubleScore > bestScore {
			bestScore = troubleScore
			primaryIdx = i
		}
	}

	pod := podList.Items[primaryIdx]
	ps := summaries[primaryIdx]

	logs, _ := kubernetes.GetPodLogs(clientset, namespace, pod.Name)
	events, _ := kubernetes.GetPodEvents(clientset, namespace, pod.Name)
	resources, _ := kubernetes.GetPodResourceUsage(clientset, namespace, pod.Name)
	lastError := ""
	if len(pod.Status.ContainerStatuses) > 0 && pod.Status.ContainerStatuses[0].LastTerminationState.Terminated != nil {
		lastError = pod.Status.ContainerStatuses[0].LastTerminationState.Terminated.Message
	}

	return &Report{
		AppName:      appName,
		Namespace:    namespace,
		PodCount:     len(podList.Items),
		Pods:         summaries,
		PodName:      ps.Name,
		Status:       ps.Status,
		Ready:        ps.Ready,
		RestartCount: ps.RestartCount,
		Age:          ps.Age,
		HealthScore:  computeHealthScore(ps.Status, ps.Ready, ps.RestartCount, events),
		Logs:         logs,
		Events:       events,
		Resources:    resources,
		AIHint:       diagnostics.AnalyzeWithContext(appName, namespace, ps.Name, ps.Status, ps.RestartCount, logs, events),
		CopilotFix:   formatCopilotSuggestion(appName, namespace, ps.Name, ps.Status, ps.RestartCount, logs, events),
		Suggestions:  diagnostics.SuggestFixForPod(ps.Status, lastError, ps.Name, namespace),
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

// ─── lipgloss styles ─────────────────────────────────────────────────────────

var (
	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(0, 1)

	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("63")).
			MarginBottom(1)

	styleLabel = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("250"))
	styleValue = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	styleGreen  = lipgloss.NewStyle().Foreground(lipgloss.Color("82"))
	styleYellow = lipgloss.NewStyle().Foreground(lipgloss.Color("220"))
	styleRed    = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))

	styleDim      = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleSectionH = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	styleHint     = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Italic(true)
)

func healthColor(score int) lipgloss.Style {
	if score >= 80 {
		return styleGreen
	}
	if score >= 50 {
		return styleYellow
	}
	return styleRed
}

func statusColor(status string) string {
	s := strings.ToLower(status)
	switch s {
	case "running":
		return styleGreen.Render(status)
	case "pending", "containercreating":
		return styleYellow.Render(status)
	default:
		return styleRed.Render(status)
	}
}

func healthBar(score int) string {
	filled := score / 10
	bar := strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
	return healthColor(score).Render(bar) + fmt.Sprintf("  %d/100", score)
}

func divider(title string) string {
	return "\n" + styleSectionH.Render("── "+title+" ") + styleDim.Render(strings.Repeat("─", 50)) + "\n"
}

func kv(label, value string) string {
	return styleLabel.Render(fmt.Sprintf("  %-16s", label)) + styleValue.Render(value)
}

// AnalyzeApp performs analysis and returns a rich formatted string for the terminal.
func AnalyzeApp(appName, namespace string) string {
	r, err := AnalyzeAppReport(appName, namespace)
	if err != nil {
		return styleRed.Render("❌  "+err.Error()) + "\n"
	}
	return renderReport(r)
}

func renderReport(r *Report) string {
	var b strings.Builder

	// ── Header ───────────────────────────────────────────────────────────────
	header := styleTitle.Render(fmt.Sprintf("  KubeAid  ·  %s  ·  ns: %s", r.AppName, r.Namespace))
	b.WriteString(styleBorder.Render(header) + "\n")

	// ── Pod Overview ─────────────────────────────────────────────────────────
	b.WriteString(divider("POD OVERVIEW"))
	b.WriteString(kv("App:", r.AppName) + "\n")
	b.WriteString(kv("Namespace:", r.Namespace) + "\n")
	b.WriteString(kv("Total Pods:", fmt.Sprintf("%d", r.PodCount)) + "\n")

	if len(r.Pods) > 1 {
		b.WriteString("\n")
		b.WriteString(styleLabel.Render(fmt.Sprintf("  %-42s %-16s %-6s %-9s %s\n", "POD", "STATUS", "READY", "RESTARTS", "AGE")))
		for _, p := range r.Pods {
			b.WriteString(styleDim.Render(fmt.Sprintf("  %-42s ", p.Name)))
			b.WriteString(statusColor(p.Status) + "  ")
			b.WriteString(styleValue.Render(fmt.Sprintf("%-6s %-9d %s\n", p.Ready, p.RestartCount, p.Age)))
		}
	}

	// ── Primary Pod Detail ───────────────────────────────────────────────────
	b.WriteString(divider("PRIMARY POD"))
	b.WriteString(kv("Name:", r.PodName) + "\n")
	b.WriteString(kv("Status:", statusColor(r.Status)) + "\n")
	b.WriteString(kv("Ready:", r.Ready) + "\n")
	b.WriteString(kv("Restarts:", fmt.Sprintf("%d", r.RestartCount)) + "\n")
	b.WriteString(kv("Age:", r.Age) + "\n")

	// ── Health Score ─────────────────────────────────────────────────────────
	b.WriteString(divider("HEALTH SCORE"))
	b.WriteString("  " + healthBar(r.HealthScore) + "\n")

	// ── AI Hint ──────────────────────────────────────────────────────────────
	b.WriteString(divider("AI ANALYSIS"))
	b.WriteString("  " + styleHint.Render(r.AIHint) + "\n")

	// ── Copilot Fix ──────────────────────────────────────────────────────────
	if r.CopilotFix != "" {
		b.WriteString(divider("🤖 COPILOT FIX"))
		b.WriteString(r.CopilotFix + "\n")
	}

	// ── Suggestions ──────────────────────────────────────────────────────────
	b.WriteString(divider("SUGGESTIONS"))
	for i, s := range r.Suggestions {
		b.WriteString(styleYellow.Render(fmt.Sprintf("  %d. ", i+1)) + styleValue.Render(s) + "\n")
	}

	// ── Events ───────────────────────────────────────────────────────────────
	b.WriteString(divider("EVENTS"))
	for _, line := range strings.Split(strings.TrimSpace(r.Events), "\n") {
		if line == "" {
			continue
		}
		l := strings.ToLower(line)
		if strings.Contains(l, "warning") || strings.Contains(l, "failed") || strings.Contains(l, "kill") {
			b.WriteString(styleRed.Render("  ⚠  "+line) + "\n")
		} else {
			b.WriteString(styleDim.Render("  ·  "+line) + "\n")
		}
	}

	// ── Resource Usage ───────────────────────────────────────────────────────
	b.WriteString(divider("RESOURCE USAGE"))
	for _, line := range strings.Split(strings.TrimSpace(r.Resources), "\n") {
		b.WriteString("  " + styleValue.Render(line) + "\n")
	}

	// ── Last Logs ────────────────────────────────────────────────────────────
	b.WriteString(divider("LAST LOGS (tail 20)"))
	for _, line := range strings.Split(strings.TrimSpace(r.Logs), "\n") {
		if line == "" {
			continue
		}
		l := strings.ToLower(line)
		if strings.ContainsAny(l, "error") || strings.Contains(l, "fatal") || strings.Contains(l, "panic") {
			b.WriteString(styleRed.Render("  "+line) + "\n")
		} else if strings.Contains(l, "warn") {
			b.WriteString(styleYellow.Render("  "+line) + "\n")
		} else {
			b.WriteString(styleDim.Render("  "+line) + "\n")
		}
	}

	b.WriteString("\n" + styleDim.Render(fmt.Sprintf("  Generated at %s", r.GeneratedAt.Format("2006-01-02 15:04:05 UTC"))) + "\n")
	return b.String()
}

// RenderReport renders a Report to a terminal string (same as AnalyzeApp but accepts an existing Report).
func RenderReport(r *Report) string {
	return renderReport(r)
}

// AnalyzeAllNamespaces scans for the app in all namespaces and returns a combined formatted output.
func AnalyzeAllNamespaces(appName string) string {
	clientset, err := kubernetes.GetKubeClient()
	if err != nil {
		return styleRed.Render("❌  "+err.Error()) + "\n"
	}
	nsList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return styleRed.Render("❌  Failed to list namespaces: "+err.Error()) + "\n"
	}

	var b strings.Builder
	found := 0
	for _, ns := range nsList.Items {
		r, err := AnalyzeAppReport(appName, ns.Name)
		if err != nil {
			continue // app not in this namespace
		}
		found++
		b.WriteString(renderReport(r))
		b.WriteString("\n")
	}
	if found == 0 {
		b.WriteString(styleRed.Render(fmt.Sprintf("❌  No pods found for app '%s' in any namespace.", appName)) + "\n")
	}
	return b.String()
}

// CrashLoopInfo holds previous-container log excerpt for a CrashLoopBackOff pod.
type CrashLoopInfo struct {
	PodName     string
	Namespace   string
	Restarts    int32
	PreviousLog string
}

// DetectCrashLoops returns pods in CrashLoopBackOff with their previous container logs.
func DetectCrashLoops(namespace string) ([]CrashLoopInfo, error) {
	if namespace == "" {
		namespace = "default"
	}
	clientset, err := kubernetes.GetKubeClient()
	if err != nil {
		return nil, err
	}
	podList, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var results []CrashLoopInfo
	for _, pod := range podList.Items {
		for _, cs := range pod.Status.ContainerStatuses {
			isCrashLoop := cs.State.Waiting != nil && strings.EqualFold(cs.State.Waiting.Reason, "CrashLoopBackOff")
			hasBackoffSignals := cs.RestartCount > 0 && (!cs.Ready || (cs.State.Waiting != nil && strings.Contains(strings.ToLower(cs.State.Waiting.Reason), "backoff")))
			if isCrashLoop || hasBackoffSignals {
				prevLog, _ := kubernetes.GetPodPreviousLogs(clientset, namespace, pod.Name, cs.Name)
				results = append(results, CrashLoopInfo{
					PodName:     pod.Name,
					Namespace:   namespace,
					Restarts:    cs.RestartCount,
					PreviousLog: prevLog,
				})
			}
		}
	}
	return results, nil
}

// formatCopilotSuggestion generates Copilot-style structured suggestions
func formatCopilotSuggestion(appName, namespace, podName, status string, restarts int32, logs, events string) string {
	suggestion := diagnostics.EnhancedAnalyzeWithContext(appName, namespace, podName, status, restarts, logs, events)
	if suggestion == nil {
		return ""
	}
	return suggestion.Format()
}
