package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"kube-debugger/pkg/kubernetes"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "List and switch Kubernetes contexts",
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available kubeconfig contexts",
	Run: func(cmd *cobra.Command, args []string) {
		contexts, err := kubernetes.ListKubeContexts()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Available contexts:")
		for _, ctx := range contexts {
			fmt.Println("-", ctx)
		}
	},
}

var contextSwitchCmd = &cobra.Command{
	Use:   "switch [context-name]",
	Short: "Switch to a different kubeconfig context",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := kubernetes.SwitchKubeContext(args[0])
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Switched to context:", args[0])
	},
}

func init() {
	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextSwitchCmd)
	rootCmd.AddCommand(contextCmd)
}
