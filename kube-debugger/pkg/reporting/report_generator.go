package reporting

import (
	"fmt"
	"time"
)

// Report represents a comprehensive report
type Report struct {
	Title            string
	Summary          string
	GeneratedAt      time.Time
	Period           string
	HealthScore      int
	TotalIssues      int
	CriticalIssues   int
	WarningIssues    int
	MetricsSummary   map[string]interface{}
	Recommendations  []string
	CostAnalysis     map[string]interface{}
	SecurityFindings []string
	Timeline         []ReportEvent
}

// ReportEvent represents an event in the report
type ReportEvent struct {
	Time       time.Time
	Type       string // alert, error, warning, info
	Title      string
	Message    string
	Severity   string
}

// ReportGenerator generates reports
type ReportGenerator struct {
	report Report
}

// New creates a new report generator
func New(title, summary string) *ReportGenerator {
	return &ReportGenerator{
		report: Report{
			Title:           title,
			Summary:         summary,
			GeneratedAt:     time.Now(),
			MetricsSummary:  make(map[string]interface{}),
			CostAnalysis:    make(map[string]interface{}),
			Recommendations: make([]string, 0),
			SecurityFindings: make([]string, 0),
			Timeline:        make([]ReportEvent, 0),
		},
	}
}

// SetHealthScore sets the health score
func (rg *ReportGenerator) SetHealthScore(score int) {
	rg.report.HealthScore = score
}

// AddIssues adds issues to the report
func (rg *ReportGenerator) AddIssues(critical, warning int) {
	rg.report.CriticalIssues = critical
	rg.report.WarningIssues = warning
	rg.report.TotalIssues = critical + warning
}

// AddMetric adds a metric to the report
func (rg *ReportGenerator) AddMetric(name string, value interface{}) {
	rg.report.MetricsSummary[name] = value
}

// AddRecommendation adds a recommendation
func (rg *ReportGenerator) AddRecommendation(rec string) {
	rg.report.Recommendations = append(rg.report.Recommendations, rec)
}

// AddSecurityFinding adds a security finding
func (rg *ReportGenerator) AddSecurityFinding(finding string) {
	rg.report.SecurityFindings = append(rg.report.SecurityFindings, finding)
}

// AddEvent adds an event to the timeline
func (rg *ReportGenerator) AddEvent(eventType, title, message, severity string) {
	rg.report.Timeline = append(rg.report.Timeline, ReportEvent{
		Time:     time.Now(),
		Type:     eventType,
		Title:    title,
		Message:  message,
		Severity: severity,
	})
}

// SetCostAnalysis sets cost analysis data
func (rg *ReportGenerator) SetCostAnalysis(data map[string]interface{}) {
	rg.report.CostAnalysis = data
}

// SetPeriod sets the reporting period
func (rg *ReportGenerator) SetPeriod(period string) {
	rg.report.Period = period
}

// GenerateMarkdown generates a Markdown report
func (rg *ReportGenerator) GenerateMarkdown() string {
	markdown := fmt.Sprintf("# %s\n\n", rg.report.Title)

	// Summary
	markdown += fmt.Sprintf("**Generated:** %s\n", rg.report.GeneratedAt.Format("2006-01-02 15:04:05"))
	markdown += fmt.Sprintf("**Period:** %s\n\n", rg.report.Period)

	// Health Score
	markdown += "## Health Overview\n\n"
	healthStatus := "🟢 Healthy"
	if rg.report.HealthScore < 70 {
		healthStatus = "🔴 Critical"
	} else if rg.report.HealthScore < 85 {
		healthStatus = "🟡 Warning"
	}
	markdown += fmt.Sprintf("- **Overall Health:** %s (%d/100)\n", healthStatus, rg.report.HealthScore)
	markdown += fmt.Sprintf("- **Critical Issues:** %d\n", rg.report.CriticalIssues)
	markdown += fmt.Sprintf("- **Warnings:** %d\n\n", rg.report.WarningIssues)

	// Metrics Summary
	if len(rg.report.MetricsSummary) > 0 {
		markdown += "## Metrics Summary\n\n"
		for key, value := range rg.report.MetricsSummary {
			markdown += fmt.Sprintf("- **%s:** %v\n", key, value)
		}
		markdown += "\n"
	}

	// Cost Analysis
	if len(rg.report.CostAnalysis) > 0 {
		markdown += "## Cost Analysis\n\n"
		for key, value := range rg.report.CostAnalysis {
			markdown += fmt.Sprintf("- **%s:** %v\n", key, value)
		}
		markdown += "\n"
	}

	// Security Findings
	if len(rg.report.SecurityFindings) > 0 {
		markdown += "## Security Findings\n\n"
		for _, finding := range rg.report.SecurityFindings {
			markdown += fmt.Sprintf("- ⚠️ %s\n", finding)
		}
		markdown += "\n"
	}

	// Recommendations
	if len(rg.report.Recommendations) > 0 {
		markdown += "## Recommendations\n\n"
		for i, rec := range rg.report.Recommendations {
			markdown += fmt.Sprintf("%d. %s\n", i+1, rec)
		}
		markdown += "\n"
	}

	// Timeline
	if len(rg.report.Timeline) > 0 {
		markdown += "## Event Timeline\n\n"
		markdown += "| Time | Type | Title | Severity |\n"
		markdown += "|------|------|-------|----------|\n"
		for _, event := range rg.report.Timeline {
			markdown += fmt.Sprintf("| %s | %s | %s | %s |\n",
				event.Time.Format("15:04:05"), event.Type, event.Title, event.Severity)
		}
	}

	return markdown
}

