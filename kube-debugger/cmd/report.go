package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kube-debugger/pkg/analyzer"
)

var reportCmd = &cobra.Command{
	Use:   "report [app-name]",
	Short: "Export debug report for an app",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		report := analyzer.AnalyzeApp(appName)
		fmt.Println(report)
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)
}
