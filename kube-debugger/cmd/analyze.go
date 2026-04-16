package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kube-debugger/pkg/analyzer"
)

var analyzeNamespace string

var analyzeCmd = &cobra.Command{
	Use:   "analyze [app-name]",
	Short: "Analyze a Kubernetes app for issues",
	Long:  "Analyze pod status, logs, events, health score, and get AI fix suggestions.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		result := analyzer.AnalyzeApp(args[0], analyzeNamespace)
		fmt.Print(result)
	},
}

func init() {
	analyzeCmd.Flags().StringVarP(&analyzeNamespace, "namespace", "n", "", "Kubernetes namespace (default: \"default\")")
	rootCmd.AddCommand(analyzeCmd)
}
