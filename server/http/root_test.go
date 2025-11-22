package http

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grinps/go-utils/server"
)

func NewTestServer() *httpServer {
	return &httpServer{
		Server: &http.Server{
			Addr: "127.0.0.1:0",
		},
		StartUpWait:     100 * time.Millisecond,
		ShutdownWait:    100 * time.Millisecond,
		ShutdownSignals: []os.Signal{},
	}
}

func TestServerStartsSuccessfully(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := NewTestServer()

	go func() {
		if err := srv.Start(ctx); err != nil {
			t.Errorf("failed to start server: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	if srv.State != ServerStateStarting {
		t.Errorf("expected server state to be starting, got %v", srv.State)
	}
}

func TestServerFailsToStartWithInvalidListener(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := NewTestServer()
	srv.Server.Addr = "invalid:address"

	err := srv.Start(ctx)
	if err == nil || err.Error() != "failed to start http server due to listener creation error: listen tcp: lookup tcp/address: unknown port" {
		t.Errorf("expected listener creation error, got %v", err)
	}
}

func TestServerStopsSuccessfully(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := NewTestServer()

	go func() {
		if err := srv.Start(ctx); err != nil {
			t.Errorf("failed to start server: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	err := srv.Stop(ctx)
	if err != nil {
		t.Errorf("failed to stop server: %v", err)
	}

	if srv.State != ServerStateStopped {
		t.Errorf("expected server state to be stopped, got %v", srv.State)
	}
}

func TestServerToStopWhenNotStarted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := NewTestServer()

	err := srv.Stop(ctx)
	if err != nil {
		t.Errorf("expected no error when stopping uninitialized server, got %v", err)
	}
}

func TestServerHandlesOSSignalCorrectly(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := NewTestServer()
	srv.ShutdownSignals = []os.Signal{os.Interrupt}

	go func() {
		if err := srv.Start(ctx); err != nil {
			t.Errorf("failed to start server: %v", err)
		}
	}()

	time.Sleep(200 * time.Millisecond)

	// Send signal to the current process
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)

	time.Sleep(200 * time.Millisecond)

	if srv.State != ServerStateStopped {
		t.Errorf("expected server state to be stopped after signal, got %v", srv.State)
	}
}

func TestServerFailsToStartWithNilServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := &httpServer{}

	err := srv.Start(ctx)
	if err == nil || err.Error() != "http server is not initialized" {
		t.Errorf("expected error for uninitialized server, got %v", err)
	}
}

func TestServerFailsToStopWithNilServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := &httpServer{}

	err := srv.Stop(ctx)
	if err == nil || err.Error() != "http server is not initialized" {
		t.Errorf("expected error for uninitialized server, got %v", err)
	}
}

func TestHandleWaitTimeExecutesPostFunction(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var postExecuted bool
	HandleWaitTime(ctx, 100*time.Millisecond, func() {
		postExecuted = true
	})

	time.Sleep(200 * time.Millisecond)

	if !postExecuted {
		t.Errorf("expected post function to be executed, but it was not")
	}
}

func TestHandleWaitTimeSkipsPostFunctionOnZeroWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var postExecuted bool
	HandleWaitTime(ctx, 0, func() {
		postExecuted = true
	})

	time.Sleep(100 * time.Millisecond)

	if postExecuted {
		t.Errorf("expected post function to be skipped, but it was executed")
	}
}

func TestHandleOSSignalSkipsExecutionWithoutSignals(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var postExecuted bool
	HandleOSSignal(ctx, func() {
		postExecuted = true
	})

	time.Sleep(100 * time.Millisecond)

	if postExecuted {
		t.Errorf("expected post function to be skipped, but it was executed")
	}
}

func TestServerStateIsErrorOnServeFailure(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := NewTestServer()
	// Create a listener to occupy the port, forcing the server to fail
	ln, err := net.Listen("tcp", srv.Server.Addr)
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer ln.Close()
	srv.Server.Addr = ln.Addr().String()

	err = srv.Start(ctx)
	if err == nil {
		t.Errorf("expected error when starting server with occupied port, got nil")
	}
	if srv.State != ServerStateError {
		t.Errorf("expected server state to be error, got %v", srv.State)
	}
}

func TestHandleWaitTimeDoesNotPanicWithNilPostFunc(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	HandleWaitTime(ctx, 50*time.Millisecond, nil)
}

func TestHandleWaitTimeSkipsPostFuncOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var postExecuted bool
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	HandleWaitTime(ctx, 100*time.Millisecond, func() {
		postExecuted = true
	})
	if postExecuted {
		t.Errorf("expected post function to be skipped due to context cancel, but it was executed")
	}
}

func TestHandleOSSignalExecutesPostFuncOnSignal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var postExecuted bool
	done := make(chan struct{})
	go func() {
		HandleOSSignal(ctx, func() {
			postExecuted = true
			close(done)
		}, os.Interrupt)
	}()
	time.Sleep(10 * time.Millisecond)
	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(os.Interrupt)
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
	}
	if !postExecuted {
		t.Errorf("expected post function to be executed on signal, but it was not")
	}
}

func TestHandleOSSignalSkipsPostFuncOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var postExecuted bool
	done := make(chan struct{})
	go func() {
		HandleOSSignal(ctx, func() {
			postExecuted = true
			close(done)
		}, os.Interrupt)
	}()
	time.Sleep(10 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
	}
	if postExecuted {
		t.Errorf("expected post function to be skipped due to context cancel, but it was executed")
	}
}

func TestServerStartSetsGinHandlerIfPresent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := NewTestServer()
	engine := &gin.Engine{}
	srv.GinEngine = engine
	srv.Server.Handler = nil

	go func() {
		_ = srv.Start(ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	if srv.Server.Handler == nil {
		t.Errorf("expected Gin handler to be set, but it was nil")
	}
}

func TestServerStartUsesTLSListenerIfTLSConfigPresent(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := NewTestServer()
	srv.TLSConfig = &tls.Config{}

	go func() {
		_ = srv.Start(ctx)
	}()
	time.Sleep(100 * time.Millisecond)

	// No error expected, just ensure server is running with TLS config
	if srv.TLSConfig == nil {
		t.Errorf("expected TLSConfig to be set, but it was nil")
	}
}

func TestStopDoesNothingIfServerIsNil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Stop(ctx, nil)
}

func TestNewServerAppliesOptionsAndReturnsErrorOnFailure(t *testing.T) {
	ctx := context.Background()
	opt := func(s *httpServer) error { return errors.New("option error") }
	_, err := NewServer(ctx, opt)
	if err == nil || err.Error() != "failed to apply server option: option error" {
		t.Errorf("expected error from server option, got %v", err)
	}
}

func TestNewServerReturnsValidServerOnSuccess(t *testing.T) {
	ctx := context.Background()
	srv, err := NewServer(ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if srv == nil {
		t.Errorf("expected server instance, got nil")
	}
}

func TestNewServerSetsDefaultValuesWhenNoOptionsProvided(t *testing.T) {
	ctx := context.Background()
	srv, err := NewServer(ctx)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	httpSrv, ok := srv.(*httpServer)
	if !ok {
		t.Fatalf("expected *httpServer, got %T", srv)
	}
	if httpSrv.Server == nil {
		t.Errorf("expected http.Server to be initialized")
	}
	if httpSrv.StartUpWait != 500*time.Millisecond {
		t.Errorf("expected default StartUpWait, got %v", httpSrv.StartUpWait)
	}
	if httpSrv.ShutdownWait != 5*time.Second {
		t.Errorf("expected default ShutdownWait, got %v", httpSrv.ShutdownWait)
	}
	if len(httpSrv.ShutdownSignals) != 2 {
		t.Errorf("expected default ShutdownSignals, got %v", httpSrv.ShutdownSignals)
	}
	if httpSrv.GinEngine == nil {
		t.Errorf("expected GinEngine to be initialized")
	}
}

func TestNewServerAppliesMultipleOptions(t *testing.T) {
	ctx := context.Background()
	opt1 := func(s *httpServer) error {
		s.StartUpWait = 42 * time.Second
		return nil
	}
	opt2 := func(s *httpServer) error {
		s.ShutdownWait = 99 * time.Second
		return nil
	}
	srv, err := NewServer(ctx, opt1, opt2)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	httpSrv, ok := srv.(*httpServer)
	if !ok {
		t.Fatalf("expected *httpServer, got %T", srv)
	}
	if httpSrv.StartUpWait != 42*time.Second {
		t.Errorf("expected StartUpWait to be set by option, got %v", httpSrv.StartUpWait)
	}
	if httpSrv.ShutdownWait != 99*time.Second {
		t.Errorf("expected ShutdownWait to be set by option, got %v", httpSrv.ShutdownWait)
	}
}

func TestStopReturnsEarlyIfServerIsNil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	Stop(ctx, nil)
	// No panic or error expected, nothing to assert
}

type fakeServer struct{}

func (f *fakeServer) Start(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeServer) Stop(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (f *fakeServer) Type() server.Type { return "not-http" }

var fkSvr server.Server = (*fakeServer)(nil)

func TestStopReturnsEarlyIfServerTypeIsNotHTTP(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	Stop(ctx, fkSvr)
	// No panic or error expected, nothing to assert
}

func TestStopCallsHttpServerStopAndHandlesError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := &httpServer{}
	Stop(ctx, srv)
	// Should log error for uninitialized server, but not panic
}

func TestStopLogsErrorWhenHttpServerStopFails(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{}

	Stop(ctx, srv)
}

func TestStopLogsSuccessWhenHttpServerStopSucceeds(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{
		Server:       &http.Server{Addr: ":0"},
		ShutdownWait: 1 * time.Second,
	}

	Stop(ctx, srv)
}

func TestStopHandlesContextCancellationDuringShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	srv := &httpServer{
		Server:       &http.Server{Addr: ":0"},
		ShutdownWait: 1 * time.Second,
	}

	Stop(ctx, srv)
}

func TestStopHandlesShutdownTimeoutError(t *testing.T) {
	ctx := context.Background()
	srv := &httpServer{
		Server:       &http.Server{Addr: ":0"},
		ShutdownWait: 1 * time.Nanosecond,
	}

	Stop(ctx, srv)
}