// GenerateHTML generates an HTML report
func (rg *ReportGenerator) GenerateHTML() string {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>KubeAid Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 20px; border-radius: 5px; }
        h1 { color: #333; border-bottom: 3px solid #0066cc; }
        .health-score { font-size: 24px; font-weight: bold; }
        .metric { display: inline-block; width: 23%; margin: 1%; padding: 10px; background: #f9f9f9; border-radius: 5px; }
        .critical { color: #d32f2f; }
        .warning { color: #f57c00; }
        .ok { color: #388e3c; }
        table { width: 100%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background: #0066cc; color: white; }
    </style>
</head>
<body>
    <div class="container">
`

	html += fmt.Sprintf("<h1>%s</h1>\n", rg.report.Title)
	html += fmt.Sprintf("<p>Generated: %s | Period: %s</p>\n", rg.report.GeneratedAt.Format("2006-01-02 15:04:05"), rg.report.Period)

	// Health Section
	healthClass := "ok"
	if rg.report.HealthScore < 70 {
		healthClass = "critical"
	} else if rg.report.HealthScore < 85 {
		healthClass = "warning"
	}

	html += fmt.Sprintf(`
    <div class="metric">
        <div class="health-score %s">%d/100</div>
        <p>Health Score</p>
    </div>
    <div class="metric critical">%d<br>Critical Issues</div>
    <div class="metric warning">%d<br>Warnings</div>
`, healthClass, rg.report.HealthScore, rg.report.CriticalIssues, rg.report.WarningIssues)

	// Recommendations
	if len(rg.report.Recommendations) > 0 {
		html += "<h2>Recommendations</h2><ul>\n"
		for _, rec := range rg.report.Recommendations {
			html += fmt.Sprintf("<li>%s</li>\n", rec)
		}
		html += "</ul>\n"
	}

	html += `
    </div>
</body>
</html>
`
	return html
}

// GenerateJSON generates a JSON report (simplified)
func (rg *ReportGenerator) GenerateJSON() string {
	return fmt.Sprintf(`{
  "title": "%s",
  "generated_at": "%s",
  "period": "%s",
  "health_score": %d,
  "total_issues": %d,
  "critical_issues": %d,
  "warning_issues": %d,
  "recommendations_count": %d,
  "security_findings_count": %d
}`,
		rg.report.Title,
		rg.report.GeneratedAt.Format("2006-01-02T15:04:05Z07:00"),
		rg.report.Period,
		rg.report.HealthScore,
		rg.report.TotalIssues,
		rg.report.CriticalIssues,
		rg.report.WarningIssues,
		len(rg.report.Recommendations),
		len(rg.report.SecurityFindings),
	)
}

// GenerateReport generates report and returns as string
func (rg *ReportGenerator) GenerateReport(format string) (string, error) {
	switch format {
	case "markdown":
		return rg.GenerateMarkdown(), nil
	case "html":
		return rg.GenerateHTML(), nil
	case "json":
		return rg.GenerateJSON(), nil
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
