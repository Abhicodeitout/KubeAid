package analyzer

import (
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kube-debugger/pkg/kubernetes"
)

func TestBuildTimelineEntriesOrdersAndMarksCause(t *testing.T) {
	base := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	pod := &corev1.Pod{
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "api",
				LastTerminationState: corev1.ContainerState{
					Terminated: &corev1.ContainerStateTerminated{
						Reason:     "Error",
						Message:    "container exited after crash",
						ExitCode:   1,
						FinishedAt: metav1.NewTime(base.Add(2 * time.Minute)),
					},
				},
			}},
		},
	}

	events := []kubernetes.PodEvent{
		{Time: base.Add(1 * time.Minute), Type: "Warning", Reason: "FailedMount", Message: "volume missing"},
		{Time: base.Add(3 * time.Minute), Type: "Warning", Reason: "BackOff", Message: "back-off restarting failed container"},
		{Time: base.Add(4 * time.Minute), Type: "Normal", Reason: "Pulled", Message: "container image pulled"},
	}
	logs := []kubernetes.LogLine{{Time: base.Add(90 * time.Second), Message: "panic: failed to load config"}}

	entries := buildTimelineEntries("default", "api-123", pod, events, logs, base.Add(5*time.Minute))
	if len(entries) != 5 {
		t.Fatalf("expected 5 timeline entries, got %d", len(entries))
	}

	for i := 1; i < len(entries); i++ {
		if entries[i].Time.Before(entries[i-1].Time) {
			t.Fatalf("entries not sorted: %v before %v", entries[i].Time, entries[i-1].Time)
		}
	}

	if entries[0].Role != "first-cause" {
		t.Fatalf("expected first entry to be first-cause, got %q", entries[0].Role)
	}
	if entries[0].Summary != "FailedMount" {
		t.Fatalf("expected first cause summary FailedMount, got %q", entries[0].Summary)
	}
	if entries[1].Kind != "log" {
		t.Fatalf("expected second entry to be log, got %q", entries[1].Kind)
	}
	if entries[1].Role != "impact" {
		t.Fatalf("expected second entry to be marked as impact, got %q", entries[1].Role)
	}
}

func TestRenderTimelineShowsFirstCauseAndImpacts(t *testing.T) {
	base := time.Date(2026, 5, 1, 12, 0, 0, 0, time.UTC)
	report := &TimelineReport{
		AppName:     "payments",
		Namespace:   "prod",
		PodName:     "payments-abc",
		GeneratedAt: base.Add(10 * time.Minute),
		HealthScore: 42,
		FirstCause: &TimelineEntry{
			Time:     base,
			Kind:     "event",
			Summary:  "FailedScheduling",
			Details:  "0/3 nodes available",
			Severity: "critical",
			Role:     "first-cause",
		},
		Timeline: []TimelineEntry{
			{Time: base, Kind: "event", Summary: "FailedScheduling", Details: "0/3 nodes available", Severity: "critical", Role: "first-cause"},
			{Time: base.Add(2 * time.Minute), Kind: "probe", Summary: "Unhealthy", Details: "Readiness probe failed", Severity: "critical", Role: "impact"},
		},
	}

	output := RenderTimeline(report)
	for _, want := range []string{"First-cause candidate:", "FailedScheduling", "IMPACT", "ORDERED INCIDENT FLOW"} {
		if !strings.Contains(output, want) {
			t.Fatalf("expected rendered output to contain %q", want)
		}
	}
}