package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"kube-debugger/pkg/history"
)

var historyNamespace string

var historyCmd = &cobra.Command{
	Use:   "history [app-name]",
	Short: "Show health score history for an app",
	Long:  "Display trending health scores recorded by previous `analyze` runs.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setAdvisorContextLine("command=history")
		setAdvisorContextLine("app=" + args[0])
		if historyNamespace != "" {
			setAdvisorContextLine("namespace=" + historyNamespace)
		}
		fmt.Print(history.RenderHistory(args[0], historyNamespace))
	},
}

var clearHistoryCmd = &cobra.Command{
	Use:   "clear [app-name]",
	Short: "Clear health score history for an app (or all apps)",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		setAdvisorContextLine("command=history clear")
		dataFile := history.DataFilePath()
		if err := os.Remove(dataFile); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ History cleared.")
	},
}

func init() {
	historyCmd.Flags().StringVarP(&historyNamespace, "namespace", "n", "", "Filter by namespace")
	historyCmd.AddCommand(clearHistoryCmd)
	rootCmd.AddCommand(historyCmd)
}
