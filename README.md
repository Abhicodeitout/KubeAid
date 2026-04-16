
# KubeAid

[![Go Version](https://img.shields.io/badge/Go-1.26-blue.svg)](https://golang.org)
[![Build](https://img.shields.io/badge/build-passing-brightgreen)]()
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://makeapullrequest.com)

> 🚀 Smart Kubernetes Debug CLI – Diagnose and fix pod issues instantly!

---


## Features
- 🔍 Analyze pod status, logs, events, and resource usage in one command
- 🤖 Smart diagnosis with fix suggestions for 9+ failure states (CrashLoopBackOff, OOMKilled, Evicted, Terminating, and more)
- 🧠 AI-powered analysis — pluggable **real LLM** (Ollama local or Groq cloud free tier) with 10-pattern fallback
- 📄 Export debug reports as **text**, **JSON**, or **HTML** with `--format` / `--output` flags
- 🖥️ Interactive TUI mode
- 🌐 Multi-cluster/context support
- 🧩 Shell completion for bash/zsh/fish/powershell
- 📦 Easy install via Homebrew or Go
- 🏷️ Professional CLI with versioning


---

## AI Configuration

KubeAid supports real LLM analysis via two **free** providers. Set environment variables before running any command.

By default, this repo already includes a pre-configured env file at `kube-debugger/env/kube-debugger.env`.
`make bootstrap` auto-loads this file, so users do not need to export variables manually.

### Option 1 — Ollama (local, no API key, fully private)

Fastest path with built-in OS detection:

```sh
cd kube-debugger
make bootstrap SKIP_KUBECONFIG_CHECK=true

# Optional: if running binary directly without make, source the env file
source env/kube-debugger.env
```

[Install Ollama](https://ollama.com) and pull a model, then set:

```sh
ollama pull llama3.2

export KUBEAID_AI_PROVIDER=ollama
export KUBEAID_AI_MODEL=llama3.2        # optional, default: llama3.2
export KUBEAID_OLLAMA_URL=http://localhost:11434  # optional, this is the default

kube-debugger analyze my-app
```

### Option 2 — Groq (cloud, free tier, no credit card)

1. Sign up at [console.groq.com](https://console.groq.com) → create a free API key  
2. Set:

```sh
export KUBEAID_AI_PROVIDER=groq
export GROQ_API_KEY=gsk_xxxxxxxxxxxx
export KUBEAID_AI_MODEL=llama3-8b-8192  # optional, default: llama3-8b-8192

kube-debugger analyze my-app
```

### Fallback behaviour

If `KUBEAID_AI_PROVIDER` is not set, or the LLM call fails, kube-debugger automatically falls back to fast built-in pattern matching — no configuration required and no errors shown.

| Variable | Default | Description |
|---|---|---|
| `KUBEAID_AI_PROVIDER` | _(unset)_ | `ollama` or `groq` |
| `KUBEAID_AI_MODEL` | provider default | Model name override |
| `KUBEAID_OLLAMA_URL` | `http://localhost:11434` | Custom Ollama endpoint |
| `KUBEAID_AI_TIMEOUT_SECONDS` | provider default | Override AI request timeout for larger prompts or slower local hardware |
| `GROQ_API_KEY` | _(unset)_ | Groq API key |

---

## Install


### Homebrew (recommended)
```sh
brew tap abhicodeitout/tap
brew install kube-debugger
```

### Go
```sh
GO111MODULE=on go install github.com/Abhicodeitout/KubeAid/kube-debugger@latest
```

---

## Run

After installation, run any command from your terminal:

```sh
kube-debugger --help
```

---

## Available Commands

| Command                                 | Description                                 |
|-----------------------------------------|---------------------------------------------|
| kube-debugger analyze [app-name]        | Analyze a Kubernetes app for issues         |
| kube-debugger report [app-name]         | Export debug report for an app              |
| kube-debugger tui                       | Launch interactive terminal UI              |
| kube-debugger context list              | List available kubeconfig contexts          |
| kube-debugger context switch [context]  | Switch to a different kubeconfig context    |
| kube-debugger completion [shell]        | Generate shell completion scripts           |
| kube-debugger version                   | Show kube-debugger version                  |
| kube-debugger --help                    | Show help for all commands                  |
| kube-debugger bootstrap                 | Run cross-platform environment pre-checks   |

Full command examples are available in [kube-debugger/COMMANDS.md](kube-debugger/COMMANDS.md).


---

## Usage

### Analyze an app
```sh
kube-debugger analyze my-app
```

### Export a debug report

The `report` command supports multiple output formats and can write directly to a file.

```sh
# Plain text to stdout (default)
kube-debugger report my-app

# JSON report saved to a file
kube-debugger report my-app --format json --output report.json

# HTML report saved to a file
kube-debugger report my-app --format html --output report.html
```

| Flag | Short | Values | Default | Description |
|------|-------|--------|---------|-------------|
| `--format` | `-f` | `text`, `json`, `html` | `text` | Output format |
| `--output` | `-o` | file path | stdout | Write report to this file |

### Interactive TUI (coming soon)
```sh
kube-debugger tui
```

### List/switch Kubernetes contexts
```sh
kube-debugger context list
kube-debugger context switch my-context
```

### Shell completion
```sh
kube-debugger completion bash
kube-debugger completion zsh
```

### Show version
```sh
kube-debugger version
```

### Bootstrap (Pre-check)
```sh
kube-debugger bootstrap
kube-debugger bootstrap --install-ollama --start-ollama --pull-model llama3.2 --skip-kubeconfig-check
```
Runs environment pre-checks (Go version, kubeconfig presence) and can install/start Ollama automatically on Linux, macOS, and Windows.

| Flag | Description |
|---|---|
| `--install-ollama` | Install Ollama when it is missing |
| `--start-ollama` | Start Ollama if it is installed but not serving |
| `--pull-model llama3.2` | Pull a model after setup |
| `--skip-kubeconfig-check` | Skip kubeconfig validation for AI-only setup |

---

## Demo
![Demo GIF](demo.gif)

---

## Why kube-debugger?
- One command = full diagnosis
- Smarter than `kubectl describe`
- AI-like hints for common issues
- Multi-cluster ready
- Professional CLI experience

---

## Diagnostic Coverage

### AI Log Hints
Patterns detected automatically from pod logs:

| Pattern | Hint |
|---------|------|
| `connection refused`, `dial tcp`, `no such host` | Network connectivity issue — check service DNS and network policies |
| `oomkilled`, `out of memory` | OOM — increase memory limits or fix leaks |
| `imagepullbackoff`, `pull access denied` | Image pull failed — verify name, tag, and pull secret |
| `crashloopbackoff` | Crash loop — check previous logs and liveness probe settings |
| `probe failed`, `livenessprobe` | Health probe failing — check endpoint and timeout settings |
| `permission denied`, `forbidden` | RBAC or filesystem permission issue |
| `configmap`, `secret not found` | Missing ConfigMap or Secret in namespace |
| `timeout`, `context deadline exceeded` | Request timed out — check upstream service and timeouts |
| `evicted` | Node under resource pressure — review node conditions |
| `panic`, `fatal error`, `segfault` | Application crash — review stack trace |
| `certificate`, `tls`, `x509` | TLS/certificate error — check cert validity and CA bundles |

### Fix Suggestions by Status

| Pod Status | Key Suggestions |
|---|---|
| `CrashLoopBackOff` | Check previous logs, env vars, secrets; contextual OOM/network hints |
| `OOMKilled` | Increase memory limits, profile leaks, consider VPA |
| `ImagePullBackOff` / `ErrImagePull` | Verify image name/tag, pull secret, registry access |
| `Pending` | Check node resources, taints, affinity rules, PVC binding |
| `Evicted` | Node pressure (disk/memory/PID), set resource requests, use PodDisruptionBudget |
| `Terminating` | Check for stuck finalizers, force-delete |
| `ContainerCreating` | Check volume mounts, ConfigMap/Secret existence |
| `RunContainerError` | Review entrypoint/command in pod spec |
| Probe failures | Tune probe endpoint, port, and initial delay |

---

## Comparison
| Feature                        | kube-debugger | kubectl describe |
|--------------------------------|:-------------:|:----------------:|
| Smart fix suggestions          | ✅            | ❌               |
| 9+ failure state coverage      | ✅            | ❌               |
| AI log pattern detection       | ✅            | ❌               |
| JSON / HTML report export      | ✅            | ❌               |
| Resource usage                 | ✅            | ❌               |
| TUI mode                       | ✅            | ❌               |
| Multi-cluster support          | ✅            | ❌               |
| Shell completion               | ✅            | ❌               |

---

## Contributing
PRs welcome! Please open issues for feature requests or bugs.

---

## License
Apache-2.0
