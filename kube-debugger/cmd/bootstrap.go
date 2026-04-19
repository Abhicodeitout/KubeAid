package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	bootstrapInstallOllama bool
	bootstrapStartOllama   bool
	bootstrapPullModel     string
	bootstrapSkipKube      bool
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Pre-check and optionally set up local AI tooling",
	Long:  `Checks the local environment and can install/start Ollama across Linux, macOS, and Windows.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		setAdvisorContextLine("command=bootstrap")
		setAdvisorContextLine(fmt.Sprintf("os=%s/%s", runtime.GOOS, runtime.GOARCH))
		setAdvisorContextLine(fmt.Sprintf("skip_kubeconfig_check=%t", bootstrapSkipKube))
		fmt.Println("Running environment pre-checks...")
		fmt.Printf("Detected OS: %s/%s\n", runtime.GOOS, runtime.GOARCH)

		if err := checkGoVersion(); err != nil {
			return err
		}
		fmt.Printf("Go version: %s\n", runtime.Version())

		if !bootstrapSkipKube {
			if err := checkKubeconfig(); err != nil {
				setAdvisorContextLine("kubeconfig=missing")
				return err
			}
			fmt.Println("kubeconfig: OK")
			setAdvisorContextLine("kubeconfig=ok")
		}

		if bootstrapInstallOllama || bootstrapStartOllama || bootstrapPullModel != "" {
			if err := ensureOllama(); err != nil {
				return err
			}
		}

		fmt.Println("All pre-checks passed.")
		setAdvisorContextLine("result=ok")
		return nil
	},
}

func init() {
	bootstrapCmd.Flags().BoolVar(&bootstrapInstallOllama, "install-ollama", false, "Install Ollama if it is missing")
	bootstrapCmd.Flags().BoolVar(&bootstrapStartOllama, "start-ollama", false, "Start Ollama if it is installed but not running")
	bootstrapCmd.Flags().StringVar(&bootstrapPullModel, "pull-model", "", "Pull an Ollama model after installation/start, e.g. llama3.2")
	bootstrapCmd.Flags().BoolVar(&bootstrapSkipKube, "skip-kubeconfig-check", false, "Skip kubeconfig presence check")
	rootCmd.AddCommand(bootstrapCmd)
}

func checkGoVersion() error {
	version := runtime.Version()
	if version == "" || !strings.HasPrefix(version, "go1.") {
		return fmt.Errorf("unable to determine Go version")
	}
	return nil
}

func checkKubeconfig() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("could not determine home directory: %w", err)
	}
	kubeconfig := filepath.Join(home, ".kube", "config")
	if _, err := os.Stat(kubeconfig); err != nil {
		return fmt.Errorf("kubeconfig not found at %s", kubeconfig)
	}
	return nil
}

func ensureOllama() error {
	installed := commandExists("ollama")
	if installed {
		fmt.Println("Ollama: installed")
	} else {
		fmt.Println("Ollama: not installed")
	}

	if !installed {
		if !bootstrapInstallOllama {
			return fmt.Errorf("Ollama is not installed. Re-run with --install-ollama to install it automatically")
		}
		fmt.Println("Installing Ollama...")
		if err := installOllama(); err != nil {
			return err
		}
		fmt.Println("Ollama install finished.")
	}

	needsRunning := bootstrapStartOllama || bootstrapPullModel != ""
	if needsRunning && !ollamaReachable() {
		fmt.Println("Ollama API is not responding. Starting Ollama...")
		if err := startOllama(); err != nil {
			return err
		}
		if err := waitForOllama(45 * time.Second); err != nil {
			return err
		}
		fmt.Println("Ollama: running")
	} else if needsRunning {
		fmt.Println("Ollama: already running")
	}

	if bootstrapPullModel != "" {
		fmt.Printf("Pulling Ollama model: %s\n", bootstrapPullModel)
		if err := runCommandStreaming("ollama", "pull", bootstrapPullModel); err != nil {
			return fmt.Errorf("failed to pull Ollama model %q: %w", bootstrapPullModel, err)
		}
		fmt.Printf("Model ready: %s\n", bootstrapPullModel)
		fmt.Println("Use these environment variables:")
		fmt.Println("  export KUBEAID_AI_PROVIDER=ollama")
		fmt.Printf("  export KUBEAID_AI_MODEL=%s\n", bootstrapPullModel)
	}

	return nil
}

func installOllama() error {
	switch runtime.GOOS {
	case "linux":
		return runCommandStreaming("sh", "-c", "curl -fsSL https://ollama.com/install.sh | sh")
	case "darwin":
		if commandExists("brew") {
			return runCommandStreaming("brew", "install", "ollama")
		}
		return runCommandStreaming("sh", "-c", "curl -fsSL https://ollama.com/install.sh | sh")
	case "windows":
		if commandExists("winget") {
			return runCommandStreaming("winget", "install", "-e", "--id", "Ollama.Ollama")
		}
		if commandExists("choco") {
			return runCommandStreaming("choco", "install", "ollama", "-y")
		}
		return errors.New("no supported Windows package manager found; install Ollama manually from https://ollama.com/download")
	default:
		return fmt.Errorf("unsupported OS for automatic Ollama install: %s", runtime.GOOS)
	}
}

func startOllama() error {
	if ollamaReachable() {
		return nil
	}

	switch runtime.GOOS {
	case "darwin":
		if err := exec.Command("open", "-a", "Ollama").Run(); err == nil {
			return nil
		}
	case "windows":
		if err := exec.Command("cmd", "/c", "start", "", "ollama", "app").Run(); err == nil {
			return nil
		}
	}

	cmd := exec.Command("ollama", "serve")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start Ollama: %w", err)
	}
	return nil
}

func waitForOllama(timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ollamaReachable() {
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("Ollama did not become ready within %s", timeout)
}

func ollamaReachable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runCommandStreaming(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
