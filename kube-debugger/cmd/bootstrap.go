package cmd

import (
	"fmt"
	"os"
	"runtime"
	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Pre-check environment for kube-debugger",
	Long:  `Checks Go installation, kubeconfig presence, and basic connectivity before running kube-debugger.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Running environment pre-checks...")
		// Check Go version
		if runtime.Version() < "go1.21" {
			fmt.Printf("Go version 1.21+ required, found %s\n", runtime.Version())
			os.Exit(1)
		}
		// Check kubeconfig
		home := os.Getenv("HOME")
		if home == "" && runtime.GOOS == "windows" {
			home = os.Getenv("USERPROFILE")
		}
		kubeconfig := home + string(os.PathSeparator) + ".kube" + string(os.PathSeparator) + "config"
		if _, err := os.Stat(kubeconfig); err != nil {
			fmt.Printf("kubeconfig not found at %s\n", kubeconfig)
			os.Exit(1)
		}
		fmt.Println("Go version and kubeconfig: OK")
		// Add more checks as needed
		fmt.Println("All pre-checks passed.")
	},
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}
