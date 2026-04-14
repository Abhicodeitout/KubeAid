package cmd

import (
	"os"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kube-debugger",
	Short: "Smart Kubernetes Debug CLI",
	Long:  `Diagnose and fix Kubernetes pod issues instantly.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
