package memory_source

import (
	"context"
	"github.com/grinps/go-utils/base-utils/ioutils"
	"reflect"
)

const MemorySourceTypeName ioutils.SourceTypeName = "MemorySourceType"

var defaultMemorySourceType = &MemorySourceType{name: MemorySourceTypeName}

func init() {
	ioutils.Register(nil, MemorySourceTypeName, defaultMemorySourceType)
}

type MemorySourceType struct {
	name ioutils.SourceTypeName
}

func (memory *MemorySourceType) String() string {
	return string(memory.name)
}

func (memory *MemorySourceType) NewSource(context context.Context, config ioutils.SourceConfig) (ioutils.Source, error) {
	var returnSource ioutils.Source = nil
	var returnError error = nil
	if config == nil {
		return nil, MemorySourceInvalidConfiguration.New("No configuration was provided.")
	}
	if memorySourceConfig, isCorrectConfig := config.(*MemorySourceConfig); !isCorrectConfig {
		return nil, MemorySourceInvalidConfiguration.NewF("Memory source config expected, actual", reflect.TypeOf(config))
	} else {
		memorySource := &MemorySource{
			config: memorySourceConfig,
			memory: &Buffer{
				buf:      nil,
				readOff:  0,
				writeOff: 0,
				config:   memorySourceConfig.config,
			},
		}
		returnSource = memorySource
		returnError = nil
	}
	return returnSource, returnError
}

func (memory *MemorySourceType) NewConfig(context context.Context) (ioutils.SourceConfig, error) {
	return &MemorySourceConfig{
		config: BufferConfig{
			InitialSize:  BufferMinSize,
			MaxSize:      BufferMinSize,
			ExtendBy:     0,
			OnBufferFull: BufferFullStopOnEnd,
			OnEndOfFile:  BufferEndOfFileIfNothingToRead,
		},
	}, nil
}
