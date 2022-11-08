package ioutils

import (
	"context"
	"io"
)

type SourceCapability int

const (
	Reader SourceCapability = 0x1
	Writer SourceCapability = 0x10
	Stream SourceCapability = 0x100
	RPC    SourceCapability = 0x1000
)

type SourceStreamReader interface {
	GetReader(ctx context.Context) (io.Reader, error)
}

type SourceStreamWriter interface {
	GetWriter(ctx context.Context) (io.Writer, error)
}

type SourceRPC interface {
	GetProtocolHandler(ctx context.Context) (ProtocolHandler, error)
	IsStateful(ctx context.Context) (bool, error)
	GetProtocolHandlerForState(ctx context.Context, state State) (ProtocolHandler, error)
}

type ProtocolHandler interface {
	Handle(ctx context.Context, request any) (response any, state State, err error)
}

type State interface {
	Get(ctx context.Context, key any) any
	Set(ctx context.Context, key any, value any)
}
