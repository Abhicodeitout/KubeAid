package security

import (
	"fmt"
	"os"
	"sync"
)

// SecurityManager orchestrates all security features
type SecurityManager struct {
	auditLogger      *AuditLogger
	configHandler    *ConfigHandler
	rateLimiter      *RateLimiter
	operationLimiter *OperationLimiter
	outputFilter     *OutputFilter
}

var (
	globalSecurityManager *SecurityManager
	secMgrOnce            sync.Once
)

// InitSecurityManager initializes the global security manager
func InitSecurityManager(auditDir, configDir string, enableAudit, filterOutput bool) (*SecurityManager, error) {
	var err error
	secMgrOnce.Do(func() {
		// Initialize audit logger
		auditErr := InitAuditLogger(auditDir, enableAudit)
		if auditErr != nil {
			err = fmt.Errorf("failed to initialize audit logger: %w", auditErr)
			return
		}

		globalSecurityManager = &SecurityManager{
			auditLogger:      GetAuditLogger(),
			configHandler:    NewConfigHandler(configDir),
			rateLimiter:      GetGlobalRateLimiter(),
			operationLimiter: GetGlobalOperationLimiter(),
			outputFilter:     NewOutputFilter(true, false),
		}

		// Log initialization
		globalSecurityManager.auditLogger.LogSecurityEvent("security_manager_initialized", "info", map[string]interface{}{
			"audit_enabled": enableAudit,
			"filter_output": filterOutput,
		})
	})

	return globalSecurityManager, err
}

// GetSecurityManager returns the global security manager
func GetSecurityManager() *SecurityManager {
	if globalSecurityManager == nil {
		globalSecurityManager = &SecurityManager{
			auditLogger:      GetAuditLogger(),
			configHandler:    NewConfigHandler(os.TempDir()),
			rateLimiter:      GetGlobalRateLimiter(),
			operationLimiter: GetGlobalOperationLimiter(),
			outputFilter:     NewOutputFilter(true, false),
		}
	}
	return globalSecurityManager
}

// ValidateInput validates input before processing
func (sm *SecurityManager) ValidateInput(appName, namespace string) error {
	if err := ValidateAppName(appName); err != nil {
		sm.auditLogger.LogSecurityEvent("input_validation_failed", "warning", map[string]interface{}{
			"error": err.Error(),
			"app":   appName,
		})
		return err
	}

	if namespace != "" {
		if err := ValidateNamespace(namespace); err != nil {
			sm.auditLogger.LogSecurityEvent("input_validation_failed", "warning", map[string]interface{}{
				"error":     err.Error(),
				"namespace": namespace,
			})
			return err
		}
	}

	return nil
}

// LogCommand logs a command execution with security context
func (sm *SecurityManager) LogCommand(cmd string, args []string, appName, namespace string, err error) error {
	// Redact arguments
	redactedArgs := make([]string, len(args))
	for i, arg := range args {
		redactedArgs[i] = RedactSecrets(SanitizeInput(arg))
	}

	return sm.auditLogger.LogCommand(cmd, redactedArgs, appName, namespace, err)
}

// FilterOutput applies output filtering
func (sm *SecurityManager) FilterOutput(output string) string {
	return sm.outputFilter.FilterOutput(output)
}

// CheckConfigSecurity validates configuration security
func (sm *SecurityManager) CheckConfigSecurity() error {
	warnings := sm.configHandler.ValidateEnvironmentVariables()
	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "⚠️  Security Warning: %s\n", warning)
	}

	// Always try to securely handle kubeconfig
	_ = sm.configHandler.SecureKubeconfig()

	return nil
}

// EnforceRateLimit enforces rate limiting for an operation
func (sm *SecurityManager) EnforceRateLimit(operation string) error {
	sm.operationLimiter.Wait(operation)
	return nil
}

// AuditLog provides access to audit logger
func (sm *SecurityManager) AuditLog() *AuditLogger {
	return sm.auditLogger
}

// ConfigHandler provides access to config handler
func (sm *SecurityManager) Config() *ConfigHandler {
	return sm.configHandler
}

// PerformSecurityChecks performs all security checks at startup
func PerformSecurityChecks() {
	manager := GetSecurityManager()

	// Validate kubeconfig
	if err := manager.configHandler.ValidateKubeconfig(); err != nil {
		fmt.Fprintf(os.Stderr, "⚠️  Kubeconfig Warning: %v\n", err)
	}

	// Check environment variables
	warnings := manager.configHandler.ValidateEnvironmentVariables()
	for _, warning := range warnings {
		fmt.Fprintf(os.Stderr, "⚠️  Security Warning: %s\n", warning)
	}

	// Check TLS security
	WarnIfInsecureSkipVerify()

	// Log security context
	secContext := CheckSecurityContext()
	if runningAsRoot, ok := secContext["running_as_root"].(bool); ok && runningAsRoot {
		fmt.Fprintf(os.Stderr, "⚠️  Warning: Running as root\n")
	}
}
