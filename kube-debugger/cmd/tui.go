package cmd

import (
	"github.com/spf13/cobra"
	"kube-debugger/pkg/tui"
)

var tuiNamespace string

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	Long:  "Interactive TUI with pod selector, live log viewer, and context switcher.",
	Run: func(cmd *cobra.Command, args []string) {
		setAdvisorContextLine("command=tui")
		setAdvisorContextLine("namespace=" + tuiNamespace)
		tui.StartTUI(tuiNamespace)
	},
}

func init() {
	tuiCmd.Flags().StringVarP(&tuiNamespace, "namespace", "n", "default", "Kubernetes namespace to inspect")
	rootCmd.AddCommand(tuiCmd)
}
