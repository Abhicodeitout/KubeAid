package cmd

import (
	"os"
	"strings"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kube-debugger",
	Short: "Smart Kubernetes Debug CLI",
	Long:  `Diagnose and fix Kubernetes pod issues instantly.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		resetAdvisorState()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if cmd == nil || cmd.Parent() == nil {
			return
		}
		name := cmd.CommandPath()
		// Do not append advisor output for completion script generation commands.
		if strings.Contains(name, "completion") {
			return
		}
		code := printAICommandAdvisor(cmd, args)
		if code > 0 {
			os.Exit(code)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
