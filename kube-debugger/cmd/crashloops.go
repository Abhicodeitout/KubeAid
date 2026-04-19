package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"kube-debugger/pkg/analyzer"
)

var crashNamespace string

var crashloopsCmd = &cobra.Command{
	Use:   "crashloops",
	Short: "Detect CrashLoopBackOff pods and show previous container logs",
	Long:  "Scans all pods in the given namespace for CrashLoopBackOff and shows the previous container logs for each.",
	Run: func(cmd *cobra.Command, args []string) {
		setAdvisorContextLine("command=crashloops")
		setAdvisorContextLine("namespace=" + crashNamespace)
		infos, err := analyzer.DetectCrashLoops(crashNamespace)
		if err != nil {
			fmt.Fprintf(os.Stderr, "❌ %v\n", err)
			setAdvisorContextLine("result=error")
			os.Exit(1)
		}
		setAdvisorContextLine(fmt.Sprintf("crashloop_count=%d", len(infos)))
		if len(infos) == 0 {
			fmt.Printf("✅ No CrashLoopBackOff pods found in namespace '%s'\n", crashNamespace)
			return
		}
		fmt.Printf("🔴 Found %d pod(s) in CrashLoopBackOff in namespace '%s'\n\n", len(infos), crashNamespace)
		for _, info := range infos {
			fmt.Printf("Pod: %s  (restarts: %d)\n", info.PodName, info.Restarts)
			fmt.Println(strings.Repeat("─", 60))
			if info.PreviousLog == "" {
				fmt.Println("  (no previous log available)")
			} else {
				for _, line := range strings.Split(strings.TrimSpace(info.PreviousLog), "\n") {
					fmt.Println("  " + line)
				}
			}
			fmt.Println()
		}
	},
}

func init() {
	crashloopsCmd.Flags().StringVarP(&crashNamespace, "namespace", "n", "default", "Kubernetes namespace to scan")
	rootCmd.AddCommand(crashloopsCmd)
}
