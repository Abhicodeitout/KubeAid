# KubeAid Copilot Features

KubeAid now includes **GitHub Copilot-like** AI-powered troubleshooting suggestions to help you fix Kubernetes issues faster.

## What's New

### 🤖 Copilot-Powered Suggestions

When you run `kube-debugger analyze`, it now provides:

1. **Structured Problem Analysis**
   - Root cause identification
   - Severity levels (critical/warning/info)
   - Clear descriptions of what's wrong

2. **Step-by-Step Fix Instructions**
   - Numbered steps to resolve the issue
   - Exact commands to run
   - YAML configuration fixes

3. **Quick Commands**
   - Copy-paste ready kubectl commands
   - Targeted debugging commands
   - Real-time monitoring commands

4. **Learning Resources**
   - Links to Kubernetes documentation
   - Best practice recommendations

## Features

### Pattern-Based Intelligence

Out of the box, KubeAid detects and provides specific fixes for:

- ✅ **Network Connectivity** - Detects connection refused, DNS issues
- ✅ **Out of Memory (OOMKilled)** - Memory limit suggestions
- ✅ **Image Pull Failures** - Registry and credential fixes
- ✅ **CrashLoopBackOff** - Debug steps and fix patterns
- ✅ **Failed Health Probes** - Probe configuration fixes
- ✅ **Permission Errors** - RBAC troubleshooting
- ✅ **Timeout Issues** - Network latency diagnosis
- ✅ **Pod Evictions** - Node pressure analysis
- ✅ **TLS/Certificate Errors** - Certificate debugging

### AI Provider Integration

Enhance with real AI (Optional):

```bash
# Set up Ollama (local AI)
export KUBEAID_AI_PROVIDER=ollama
export KUBEAID_OLLAMA_URL=http://localhost:11434

# Or use Groq (cloud AI)
export KUBEAID_AI_PROVIDER=groq
export GROQ_API_KEY=your-api-key
```

With AI enabled, KubeAid provides even deeper contextual analysis and suggestions.

## Examples

### Example 1: CrashLoopBackOff

```bash
$ kube-debugger analyze my-app -n default
```

Output includes:

```
🤖 COPILOT FIX
## 🔄 CrashLoopBackOff [critical]

**Description:** Pod is continuously crashing and restarting

**Fix Steps:**
- 1. Check current logs: kubectl logs my-app -n default
- 2. Check previous logs: kubectl logs my-app -n default --previous
- 3. Check pod events: kubectl describe pod my-app -n default
- 4. Fix the root cause (typically: app error, missing config, bad probe settings)
- 5. Monitor restarts: watch kubectl get pods -n default

**Quick Commands:**
```bash
kubectl logs my-app -n default --previous
```

**Learn More:**
- https://kubernetes.io/docs/tasks/debug-application-cluster/debug-running-pod/
```

### Example 2: Out of Memory (OOMKilled)

```bash
$ kube-debugger analyze memory-heavy-app -n production
```

Output includes:

```
🤖 COPILOT FIX
## 💾 Out of Memory (OOMKilled) [critical]

**Description:** Container killed due to exceeding memory limit

**Fix Steps:**
- 1. Check current memory limit: kubectl get pod memory-heavy-app -n production -o yaml | grep -A2 resources
- 2. Monitor memory usage: kubectl top pod memory-heavy-app -n production
- 3. Identify memory leak patterns in logs
- 4. Increase memory limit or fix memory leak

**YAML Configuration Fix:**
```yaml
resources:
  limits:
    memory: "1Gi"  # Increase this value
  requests:
    memory: "512Mi"
```

**Quick Commands:**
```bash
kubectl set resources pod memory-heavy-app --limits=memory=1Gi -n production
```
```

### Example 3: Network Connectivity Issue

```bash
$ kube-debugger analyze db-client -n default
```

Output includes:

```
🤖 COPILOT FIX
## 🔗 Network Connectivity Issue [critical]

**Description:** Pod cannot establish connection to a service or external resource

**Fix Steps:**
- 1. Verify the target service is running: kubectl get svc -n default
- 2. Test DNS resolution: kubectl exec -it db-client -n default -- nslookup [service-name]
- 3. Check network policies: kubectl get networkpolicies -n default
- 4. Verify service endpoints: kubectl get endpoints -n default

**Quick Commands:**
```bash
kubectl describe svc [service-name] -n default
kubectl logs db-client -n default | grep -i connection
kubectl get networkpolicies -n default -o yaml
```

**Learn More:**
- https://kubernetes.io/docs/tasks/debug-application-cluster/debug-service/
```

