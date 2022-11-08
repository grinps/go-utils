package ioutils

import "context"

type Source interface {
	Supports(context context.Context, capability SourceCapability) bool
}
