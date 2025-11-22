package http

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/grinps/go-utils/server"
)

const (
	TypeHTTP server.Type = "http"
)

type TimeOut string

const (
	TimeOutUnknown    TimeOut = ""
	TimeOutIdle       TimeOut = "idle"
	TimeOutReadHeader TimeOut = "readHeader"
	TimeOutRead       TimeOut = "read"
	TimeOutWrite      TimeOut = "write"
)

type Protocol string

const (
	ProtocolUnknown          Protocol = ""
	ProtocolHTTP1                     = "http1"
	ProtocolHTTP2                     = "http2"
	ProtocolUnencryptedHTTP2          = "http2+unencrypted"
)

type ServerState string

const (
	ServerStateUnknown  ServerState = ""
	ServerStateStarting ServerState = "starting"
	ServerStateStopping ServerState = "stopping"
	ServerStateStopped  ServerState = "stopped"
	ServerStateError    ServerState = "error"
)

type OS interface {
	ReadFile(path string) ([]byte, error)
}
type httpServer struct {
	Server          *http.Server
	State           ServerState
	StartUpWait     time.Duration
	StartupError    error
	ShutdownWait    time.Duration
	ShutdownErr     error
	ShutdownSignals []os.Signal
	TLSConfig       *tls.Config
	OSReference     OS
	GinEngine       *gin.Engine
}

func (s *httpServer) Type() server.Type {
	return TypeHTTP
}

func (s *httpServer) Start(ctx context.Context) error {
	if s.Server == nil {
		return fmt.Errorf("http server is not initialized")
	}
	slog.Debug("starting http server", "address", s.Server.Addr)
	var applicableAddress string = ":http"
	if s.Server.Addr != "" {
		applicableAddress = s.Server.Addr
	} else if s.TLSConfig != nil {
		applicableAddress = ":https"
	}
	var applicableListener net.Listener
	slog.Debug("applicable address for http server", "address", applicableAddress)
	listener, err := net.Listen("tcp", applicableAddress)
	if err != nil {
		s.State = ServerStateError
		return fmt.Errorf("failed to start http server due to listener creation error: %w", err)
	}
	applicableListener = listener
	if s.TLSConfig != nil {
		applicableListener = tls.NewListener(listener, s.TLSConfig)
	}
	if s.GinEngine != nil {
		s.Server.Handler = s.GinEngine.Handler()
	}
	s.State = ServerStateStarting
	go HandleOSSignal(ctx, func() { Stop(ctx, s) }, s.ShutdownSignals...)
	go func() {
		serveErr := s.Server.Serve(applicableListener)
		if serveErr != nil && !errors.Is(http.ErrServerClosed, serveErr) {
			slog.Warn("failed to start http server", "error", serveErr)
			s.StartupError = serveErr
			s.State = ServerStateError
		}
	}()
	HandleWaitTime(ctx, s.StartUpWait, func() {
		// Check if an error occurred during startup
		if s.StartupError != nil {
			s.State = ServerStateError
		}
	})
	return s.StartupError
}

func Stop(ctx context.Context, s server.Server) {
	if s == nil {
		slog.Error("http server is not initialized")
		return
	}
	if s.Type() != TypeHTTP {
		slog.Error("server type is not http", "type", s.Type())
		return
	}
	if httpServer, ok := s.(*httpServer); ok {
		if err := httpServer.Stop(ctx); err != nil {
			slog.Error("failed to stop http server", "error", err)
		} else {
			slog.Debug("http server stopped successfully")
		}
	}
}

// HandleWaitTime waits for a specified duration before executing a post function.
// if the wait time is zero or negative OR passed context is cancelled,
// it does not execute the post function.
func HandleWaitTime(ctx context.Context, waitTime time.Duration, postFunc func()) {
	slog.Debug("waiting for", "wait_duration", waitTime)
	var waitTimeFinished = false
	if waitTime <= 0 {
		slog.Debug("no wait time specified, skipping wait")
	} else {
		select {
		case <-ctx.Done():
			slog.Debug("context cancelled before http server started", "error", ctx.Err())
		case <-time.After(waitTime):
			waitTimeFinished = true
			slog.Debug("http server started or wait time exceeded")
		}
	}
	if waitTimeFinished && postFunc != nil {
		slog.Debug("executing post function after wait time")
		postFunc()
	}
}

// HandleOSSignal listens for specified OS signals and executes a post function when a signal is received.
// If the context is cancelled before a signal is received, it logs the cancellation and does not execute the post function.
func HandleOSSignal(ctx context.Context, postFunc func(), signals ...os.Signal) {
	if len(signals) == 0 {
		slog.Debug("no signals specified, skipping signal handler")
		return
	}
	stopChan := make(chan os.Signal, 1)
	slog.Debug("Registering signal channel", "signals", signals)
	signal.Notify(stopChan, signals...)
	slog.Debug("starting signal handler")
	var signalReceived bool
	select {
	case doneVal := <-ctx.Done():
		slog.Debug("context cancelled before signal handler triggered", "receivedValue", doneVal, "error", ctx.Err())
	case val := <-stopChan:
		signalReceived = true
		slog.Debug("signal received", "signal", val)
	}
	if signalReceived {
		slog.Debug("received stop signal, executing function")
		postFunc()
	} else {
		slog.Debug("skipping executing function since no signal was received")
	}
}

func (s *httpServer) Stop(ctx context.Context) error {
	if s.Server == nil {
		return fmt.Errorf("http server is not initialized")
	}
	s.State = ServerStateStopping
	applicableContext := ctx
	var applicableCancelFunction context.CancelFunc
	if s.ShutdownWait > 0 {
		applicableContext, applicableCancelFunction = context.WithTimeout(ctx, s.ShutdownWait)
	} else {
		applicableContext, applicableCancelFunction = context.WithCancel(ctx)
	}
	defer applicableCancelFunction()
	slog.Debug("stopping http server", "address", s.Server.Addr)
	if err := s.Server.Shutdown(applicableContext); err != nil {
		s.State = ServerStateError
		s.ShutdownErr = err
		return fmt.Errorf("failed to stop http server: %w", err)
	} else {
		s.State = ServerStateStopped
	}
	return nil
}

func NewServer(ctx context.Context, options ...ServerOption) (server.Server, error) {
	httpServer := &httpServer{
		Server: &http.Server{
			Handler: http.DefaultServeMux,
		},
		StartUpWait:     500 * time.Millisecond,
		ShutdownWait:    5 * time.Second,
		ShutdownSignals: []os.Signal{os.Interrupt, os.Kill},
		TLSConfig:       nil,
		GinEngine:       gin.New(),
	}
	for _, option := range options {
		if err := option(httpServer); err != nil {
			return nil, fmt.Errorf("failed to apply server option: %w", err)
		}
	}
	return httpServer, nil
}
