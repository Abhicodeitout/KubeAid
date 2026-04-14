
# kube-debugger

[![Go Version](https://img.shields.io/badge/Go-1.21-blue.svg)](https://golang.org)
[![Build](https://img.shields.io/badge/build-passing-brightgreen)]()
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://makeapullrequest.com)

> 🚀 Smart Kubernetes Debug CLI – Diagnose and fix pod issues instantly!

---


## Features
- 🔍 Analyze pod status, logs, events, and resource usage in one command
- 🤖 Smart diagnosis and fix suggestions
- 🧠 AI-powered log analysis and hints
- 🖥️ Interactive TUI mode
- 🌐 Multi-cluster/context support
- 📄 Export debug reports
- 🧩 Shell completion for bash/zsh/fish/powershell
- 📦 Easy install via Homebrew or Go
- 🏷️ Professional CLI with versioning


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


---

## Usage

### Analyze an app
```sh
kube-debugger analyze my-app
```

### Export a debug report
```sh
kube-debugger report my-app > report.txt
```

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

## Comparison
| Feature                | kube-debugger | kubectl describe |
|------------------------|:-------------:|:----------------:|
| Smart suggestions      | ✅            | ❌               |
| Resource usage         | ✅            | ❌               |
| TUI mode               | ✅            | ❌               |
| Multi-cluster support  | ✅            | ❌               |
| Shell completion       | ✅            | ❌               |
| Export report          | ✅            | ❌               |

---

## Contributing
PRs welcome! Please open issues for feature requests or bugs.

---

## License
Apache-2.0
