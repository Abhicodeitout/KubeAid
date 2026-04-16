APP_NAME=kube-debugger
OLLAMA_MODEL?=llama3.2
SKIP_KUBECONFIG_CHECK?=false
ENV_FILE?=env/kube-debugger.env

.PHONY: build test run deps-check bootstrap build-bootstrap

build:
	go build -o $(APP_NAME) ./main.go

test:
	go test ./...

run:
	@set -a; \
	if [ -f "$(ENV_FILE)" ]; then . "$(ENV_FILE)"; fi; \
	set +a; \
	go run ./main.go

deps-check:
	@echo "Checking required tooling..."
	@command -v go >/dev/null || (echo "Missing dependency: go" && exit 1)
	@command -v curl >/dev/null || (echo "Missing dependency: curl (required for Ollama install on Linux/macOS)" && exit 1)
	@echo "Required tooling looks good."
	@echo "Optional tooling checks (warnings only):"
	@command -v kubectl >/dev/null || echo "  - kubectl not found (needed for cluster analysis commands)"
	@command -v minikube >/dev/null || echo "  - minikube not found (optional local cluster)"
	@command -v docker >/dev/null || echo "  - docker not found (optional, used by minikube/kind setups)"

bootstrap: build deps-check
	@echo "Running kube-debugger bootstrap with Ollama setup..."
	@set -a; \
	if [ -f "$(ENV_FILE)" ]; then \
		. "$(ENV_FILE)"; \
		echo "Loaded env from $(ENV_FILE)"; \
	else \
		echo "Env file not found at $(ENV_FILE); using shell environment only"; \
	fi; \
	set +a; \
	flags="--install-ollama --start-ollama --pull-model $${KUBEAID_AI_MODEL:-$(OLLAMA_MODEL)}"; \
	if [ "$(SKIP_KUBECONFIG_CHECK)" = "true" ]; then \
		flags="$$flags --skip-kubeconfig-check"; \
	fi; \
	./$(APP_NAME) bootstrap $$flags

build-bootstrap: bootstrap
