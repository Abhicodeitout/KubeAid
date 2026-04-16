package diagnostics

import (
	"strings"
	"testing"
)

func TestAnalyzeLogsAI(t *testing.T) {
	cases := []struct {
		log string
		expect string
	}{
		{"connection refused on port 5432", "Network connectivity issue"},
		{"OOMKilled", "out-of-memory"},
		{"ImagePullBackOff", "Image pull failed"},
		{"all good", "No obvious pattern detected"},
	}
	for _, c := range cases {
		result := AnalyzeLogsAI(c.log)
		if c.expect != "" && !strings.Contains(result, c.expect) {
			t.Errorf("Expected '%s' in result for log '%s', got '%s'", c.expect, c.log, result)
		}
	}
}
