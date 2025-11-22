package http

import (
	"context"
	"crypto/tls"
	"net/http"
	"testing"
)

func TestWithServerCertificateReturnsErrorIfServerIsNil(t *testing.T) {
	ctx := context.Background()
	opt := WithServerCertificate(ctx, "cert.pem", "key.pem")
	err := opt(&httpServer{})
	if err == nil || err.Error() != "http server is not initialized" {
		t.Errorf("expected error for uninitialized server, got %v", err)
	}
}

func TestWithServerCertificateReturnsErrorIfCertOrKeyMissing(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithServerCertificate(ctx, "", "")
	err := opt(srv)
	if err == nil || err.Error() != "certFile and keyfile must be provided for TLS configuration" {
		t.Errorf("expected error for missing cert/key, got %v", err)
	}
}

func TestWithTLSVersionSetsMinAndMaxVersion(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithTLSVersion(ctx, tls.VersionTLS12, tls.VersionTLS13)
	err := opt(srv)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if srv.TLSConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinVersion to be set")
	}
	if srv.TLSConfig.MaxVersion != tls.VersionTLS13 {
		t.Errorf("expected MaxVersion to be set")
	}
}

func TestWithTLSVersionReturnsErrorIfMinGreaterThanMax(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithTLSVersion(ctx, tls.VersionTLS13, tls.VersionTLS12)
	err := opt(srv)
	if err == nil || err.Error() != "minVersion cannot be greater than maxVersion" {
		t.Errorf("expected error for minVersion > maxVersion, got %v", err)
	}
}

func TestWithCipherSuitesReturnsErrorIfNoSuitesProvided(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithCipherSuites(ctx, false)
	err := opt(srv)
	if err == nil || err.Error() != "at least one cipher suite must be provided" {
		t.Errorf("expected error for no cipher suites, got %v", err)
	}
}

func TestWithCipherSuitesReturnsErrorForUnknownSuite(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithCipherSuites(ctx, false, "UNKNOWN_SUITE")
	err := opt(srv)
	if err == nil || err.Error() != "unknown cipher suite: UNKNOWN_SUITE" {
		t.Errorf("expected error for unknown cipher suite, got %v", err)
	}
}

func TestWithCipherSuitesAddsValidCipherSuites(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	// Use a known cipher suite from the map
	var knownSuite string
	for name := range cipherSuiteMap {
		knownSuite = name
		break
	}
	opt := WithCipherSuites(ctx, false, knownSuite)
	err := opt(srv)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(srv.TLSConfig.CipherSuites) == 0 {
		t.Errorf("expected cipher suite to be added")
	}
}
