package memory_source

import (
	"context"
	ioutils "github.com/grinps/go-utils/base-utils/ioutils"
	"github.com/grinps/go-utils/errext"
	"io"
)

const MemorySourceErrors string = "MemorySourceErrors"

var MemorySourceInvalidConfiguration = errext.NewErrorCodeOfType(1, MemorySourceErrors)

type MemorySource struct {
	memory *Buffer
	config *MemorySourceConfig
}

func (source *MemorySource) Supports(context context.Context, capability ioutils.SourceCapability) bool {
	switch capability {
	case ioutils.Reader, ioutils.Writer:
		return true
	}
	return false
}

func (source *MemorySource) GetReader(ctx context.Context) (io.Reader, error) {
	if source == nil || source.memory == nil {
		return &nopReaderWriterCloser{}, nil
	}
	return source.memory, nil
}

func (source *MemorySource) GetWriter(ctx context.Context) (io.Writer, error) {
	if source == nil || source.memory == nil {
		return &nopReaderWriterCloser{}, nil
	}
	return source.memory, nil
}
