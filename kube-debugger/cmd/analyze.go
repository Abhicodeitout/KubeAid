package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"kube-debugger/pkg/analyzer"
	"kube-debugger/pkg/history"
	"kube-debugger/pkg/security"
)

var (
	analyzeNamespace    string
	analyzeAllNS        bool
	analyzeWatch        bool
	analyzeInterval     int
	analyzeExitCode     bool
	analyzeThreshold    int
	analyzeWebhook      string
	analyzeWebhookThres int
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [app-name]",
	Short: "Analyze a Kubernetes app for issues",
	Long:  "Analyze pod status, logs, events, health score, and get AI fix suggestions.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		secMgr := security.GetSecurityManager()

		// Validate inputs
		if err := security.ValidateAppName(appName); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Invalid app name: %v\n", err)
			_ = secMgr.LogCommand("analyze", args, appName, analyzeNamespace, err)
			requestExitCode(1)
			return
		}

		if analyzeNamespace != "" {
			if err := security.ValidateNamespace(analyzeNamespace); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Invalid namespace: %v\n", err)
				_ = secMgr.LogCommand("analyze", args, appName, analyzeNamespace, err)
				requestExitCode(1)
				return
			}
		}

		// Validate interval
		if analyzeWatch {
			if err := security.ValidateInterval(analyzeInterval); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Invalid interval: %v\n", err)
				requestExitCode(1)
				return
			}
		}

		// Validate threshold
		if err := security.ValidateThreshold(analyzeThreshold); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Invalid threshold: %v\n", err)
			requestExitCode(1)
			return
		}

		// Validate webhook if provided
		if analyzeWebhook != "" {
			if err := security.ValidateWebhookURL(analyzeWebhook); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Invalid webhook URL: %v\n", err)
				requestExitCode(1)
				return
			}
		}

		setAdvisorContextLine("command=analyze")
		setAdvisorContextLine("app=" + appName)
		if analyzeNamespace != "" {
			setAdvisorContextLine("namespace=" + analyzeNamespace)
		}

		runOnce := func() (int, bool) {
			if analyzeAllNS {
				fmt.Print(analyzer.AnalyzeAllNamespaces(appName))
				setAdvisorContextLine("all_namespaces=true")
				return 100, false
			}
			r, err := analyzer.AnalyzeAppReport(appName, analyzeNamespace)
			if err != nil {
				fmt.Fprintf(os.Stderr, "❌ %v\n", err)
				setAdvisorContextLine("result=error")
				setAdvisorContextLine("error=" + err.Error())
				_ = secMgr.LogCommand("analyze", args, appName, analyzeNamespace, err)
				requestExitCode(1)
				return 0, true
			}

			// Apply output filtering for sensitive data
			output := analyzer.RenderReport(r)
			filteredOutput := secMgr.FilterOutput(output)
			fmt.Print(filteredOutput)

			setAdvisorContextLine(fmt.Sprintf("namespace=%s", r.Namespace))
			setAdvisorContextLine(fmt.Sprintf("status=%s", r.Status))
			setAdvisorContextLine(fmt.Sprintf("health_score=%d", r.HealthScore))
			setAdvisorContextLine(fmt.Sprintf("pod_count=%d", r.PodCount))

			// Log successful analysis
			_ = secMgr.LogCommand("analyze", args, appName, r.Namespace, nil)

			// Save to history
			history.Record(appName, r.Namespace, r.HealthScore)

			// Webhook alert
			if analyzeWebhook != "" && r.HealthScore < analyzeWebhookThres {
				sendWebhookAlert(analyzeWebhook, appName, r.Namespace, r.HealthScore)
			}

			return r.HealthScore, false
		}

		if analyzeWatch {
			ticker := time.NewTicker(time.Duration(analyzeInterval) * time.Second)
			defer ticker.Stop()
			setAdvisorContextLine(fmt.Sprintf("watch=true interval=%d", analyzeInterval))
			score, hadError := runOnce()
			if !hadError && analyzeExitCode && score < analyzeThreshold {
				setAdvisorContextLine(fmt.Sprintf("threshold_breach=true threshold=%d", analyzeThreshold))
				requestExitCode(2)
				return
			}
			for range ticker.C {
				fmt.Printf("\n\033[2J\033[H") // clear screen
				score, hadError = runOnce()
				if !hadError && analyzeExitCode && score < analyzeThreshold {
					setAdvisorContextLine(fmt.Sprintf("threshold_breach=true threshold=%d", analyzeThreshold))
					requestExitCode(2)
					return
				}
			}
		} else {
			score, hadError := runOnce()
			if !hadError && analyzeExitCode && score < analyzeThreshold {
				setAdvisorContextLine(fmt.Sprintf("threshold_breach=true threshold=%d", analyzeThreshold))
				requestExitCode(2)
				return
			}
		}
	},
}

func sendWebhookAlert(url, app, namespace string, score int) {
	payload := map[string]interface{}{
		"text":        fmt.Sprintf("⚠️ kube-debugger alert: *%s* in namespace *%s* has health score *%d/100*", app, namespace, score),
		"app":         app,
		"namespace":   namespace,
		"health_score": score,
	}
	b, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewReader(b)) //nolint:noctx
	if err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Webhook alert failed: %v\n", err)
		return
	}
	_ = resp.Body.Close()
}

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeNamespace, "namespace", "n", "", "Kubernetes namespace (default: \"default\")")
	analyzeCmd.Flags().BoolVarP(&analyzeAllNS, "all-namespaces", "A", false, "Scan the app across all namespaces")
	analyzeCmd.Flags().BoolVar(&analyzeWatch, "watch", false, "Continuously re-run analysis")
	analyzeCmd.Flags().IntVar(&analyzeInterval, "interval", 10, "Watch interval in seconds (used with --watch)")
	analyzeCmd.Flags().BoolVar(&analyzeExitCode, "exit-code", false, "Exit with code 2 when health score is below threshold")
	analyzeCmd.Flags().IntVar(&analyzeThreshold, "threshold", 80, "Health score threshold for --exit-code")
	analyzeCmd.Flags().StringVar(&analyzeWebhook, "alert-webhook", "", "POST a JSON alert to this URL when health drops below threshold")
	analyzeCmd.Flags().IntVar(&analyzeWebhookThres, "alert-threshold", 80, "Health score threshold for --alert-webhook")
	rootCmd.AddCommand(analyzeCmd)
}
