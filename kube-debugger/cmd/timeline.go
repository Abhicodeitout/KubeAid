package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"kube-debugger/pkg/analyzer"
	"kube-debugger/pkg/security"
)

var timelineNamespace string

var timelineCmd = &cobra.Command{
	Use:   "timeline [app-name]",
	Short: "Reconstruct an ordered incident timeline for an app",
	Long:  "Reconstruct the sequence of failure causes from pod events, restarts, probe failures, and important log signals.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		secMgr := security.GetSecurityManager()

		if err := security.ValidateAppName(appName); err != nil {
			fmt.Fprintf(os.Stderr, "❌ Invalid app name: %v\n", err)
			_ = secMgr.LogCommand("timeline", args, appName, timelineNamespace, err)
			requestExitCode(1)
			return
		}

		if timelineNamespace != "" {
			if err := security.ValidateNamespace(timelineNamespace); err != nil {
				fmt.Fprintf(os.Stderr, "❌ Invalid namespace: %v\n", err)
				_ = secMgr.LogCommand("timeline", args, appName, timelineNamespace, err)
				requestExitCode(1)
				return
			}
		}

		setAdvisorContextLine("command=timeline")
		setAdvisorContextLine("app=" + appName)
		if timelineNamespace != "" {
			setAdvisorContextLine("namespace=" + timelineNamespace)
		}

		report, err := analyzer.AnalyzeTimeline(appName, timelineNamespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
			setAdvisorContextLine("result=error")
			setAdvisorContextLine("error=" + err.Error())
			_ = secMgr.LogCommand("timeline", args, appName, timelineNamespace, err)
			requestExitCode(1)
			return
		}

		setAdvisorContextLine(fmt.Sprintf("namespace=%s", report.Namespace))
		setAdvisorContextLine(fmt.Sprintf("health_score=%d", report.HealthScore))
		setAdvisorContextLine(fmt.Sprintf("timeline_entries=%d", len(report.Timeline)))
		if report.FirstCause != nil {
			setAdvisorContextLine("first_cause=" + security.RedactSecrets(report.FirstCause.Summary))
		}

		fmt.Print(secMgr.FilterOutput(analyzer.RenderTimeline(report)))
		_ = secMgr.LogCommand("timeline", args, appName, report.Namespace, nil)
	},
}

func init() {
	timelineCmd.Flags().StringVarP(&timelineNamespace, "namespace", "n", "", "Kubernetes namespace (default: \"default\")")
	rootCmd.AddCommand(timelineCmd)
}