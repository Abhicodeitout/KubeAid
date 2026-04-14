package diagnostics

import "testing"

func TestAnalyzeLogsAI(t *testing.T) {
	cases := []struct{
		log string
		expect string
	}{
		{"connection refused on port 5432", "DB connection issue"},
		{"OOMKilled", "out-of-memory"},
		{"ImagePullBackOff", "Image pull failed"},
		{"all good", "No obvious issues"},
	}
	for _, c := range cases {
		result := AnalyzeLogsAI(c.log)
		if c.expect != "" && !contains(result, c.expect) {
			t.Errorf("Expected '%s' in result for log '%s', got '%s'", c.expect, c.log, result)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && (substr == "" || (len(substr) > 0 && (stringContains(s, substr))))
}

func stringContains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[0:len(substr)] == substr || stringContains(s[1:], substr))))
}
