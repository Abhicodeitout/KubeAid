package diagnostics

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Provider names
const (
	ProviderOllama = "ollama"
	ProviderGroq   = "groq"
)

// LLMConfig is resolved once from environment variables.
type LLMConfig struct {
	Provider string // "ollama" | "groq" | ""
	Model    string
	BaseURL  string // Ollama only
	APIKey   string // Groq only
}

// ResolveLLMConfig reads environment variables to build an LLMConfig.
//
//	KUBEAID_AI_PROVIDER  = "ollama" | "groq"            (unset → pattern-based fallback)
//	KUBEAID_AI_MODEL     = model name override
//	KUBEAID_OLLAMA_URL   = Ollama base URL               (default: http://localhost:11434)
//	GROQ_API_KEY         = Groq API key
func ResolveLLMConfig() LLMConfig {
	providerRaw, providerSet := os.LookupEnv("KUBEAID_AI_PROVIDER")
	modelRaw, modelSet := os.LookupEnv("KUBEAID_AI_MODEL")

	provider := strings.ToLower(strings.TrimSpace(providerRaw))
	model := strings.TrimSpace(modelRaw)

	fileVars := map[string]string{}
	if !providerSet || !modelSet {
		fileVars = loadAIConfigFromEnvFile()
		if !providerSet {
			provider = strings.ToLower(strings.TrimSpace(fileVars["KUBEAID_AI_PROVIDER"]))
		}
		if !modelSet {
			model = strings.TrimSpace(fileVars["KUBEAID_AI_MODEL"])
		}
	}

	// Respect explicit empty provider (used by smoke tests/CI to disable AI).
	if providerSet && strings.TrimSpace(providerRaw) == "" {
		return LLMConfig{}
	}

	switch provider {
	case ProviderOllama:
		if model == "" {
			model = "llama3.2"
		}
		baseURL := strings.TrimSpace(os.Getenv("KUBEAID_OLLAMA_URL"))
		if baseURL == "" {
			baseURL = strings.TrimSpace(fileVars["KUBEAID_OLLAMA_URL"])
		}
		if baseURL == "" {
			baseURL = "http://localhost:11434"
		}
		return LLMConfig{Provider: ProviderOllama, Model: model, BaseURL: strings.TrimRight(baseURL, "/")}

	case ProviderGroq:
		if model == "" {
			model = "llama3-8b-8192"
		}
		apiKey := os.Getenv("GROQ_API_KEY")
		if strings.TrimSpace(apiKey) == "" {
			apiKey = fileVars["GROQ_API_KEY"]
		}
		return LLMConfig{
			Provider: ProviderGroq,
			Model:    model,
			BaseURL:  "https://api.groq.com",
			APIKey:   apiKey,
		}

	default:
		return LLMConfig{} // no LLM configured
	}
}

func loadAIConfigFromEnvFile() map[string]string {
	paths := []string{
		"env/kube-debugger.env",
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		paths = append(paths,
			filepath.Join(exeDir, "env", "kube-debugger.env"),
			filepath.Join(exeDir, "..", "env", "kube-debugger.env"),
		)
	}

	for _, p := range paths {
		vars, ok := parseSimpleEnvFile(p)
		if ok {
			return vars
		}
	}
	return map[string]string{}
}

func parseSimpleEnvFile(path string) (map[string]string, bool) {
	f, err := os.Open(path)
	if err != nil {
		return nil, false
	}
	defer func() { _ = f.Close() }()

	vars := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		idx := strings.Index(line, "=")
		if idx <= 0 {
			continue
		}
		k := strings.TrimSpace(line[:idx])
		v := strings.TrimSpace(line[idx+1:])
		v = strings.Trim(v, "\"'")
		vars[k] = v
	}
	if err := s.Err(); err != nil {
		return nil, false
	}
	return vars, true
}

// CallLLM sends a prompt to the configured provider and returns the response text.
// Returns an error if the provider is not configured or the call fails.
func CallLLM(cfg LLMConfig, prompt string) (string, error) {
	switch cfg.Provider {
	case ProviderOllama:
		return callOllama(cfg, prompt)
	case ProviderGroq:
		return callGroq(cfg, prompt)
	default:
		return "", fmt.Errorf("no LLM provider configured")
	}
}

// ── Ollama ────────────────────────────────────────────────────────────────────

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Error    string `json:"error,omitempty"`
}

func callOllama(cfg LLMConfig, prompt string) (string, error) {
	body, _ := json.Marshal(ollamaRequest{
		Model:  cfg.Model,
		Prompt: prompt,
		Stream: false,
	})

	client := &http.Client{Timeout: llmTimeout(180 * time.Second)}
	resp, err := client.Post(cfg.BaseURL+"/api/generate", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("ollama request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("ollama read failed: %w", err)
	}

	var result ollamaResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("ollama parse failed: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("ollama error: %s", result.Error)
	}
	return strings.TrimSpace(result.Response), nil
}

// ── Groq (OpenAI-compatible) ──────────────────────────────────────────────────

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRequest struct {
	Model       string          `json:"model"`
	Messages    []openAIMessage `json:"messages"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
}

type openAIChoice struct {
	Message openAIMessage `json:"message"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func callGroq(cfg LLMConfig, prompt string) (string, error) {
	if cfg.APIKey == "" {
		return "", fmt.Errorf("GROQ_API_KEY is not set")
	}

	reqBody, _ := json.Marshal(openAIRequest{
		Model: cfg.Model,
		Messages: []openAIMessage{
			{Role: "system", Content: "You are a Kubernetes expert. Analyze the provided pod data and return a concise, actionable diagnosis in 2-4 sentences. Focus on the root cause and the most important fix."},
			{Role: "user", Content: prompt},
		},
		MaxTokens:   300,
		Temperature: 0.3,
	})

	req, _ := http.NewRequest(http.MethodPost, cfg.BaseURL+"/openai/v1/chat/completions", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: llmTimeout(60 * time.Second)}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("groq read failed: %w", err)
	}

	var result openAIResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return "", fmt.Errorf("groq parse failed: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("groq API error: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

func llmTimeout(defaultTimeout time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv("KUBEAID_AI_TIMEOUT_SECONDS"))
	if value == "" {
		return defaultTimeout
	}
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return defaultTimeout
	}
	seconds = min(seconds, math.MaxInt32)
	return time.Duration(seconds) * time.Second
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ── Prompt builder ────────────────────────────────────────────────────────────

// BuildAnalysisPrompt constructs a focused prompt from pod data.
func BuildAnalysisPrompt(appName, namespace, podName, status string, restarts int32, logs, events string) string {
	// Truncate logs and events to avoid exceeding context limits
	logsSnip := truncate(logs, 1500)
	eventsSnip := truncate(events, 500)

	return fmt.Sprintf(`You are a Kubernetes SRE. Analyze the following pod data and provide a concise root-cause diagnosis with the top 2-3 actionable fixes. Be specific.

--- POD INFO ---
App:       %s
Namespace: %s
Pod:       %s
Status:    %s
Restarts:  %d

--- EVENTS ---
%s

--- RECENT LOGS ---
%s

Respond in plain text. Start with "Confidence: high|medium|low". No markdown headers. Max 5 sentences.`,
		appName, namespace, podName, status, restarts, eventsSnip, logsSnip)
}

func truncate(s string, maxChars int) string {
	if len(s) <= maxChars {
		return s
	}
	// Keep the tail (most recent output is more useful)
	return "...(truncated)...\n" + s[len(s)-maxChars:]
}
