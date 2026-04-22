package security

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"time"
)

// TLSConfig holds secure TLS configuration
type TLSConfig struct {
	MinVersion         uint16
	CipherSuites       []uint16
	InsecureSkipVerify bool
	CAFile             string
}

// GetDefaultTLSConfig returns a secure TLS configuration with best practices
func GetDefaultTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		MaxVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
		InsecureSkipVerify: false,
	}
}

// ValidateCertificate validates a certificate
func ValidateCertificate(certPath string) error {
	if certPath == "" {
		return fmt.Errorf("certificate path cannot be empty")
	}

	// Check if file exists
	if _, err := os.Stat(certPath); err != nil {
		return fmt.Errorf("certificate file not found: %w", err)
	}

	// Read certificate
	data, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}

	// Parse certificate
	cert, err := x509.ParseCertificate(data)
	if err == nil {
		return validateCertificateValidity(cert)
	}

	// Try parsing as PEM
	block := parsePEM(data)
	if block == nil {
		return fmt.Errorf("unable to parse certificate (not PEM format)")
	}

	cert, err = x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	return validateCertificateValidity(cert)
}

// ValidateCertificateExpiration checks if certificate will expire soon
func ValidateCertificateExpiration(certPath string, warningDays int) error {
	data, err := os.ReadFile(certPath)
	if err != nil {
		return fmt.Errorf("failed to read certificate: %w", err)
	}

	block := parsePEM(data)
	if block == nil {
		return fmt.Errorf("unable to parse certificate (not PEM format)")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	expiryWarning := time.Now().AddDate(0, 0, warningDays)
	if cert.NotAfter.Before(expiryWarning) {
		daysTillExpiry := time.Until(cert.NotAfter).Hours() / 24
		return fmt.Errorf("certificate will expire in %.0f days", daysTillExpiry)
	}

	return nil
}

// ValidateCABundle validates a CA bundle file
func ValidateCABundle(caPath string) error {
	if caPath == "" {
		return fmt.Errorf("CA bundle path cannot be empty")
	}

	if _, err := os.Stat(caPath); err != nil {
		return fmt.Errorf("CA bundle file not found: %w", err)
	}

	data, err := os.ReadFile(caPath)
	if err != nil {
		return fmt.Errorf("failed to read CA bundle: %w", err)
	}

	certs, err := x509.ParseCertificates(data)
	if err != nil {
		return fmt.Errorf("failed to parse CA bundle: %w", err)
	}

	if len(certs) == 0 {
		return fmt.Errorf("CA bundle contains no certificates")
	}

	return nil
}

// validateCertificateValidity checks if certificate is within valid time period
func validateCertificateValidity(cert *x509.Certificate) error {
	now := time.Now()

	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid (valid from %v)", cert.NotBefore)
	}

	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate has expired (expired at %v)", cert.NotAfter)
	}

	// Warn if expiring soon (within 30 days)
	warningTime := time.Now().AddDate(0, 0, 30)
	if cert.NotAfter.Before(warningTime) {
		daysTillExpiry := time.Until(cert.NotAfter).Hours() / 24
		return fmt.Errorf("warning: certificate expires soon (in %.0f days)", daysTillExpiry)
	}

	return nil
}

// parsePEM extracts a single PEM block from data
func parsePEM(data []byte) *pem.Block {
	block, _ := pem.Decode(data)
	return block
}

// ValidateTLSConnection validates a TLS connection to a server
func ValidateTLSConnection(host string, port string) error {
	tlsConfig := GetDefaultTLSConfig()

	addr := host + ":" + port
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to establish TLS connection to %s: %w", addr, err)
	}
	defer conn.Close()

	// Get peer certificate chain
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return fmt.Errorf("no peer certificates received from %s", addr)
	}

	// Validate first certificate
	return validateCertificateValidity(certs[0])
}

// IsInsecureSkipVerifyEnabled checks if insecure skip verify is enabled
func IsInsecureSkipVerifyEnabled() bool {
	// Check environment variable
	skipVerify := os.Getenv("KUBECONFIG_INSECURE_SKIP_VERIFY")
	return skipVerify == "true" || skipVerify == "1"
}

// WarnIfInsecureSkipVerify logs a warning if insecure skip verify is enabled
func WarnIfInsecureSkipVerify() {
	if IsInsecureSkipVerifyEnabled() {
		fmt.Fprintf(os.Stderr, "⚠️  WARNING: Certificate verification is disabled (KUBECONFIG_INSECURE_SKIP_VERIFY=true)\n")
	}
}
