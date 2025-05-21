package server

import "context"

type Server interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

type Type string

const (
	TypeUnknown Type = ""
)
