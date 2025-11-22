package http

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"
)

type ServerOption func(*httpServer) error

func WithListenAddress(ctx context.Context, hostIP string, port string) ServerOption {
	return func(s *httpServer) error {
		if s.Server == nil {
			return fmt.Errorf("http server is not initialized")
		}
		s.Server.Addr = fmt.Sprintf("%s:%s", hostIP, port)
		return nil
	}
}

func WithTimeout(ctx context.Context, timeout TimeOut, timeOutValue time.Duration) ServerOption {
	return func(s *httpServer) error {
		if s.Server == nil {
			return fmt.Errorf("http server is not initialized")
		}
		switch timeout {
		case TimeOutIdle:
			s.Server.IdleTimeout = timeOutValue
		case TimeOutReadHeader:
			s.Server.ReadHeaderTimeout = timeOutValue
		case TimeOutRead:
			s.Server.ReadTimeout = timeOutValue
		case TimeOutWrite:
			s.Server.WriteTimeout = timeOutValue
		default:
			return fmt.Errorf("unknown timeout type")
		}
		return nil
	}
}

func WithProtocols(ctx context.Context, protocols ...string) ServerOption {
	return func(s *httpServer) error {
		if s.Server == nil {
			return fmt.Errorf("http server is not initialized")
		}
		oProtocols := new(http.Protocols)
		for _, protocol := range protocols {
			switch protocol {
			case string(ProtocolHTTP1):
				oProtocols.SetHTTP1(true)
			case string(ProtocolHTTP2):
				oProtocols.SetHTTP2(true)
			case string(ProtocolUnencryptedHTTP2):
				oProtocols.SetUnencryptedHTTP2(true)
			default:
				return fmt.Errorf("unknown protocol type")
			}
		}
		s.Server.Protocols = oProtocols
		return nil
	}
}

func WithWaitToStart(ctx context.Context, waitTime time.Duration) ServerOption {
	return func(s *httpServer) error {
		if s.Server == nil {
			return fmt.Errorf("http server is not initialized")
		}
		s.StartUpWait = waitTime
		return nil
	}
}
func WithWaitToShutdown(ctx context.Context, waitTime time.Duration) ServerOption {
	return func(s *httpServer) error {
		if s.Server == nil {
			return fmt.Errorf("http server is not initialized")
		}
		s.ShutdownWait = waitTime
		return nil
	}
}

func WithShutdownSignals(ctx context.Context, signals ...os.Signal) ServerOption {
	return func(s *httpServer) error {
		if s.Server == nil {
			return fmt.Errorf("http server is not initialized")
		}
		s.ShutdownSignals = signals
		return nil
	}
}
