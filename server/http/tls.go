package http

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"slices"
)

// cipherSuiteMap is a map of cipher suite names to their corresponding IDs.
var cipherSuiteMap = map[string]uint16{}
var insecureCipherSuites []uint16

func init() {
	// Initialize the cipherSuiteMap with the default cipher suites.
	for _, suite := range tls.CipherSuites() {
		cipherSuiteMap[suite.Name] = suite.ID
	}
	for _, suite := range tls.InsecureCipherSuites() {
		cipherSuiteMap[suite.Name] = suite.ID
		insecureCipherSuites = append(insecureCipherSuites, suite.ID)
	}
}

func WithServerCertificate(ctx context.Context, certFile, keyfile string) ServerOption {
	return func(s *httpServer) error {
		if s.Server == nil {
			return fmt.Errorf("http server is not initialized")
		}
		if certFile == "" || keyfile == "" {
			return fmt.Errorf("certFile and keyfile must be provided for TLS configuration")
		}
		cert, err := tls.LoadX509KeyPair(certFile, keyfile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificate and key: %w", err)
		}
		if s.TLSConfig == nil {
			s.TLSConfig = &tls.Config{}
		}
		if s.TLSConfig.Certificates == nil {
			s.TLSConfig.Certificates = []tls.Certificate{cert}
		} else {
			// If there are already certificates, append the new one
			s.TLSConfig.Certificates = append(s.TLSConfig.Certificates, cert)
		}
		return nil
	}
}

// WithTLSVersion sets the minimum and maximum TLS versions for the server.
// pass 0 to skip setting the value
func WithTLSVersion(ctx context.Context, minVersion, maxVersion uint16) ServerOption {
	return func(s *httpServer) error {
		if minVersion > 0 && maxVersion > 0 && minVersion > maxVersion {
			return fmt.Errorf("minVersion cannot be greater than maxVersion")
		}
		if s.TLSConfig == nil {
			s.TLSConfig = &tls.Config{}
		}
		if minVersion > 0 {
			s.TLSConfig.MinVersion = minVersion
		}
		if maxVersion > 0 {
			s.TLSConfig.MaxVersion = maxVersion
		}
		return nil
	}
}

// WithCipherSuites sets the cipher suites for the server which overrides the default ones.
// reference: https://go.dev/src/crypto/tls/cipher_suites.go
func WithCipherSuites(ctx context.Context, skipInsecure bool, cipherSuites ...string) ServerOption {
	return func(s *httpServer) error {
		if s.TLSConfig == nil {
			s.TLSConfig = &tls.Config{}
		}
		if len(cipherSuites) == 0 {
			return fmt.Errorf("at least one cipher suite must be provided")
		}
		for _, cipherSuite := range cipherSuites {
			cipherID, ok := cipherSuiteMap[cipherSuite]
			if !ok {
				return fmt.Errorf("unknown cipher suite: %s", cipherSuite)
			}
			if skipInsecure && slices.Contains(insecureCipherSuites, cipherID) {
				slog.Warn("Skipping insecure cipher suite %s for TLS configuration of server %v", cipherSuite, s)
				continue
			}
			s.TLSConfig.CipherSuites = append(s.TLSConfig.CipherSuites, cipherID)
		}
		return nil
	}
}
