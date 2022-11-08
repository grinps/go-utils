package memory_source

import (
	"context"
	ioutils "github.com/grinps/go-utils/base-utils/ioutils"
)

type MemorySourceConfig struct {
	config BufferConfig
}

func (config *MemorySourceConfig) Supports(context context.Context, source ioutils.Source) bool {
	if source != nil {
		if _, isMemorySource := source.(*MemorySource); isMemorySource {
			return true
		}
	}
	return false
}

func WithSize(size int) ioutils.SourceConfigOpts[*MemorySource, *MemorySourceConfig] {
	return func(config *MemorySourceConfig) {
		AsStaticBuffer(size)(&config.config)
	}
}

func WithSizeForTopResults(size int) ioutils.SourceConfigOpts[*MemorySource, *MemorySourceConfig] {
	return func(config *MemorySourceConfig) {
		AsStaticBufferForTopResults(size)(&config.config)
	}
}

func WithExtendStrategy(initial int, max int, extend int) ioutils.SourceConfigOpts[*MemorySource, *MemorySourceConfig] {
	return func(config *MemorySourceConfig) {
		AsExtensibleBuffer(initial, max, extend)(&config.config)
	}
}
