package ioutils

import (
	"context"
	"fmt"
)

type SourceTypeName string

type SourceConfig interface {
	Supports(context context.Context, source Source) bool
}

type SourceConfigOpts[T Source, C SourceConfig] func(config C)

type SourceType interface {
	fmt.Stringer
	NewSource(context context.Context, config SourceConfig) (Source, error)
	NewConfig(context context.Context) (SourceConfig, error)
}
