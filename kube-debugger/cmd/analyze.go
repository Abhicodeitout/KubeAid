package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kube-debugger/pkg/analyzer"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze [app-name]",
	Short: "Analyze a Kubernetes app for issues",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		appName := args[0]
		result := analyzer.AnalyzeApp(appName)
		fmt.Println(result)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
