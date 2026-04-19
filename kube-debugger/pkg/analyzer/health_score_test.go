package analyzer

import "testing"

func TestComputeHealthScoreReadyPenalty(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		ready    string
		restarts int32
		events   string
		want     int
	}{
		{
			name:     "healthy running pod no penalties",
			status:   "Running",
			ready:    "1/1",
			restarts: 0,
			events:   "Started: Container started",
			want:     100,
		},
		{
			name:     "running but not ready gets readiness penalty",
			status:   "Running",
			ready:    "0/1",
			restarts: 0,
			events:   "Started: Container started",
			want:     70,
		},
		{
			name:     "crashloop non-ready combines penalties",
			status:   "CrashLoopBackOff",
			ready:    "0/1",
			restarts: 15,
			events:   "Warning BackOff: Back-off restarting failed container",
			want:     0,
		},
		{
			name:     "score floor at zero",
			status:   "CrashLoopBackOff",
			ready:    "0/1",
			restarts: 20,
			events:   "Warning failed\nWarning kill\nWarning failed\nWarning kill\nWarning failed\nWarning kill\nWarning failed\nWarning kill\nWarning failed\nWarning kill\n",
			want:     0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeHealthScore(tc.status, tc.ready, tc.restarts, tc.events)
			if got != tc.want {
				t.Fatalf("computeHealthScore(%q, %q, %d, events)=%d, want %d", tc.status, tc.ready, tc.restarts, got, tc.want)
			}
		})
	}
}
