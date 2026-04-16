package cmd

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"kube-debugger/pkg/analyzer"
)

var (
	reportFormat    string
	reportOutput    string
	reportNamespace string
)

var htmlReportTmpl = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>KubeAid Debug Report — {{.AppName}}</title>
  <style>
    body { font-family: sans-serif; background: #f4f6f9; color: #333; margin: 2rem; }
    h1 { color: #1a73e8; }
    h2 { margin-top: 1.5rem; border-bottom: 1px solid #ccc; padding-bottom: 4px; }
    pre { background: #272822; color: #f8f8f2; padding: 1rem; border-radius: 6px; overflow-x: auto; }
    ul { line-height: 1.8; }
    table { border-collapse: collapse; width: 100%; margin-bottom: 1rem; }
    th, td { text-align: left; padding: 6px 12px; border-bottom: 1px solid #ddd; }
    th { background: #e8eaf6; }
    .meta { color: #666; font-size: 0.9rem; margin-bottom: 1rem; }
    .hint { background: #e8f5e9; border-left: 4px solid #43a047; padding: 0.75rem 1rem; border-radius: 4px; }
    .score-bar { height: 14px; background: #e0e0e0; border-radius: 7px; overflow: hidden; display: inline-block; width: 200px; vertical-align: middle; margin-right: 8px; }
    .score-fill { height: 100%; border-radius: 7px; }
    .score-high  { background: #43a047; }
    .score-med   { background: #fb8c00; }
    .score-low   { background: #e53935; }
  </style>
</head>
<body>
  <h1>KubeAid Debug Report</h1>
  <div class="meta">
    App: <strong>{{.AppName}}</strong> &nbsp;|&nbsp;
    Namespace: <strong>{{.Namespace}}</strong> &nbsp;|&nbsp;
    Total Pods: <strong>{{.PodCount}}</strong> &nbsp;|&nbsp;
    Generated: {{.GeneratedAt}}
  </div>

  <h2>Primary Pod</h2>
  <table>
    <tr><th>Field</th><th>Value</th></tr>
    <tr><td>Pod Name</td><td>{{.PodName}}</td></tr>
    <tr><td>Status</td><td>{{.Status}}</td></tr>
    <tr><td>Ready</td><td>{{.Ready}}</td></tr>
    <tr><td>Restarts</td><td>{{.RestartCount}}</td></tr>
    <tr><td>Age</td><td>{{.Age}}</td></tr>
  </table>

  <h2>Health Score</h2>
  {{$cls := "score-high"}}{{if lt .HealthScore 80}}{{$cls = "score-med"}}{{end}}{{if lt .HealthScore 50}}{{$cls = "score-low"}}{{end}}
  <div class="score-bar"><div class="score-fill {{$cls}}" style="width:{{.HealthScore}}%"></div></div>
  <strong>{{.HealthScore}} / 100</strong>

  {{if gt .PodCount 1}}
  <h2>All Pods</h2>
  <table>
    <tr><th>Pod</th><th>Status</th><th>Ready</th><th>Restarts</th><th>Age</th></tr>
    {{range .Pods}}<tr><td>{{.Name}}</td><td>{{.Status}}</td><td>{{.Ready}}</td><td>{{.RestartCount}}</td><td>{{.Age}}</td></tr>{{end}}
  </table>
  {{end}}

  <h2>AI Hint</h2>
  <div class="hint">{{.AIHint}}</div>

  <h2>Suggestions</h2>
  <ul>{{range .Suggestions}}<li>{{.}}</li>{{end}}</ul>

  <h2>Last Logs</h2>
  <pre>{{.Logs}}</pre>

  <h2>Events</h2>
  <pre>{{.Events}}</pre>

  <h2>Resource Usage</h2>
  <pre>{{.Resources}}</pre>
</body>
</html>`

var reportCmd = &cobra.Command{
	Use:   "report [app-name]",
	Short: "Export debug report for an app",
	Long:  "Export a debug report in text, JSON, or HTML format. Use --format to choose output type and --output to write to a file.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		r, err := analyzer.AnalyzeAppReport(appName, reportNamespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
			os.Exit(1)
		}

		var content string
		switch strings.ToLower(reportFormat) {
		case "json":
			b, err := json.MarshalIndent(r, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to marshal JSON: %v\n", err)
				os.Exit(1)
			}
			content = string(b)

		case "html":
			tmpl, err := template.New("report").Parse(htmlReportTmpl)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to parse HTML template: %v\n", err)
				os.Exit(1)
			}
			var buf strings.Builder
			if err := tmpl.Execute(&buf, r); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to render HTML report: %v\n", err)
				os.Exit(1)
			}
			content = buf.String()

		default: // "text"
			var sb strings.Builder
			sb.WriteString("KubeAid Debug Report\n")
			sb.WriteString("====================\n")
			sb.WriteString(fmt.Sprintf("App:          %s\n", r.AppName))
			sb.WriteString(fmt.Sprintf("Namespace:    %s\n", r.Namespace))
			sb.WriteString(fmt.Sprintf("Total Pods:   %d\n", r.PodCount))
			sb.WriteString(fmt.Sprintf("Pod:          %s\n", r.PodName))
			sb.WriteString(fmt.Sprintf("Status:       %s\n", r.Status))
			sb.WriteString(fmt.Sprintf("Ready:        %s\n", r.Ready))
			sb.WriteString(fmt.Sprintf("Restarts:     %d\n", r.RestartCount))
			sb.WriteString(fmt.Sprintf("Age:          %s\n", r.Age))
			sb.WriteString(fmt.Sprintf("Health Score: %d/100\n", r.HealthScore))
			sb.WriteString(fmt.Sprintf("Generated:    %s\n\n", r.GeneratedAt.Format("2006-01-02 15:04:05 UTC")))
			if len(r.Pods) > 1 {
				sb.WriteString("--- All Pods ---\n")
				sb.WriteString(fmt.Sprintf("%-42s %-16s %-6s %-9s %s\n", "POD", "STATUS", "READY", "RESTARTS", "AGE"))
				for _, p := range r.Pods {
					sb.WriteString(fmt.Sprintf("%-42s %-16s %-6s %-9d %s\n", p.Name, p.Status, p.Ready, p.RestartCount, p.Age))
				}
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("AI Hint:\n%s\n\n", r.AIHint))
			sb.WriteString("Suggestions:\n")
			for _, s := range r.Suggestions {
				sb.WriteString(fmt.Sprintf("  - %s\n", s))
			}
			sb.WriteString(fmt.Sprintf("\nLast Logs:\n%s\n", r.Logs))
			sb.WriteString(fmt.Sprintf("\nEvents:\n%s\n", r.Events))
			sb.WriteString(fmt.Sprintf("\nResource Usage:\n%s\n", r.Resources))
			content = sb.String()
		}

		if reportOutput != "" {
			if err := os.WriteFile(reportOutput, []byte(content), 0o644); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Failed to write report to %s: %v\n", reportOutput, err)
				os.Exit(1)
			}
			fmt.Printf("✅ Report saved to %s\n", reportOutput)
		} else {
			fmt.Println(content)
		}
	},
}

func init() {
	reportCmd.Flags().StringVarP(&reportFormat, "format", "f", "text", "Output format: text, json, html")
	reportCmd.Flags().StringVarP(&reportOutput, "output", "o", "", "Write report to this file path (default: stdout)")
	reportCmd.Flags().StringVarP(&reportNamespace, "namespace", "n", "", "Kubernetes namespace (default: \"default\")")
	rootCmd.AddCommand(reportCmd)
}

