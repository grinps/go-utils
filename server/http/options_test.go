package http

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestWithListenAddressSetsAddressCorrectly(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithListenAddress(ctx, "127.0.0.1", "8080")
	err := opt(srv)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if srv.Server.Addr != "127.0.0.1:8080" {
		t.Errorf("expected address to be set, got %v", srv.Server.Addr)
	}
}

func TestWithListenAddressReturnsErrorIfServerIsNil(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{}
	opt := WithListenAddress(ctx, "127.0.0.1", "8080")
	err := opt(srv)
	if err == nil || err.Error() != "http server is not initialized" {
		t.Errorf("expected error for uninitialized server, got %v", err)
	}
}

func TestWithTimeoutSetsIdleTimeout(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithTimeout(ctx, TimeOutIdle, 2*time.Second)
	err := opt(srv)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if srv.Server.IdleTimeout != 2*time.Second {
		t.Errorf("expected IdleTimeout to be set, got %v", srv.Server.IdleTimeout)
	}
}

func TestWithTimeoutReturnsErrorForUnknownTimeout(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithTimeout(ctx, TimeOut("unknown"), 1*time.Second)
	err := opt(srv)
	if err == nil || err.Error() != "unknown timeout type" {
		t.Errorf("expected error for unknown timeout, got %v", err)
	}
}

func TestWithTimeoutReturnsErrorIfServerIsNil(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{}
	opt := WithTimeout(ctx, TimeOutIdle, 1*time.Second)
	err := opt(srv)
	if err == nil || err.Error() != "http server is not initialized" {
		t.Errorf("expected error for uninitialized server, got %v", err)
	}
}

func TestWithWaitToStartSetsStartUpWait(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithWaitToStart(ctx, 3*time.Second)
	err := opt(srv)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if srv.StartUpWait != 3*time.Second {
		t.Errorf("expected StartUpWait to be set, got %v", srv.StartUpWait)
	}
}

func TestWithWaitToStartReturnsErrorIfServerIsNil(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{}
	opt := WithWaitToStart(ctx, 1*time.Second)
	err := opt(srv)
	if err == nil || err.Error() != "http server is not initialized" {
		t.Errorf("expected error for uninitialized server, got %v", err)
	}
}

func TestWithWaitToShutdownSetsShutdownWait(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithWaitToShutdown(ctx, 4*time.Second)
	err := opt(srv)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if srv.ShutdownWait != 4*time.Second {
		t.Errorf("expected ShutdownWait to be set, got %v", srv.ShutdownWait)
	}
}

func TestWithWaitToShutdownReturnsErrorIfServerIsNil(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{}
	opt := WithWaitToShutdown(ctx, 1*time.Second)
	err := opt(srv)
	if err == nil || err.Error() != "http server is not initialized" {
		t.Errorf("expected error for uninitialized server, got %v", err)
	}
}

func TestWithShutdownSignalsSetsSignals(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithShutdownSignals(ctx, os.Interrupt, os.Kill)
	err := opt(srv)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if len(srv.ShutdownSignals) != 2 {
		t.Errorf("expected 2 shutdown signals, got %v", len(srv.ShutdownSignals))
	}
}

func TestWithShutdownSignalsReturnsErrorIfServerIsNil(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{}
	opt := WithShutdownSignals(ctx, os.Interrupt)
	err := opt(srv)
	if err == nil || err.Error() != "http server is not initialized" {
		t.Errorf("expected error for uninitialized server, got %v", err)
	}
}

func TestWithProtocolsSetsHTTP1AndHTTP2Correctly(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithProtocols(ctx, string(ProtocolHTTP1), string(ProtocolHTTP2))
	err := opt(srv)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if srv.Server.Protocols == nil {
		t.Errorf("expected Protocols to be set, got nil")
	}
}

func TestWithProtocolsReturnsErrorForUnknownProtocol(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{Server: &http.Server{}}
	opt := WithProtocols(ctx, "unknown")
	err := opt(srv)
	if err == nil || err.Error() != "unknown protocol type" {
		t.Errorf("expected error for unknown protocol, got %v", err)
	}
}

func TestWithProtocolsReturnsErrorIfServerIsNil(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{}
	opt := WithProtocols(ctx, string(ProtocolHTTP1))
	err := opt(srv)
	if err == nil || err.Error() != "http server is not initialized" {
		t.Errorf("expected error for uninitialized server, got %v", err)
	}
}
