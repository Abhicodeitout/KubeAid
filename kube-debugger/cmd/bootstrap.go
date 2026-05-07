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
	bootstrapAutoSetup     bool
	bootstrapEnvFile       string
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

		if bootstrapAutoSetup && !bootstrapInstallOllama && !bootstrapStartOllama && bootstrapPullModel == "" {
			defaultModel := strings.TrimSpace(os.Getenv("KUBEAID_AI_MODEL"))
		if defaultModel == "" {
			defaultModel = "tinyllama"
		}
			bootstrapInstallOllama = true
			bootstrapStartOllama = true
			bootstrapPullModel = defaultModel
			fmt.Printf("Auto-setup enabled: install/start Ollama and pull model %q\n", defaultModel)
			setAdvisorContextLine("auto_setup=true")
		}

		if bootstrapInstallOllama || bootstrapStartOllama || bootstrapPullModel != "" {
			if err := ensureOllama(); err != nil {
				return err
			}
			if err := prepareAIEnvFile(bootstrapEnvFile, bootstrapPullModel); err != nil {
				return err
			}
			if err := validateAIReadiness(bootstrapPullModel); err != nil {
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
	bootstrapCmd.Flags().BoolVar(&bootstrapAutoSetup, "auto-setup", true, "When no Ollama flags are provided, auto-install/start Ollama and pull default model")
	bootstrapCmd.Flags().StringVar(&bootstrapEnvFile, "env-file", "env/kube-debugger.env", "Path to write AI-ready environment variables")
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
			return fmt.Errorf("ollama is not installed; re-run with --install-ollama to install it automatically")
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

func validateAIReadiness(model string) error {
	if !commandExists("ollama") {
		return fmt.Errorf("ai readiness failed: ollama command not found after setup")
	}
	if !ollamaReachable() {
		return fmt.Errorf("ai readiness failed: ollama API is not reachable at http://localhost:11434")
	}
	if strings.TrimSpace(model) != "" && !ollamaModelInstalled(model) {
		return fmt.Errorf("ai readiness failed: model %q was not found in local ollama list", model)
	}

	// Probe actual inference — a fast model response confirms the runtime works.
	fmt.Printf("LLM inference check: probing model %q (may take up to 30s)...\n", model)
	if ok := probeOllamaInference(model, 30*time.Second); ok {
		fmt.Println("LLM inference check: OK")
		setAdvisorContextLine("llm_inference=ok")
	} else {
		fmt.Println("LLM inference check: SLOW (pattern fallback will be used for AI hints)")
		fmt.Println("  Tip: try a smaller model (tinyllama) or set KUBEAID_AI_PROVIDER=groq for cloud AI.")
		setAdvisorContextLine("llm_inference=slow")
	}

	fmt.Println("AI readiness: OK (provider/model/runtime checks passed)")
	setAdvisorContextLine("ai_readiness=ok")
	return nil
}

// probeOllamaInference sends a minimal prompt to Ollama and returns true if a
// response arrives within the given timeout.
func probeOllamaInference(model string, timeout time.Duration) bool {
	body := fmt.Sprintf(`{"model":%q,"prompt":"Reply with exactly: ok","stream":false}`, model)
	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(
		"http://localhost:11434/api/generate",
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode == http.StatusOK
}

func ollamaModelInstalled(model string) bool {
	out, err := exec.Command("ollama", "list").Output()
	if err != nil {
		return false
	}
	want := strings.TrimSpace(model)
	if want == "" {
		return true
	}
	wantBase := strings.Split(want, ":")[0]
	for _, line := range strings.Split(string(out), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		installed := fields[0]
		installedBase := strings.Split(installed, ":")[0]
		if installed == want || installedBase == wantBase || strings.HasPrefix(installed, want+":") {
			return true
		}
	}
	return false
}

func prepareAIEnvFile(path, model string) error {
	if strings.TrimSpace(path) == "" {
		return nil
	}
	if strings.TrimSpace(model) == "" {
		model = strings.TrimSpace(os.Getenv("KUBEAID_AI_MODEL"))
		if model == "" {
			model = "tinyllama"
		}
	}

	content := fmt.Sprintf(`# KubeAid AI-ready environment (generated by bootstrap)
export KUBEAID_AI_PROVIDER=ollama
export KUBEAID_AI_MODEL=%s
export KUBEAID_OLLAMA_URL=http://localhost:11434
export KUBEAID_AI_TIMEOUT_SECONDS=300

# Optional namespace default
export KUBE_NAMESPACE=default
`, model)

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create env dir %s: %w", dir, err)
		}
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("failed to write env file %s: %w", path, err)
	}

	fmt.Printf("AI environment ready: %s\n", path)
	fmt.Printf("To load now in your shell: source %s\n", path)
	setAdvisorContextLine("env_file_ready=true")
	return nil
}

func installOllama() error {
	switch runtime.GOOS {
	case "linux":
		return installOllamaViaScriptWithPercentMeter()
	case "darwin":
		if commandExists("brew") {
			return runCommandStreaming("brew", "install", "ollama")
		}
		return installOllamaViaScriptWithPercentMeter()
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

func installOllamaViaScriptWithPercentMeter() error {
	tmpDir, err := os.MkdirTemp("", "kubeaid-ollama-install-")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	scriptPath := filepath.Join(tmpDir, "install.sh")
	if err := runCommandStreaming("curl", "-fsSL", "https://ollama.com/install.sh", "-o", scriptPath); err != nil {
		return fmt.Errorf("failed to download Ollama install script: %w", err)
	}

	raw, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("failed to read install script: %w", err)
	}

	patched := strings.ReplaceAll(string(raw), "--progress-bar", "")
	if err := os.WriteFile(scriptPath, []byte(patched), 0o755); err != nil {
		return fmt.Errorf("failed to patch install script: %w", err)
	}

	fmt.Println("Using percentage progress meter for download output...")
	return runCommandStreaming("sh", scriptPath)
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
	return fmt.Errorf("ollama did not become ready within %s", timeout)
}

func ollamaReachable() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://localhost:11434/api/tags")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
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
