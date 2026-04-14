package cmd

import (
	"github.com/spf13/cobra"
	"kube-debugger/pkg/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive terminal UI",
	Run: func(cmd *cobra.Command, args []string) {
		tui.StartTUI()
	},
}

func init() {
	rootCmd.AddCommand(tuiCmd)
}
