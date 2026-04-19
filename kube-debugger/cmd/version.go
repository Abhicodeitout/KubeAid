package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show kube-debugger version",
	Run: func(cmd *cobra.Command, args []string) {
		setAdvisorContextLine("command=version")
		setAdvisorContextLine("version=" + version)
		fmt.Println("kube-debugger version:", version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