## How It Works

### 1. **Pattern Detection** (Always Active)
   - KubeAid analyzes logs and events for known error patterns
   - Provides specific fix for each detected pattern
   - No external dependencies required

### 2. **AI Enhancement** (Optional)
   - If AI provider is configured, sends comprehensive prompt to LLM
   - Gets context-aware, detailed suggestions
   - Falls back to pattern detection if AI fails

### 3. **Structured Output**
   - Problem title and severity
   - Step-by-step instructions
   - Copy-paste ready commands
   - YAML configuration examples
   - Learning resources

## Setting Up

### Basic Usage (No AI Required)

```bash
cd kube-debugger
make build
./kube-debugger analyze my-app -n default
```

You immediately get pattern-based Copilot suggestions!

### With Ollama (Local AI)

```bash
# Install Ollama from https://ollama.ai
ollama pull llama3.2

# Start Ollama server
ollama serve

# In another terminal:
export KUBEAID_AI_PROVIDER=ollama
export KUBEAID_OLLAMA_URL=http://localhost:11434
cd kube-debugger
./kube-debugger analyze my-app -n default
```

### With Groq (Cloud AI - Free)

```bash
# Get API key from https://console.groq.com
export KUBEAID_AI_PROVIDER=groq
export GROQ_API_KEY=gsk_...

cd kube-debugger
./kube-debugger analyze my-app -n default
```

## Output Sections

When you run `kube-debugger analyze`, you'll see:

```
📦 POD OVERVIEW          - Basic pod information
🏥 HEALTH SCORE          - 0-100 health rating
🔍 AI ANALYSIS           - Quick AI hint
🤖 COPILOT FIX          - Detailed structured suggestions
✅ SUGGESTIONS           - Additional recommendations
📋 EVENTS               - Kubernetes events
📄 LOGS                 - Last log lines
```

## Benefits Over Manual Troubleshooting

| Manual | Copilot |
|--------|---------|
| Manually search logs | Automatic pattern detection |
| Look up documentation | Instant relevant links |
| Remember kubectl syntax | Copy-paste ready commands |
| Try different approaches | Step-by-step guidance |
| Trial and error | Targeted fixes |

## Real-World Use Cases

### DevOps Team
- **Faster incident response** - Reduce MTTR with instant suggestions
- **Knowledge sharing** - New team members learn from Copilot analysis
- **Automation** - Integrate suggestions into incident workflows

### Developers
- **Local debugging** - Quick diagnosis during development
- **CI/CD integration** - Automatic analysis in pipelines
- **Learning** - Understand Kubernetes troubleshooting patterns

### Platform Teams
- **Cluster monitoring** - Regular health checks with suggestions
- **Runbook generation** - Auto-generate from Copilot outputs
- **Compliance** - Audit with structured analysis records

## Future Enhancements

- [ ] Interactive chat for deeper troubleshooting
- [ ] Custom rule creation for org-specific patterns
- [ ] Training on your cluster's specific issues
- [ ] Integration with incident management systems
- [ ] Predictive alerts based on patterns
- [ ] Cost optimization suggestions

## Troubleshooting Copilot

### AI suggestions not appearing?

1. Check if AI is configured:
   ```bash
   echo $KUBEAID_AI_PROVIDER
   ```

2. Pattern-based suggestions should still work (they're always available)

3. If using Ollama:
   ```bash
   curl http://localhost:11434/api/tags
   ```

4. If using Groq:
   ```bash
   curl -H "Authorization: Bearer $GROQ_API_KEY" \
     https://api.groq.com/openai/v1/models
   ```

### Getting generic suggestions?

- This is normal for unusual error patterns
- Provide detailed logs for better analysis
- Consider enabling AI provider for context-aware analysis

## Example Commands

```bash
# Get Copilot suggestions for an app
./kube-debugger analyze my-app -n production

# Export analysis with suggestions to JSON
./kube-debugger report my-app -f json -o analysis.json

# Watch for issues with continuous analysis
./kube-debugger analyze my-app --watch --interval 10

# Fail CI if Copilot detects critical issues
./kube-debugger analyze my-app --exit-code --threshold 70
```

## Contributing

Have a new pattern to detect? Have suggestions for improvement?

See [COMMANDS.md](COMMANDS.md) for more detailed command usage.
