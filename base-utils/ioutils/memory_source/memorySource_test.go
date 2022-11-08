package memory_source

import (
	"context"
	"errors"
	"github.com/grinps/go-utils/base-utils/ioutils"
	"github.com/grinps/go-utils/errext"
	"io"
	"testing"
)

func TestMemorySource_Nil(t *testing.T) {
	t.Run("NilSource", func(t *testing.T) {
		var memSource *MemorySource = nil
		if !memSource.Supports(nil, ioutils.Reader) {
			t.Errorf("Expecting support for Reader capability")
		}
		if memSource.Supports(nil, ioutils.RPC) {
			t.Errorf("Not expecting support for Reader capability")
		}
		nilReader, nilReaderErr := memSource.GetReader(nil)
		if nilReaderErr != nil {
			t.Errorf("Expecting no err actual %#v", nilReaderErr)
		}
		if nilReader == nil {
			t.Errorf("Expecting Reader, Actual nil reader")
		} else {
			if _, isNopReader := nilReader.(*nopReaderWriterCloser); !isNopReader {
				t.Errorf("Expecting reader of type nopReaderWriterCloser, actual %#v", nilReader)
			}
			readBytesIntoBuffer := make([]byte, 5, 5)
			readBytes, readErr := nilReader.Read(readBytesIntoBuffer)
			if readBytes != 0 {
				t.Errorf("Expecting to read 0, actual %d", readBytes)
			}
			if readErr != io.EOF {
				t.Errorf("Expecting EOF to returned, Actual %#v", readErr)
			}
		}
		if asCloseableReader, isCloseable := nilReader.(io.Closer); !isCloseable {
			t.Errorf("Expected Reader to implement Closer, Actual it does not")
		} else if closeErr := asCloseableReader.Close(); closeErr != nil {
			t.Errorf("Expecting no error while closing, Actual %#v", closeErr)
		}
		nilWriter, nilWriterErr := memSource.GetWriter(nil)
		if nilWriterErr != nil {
			t.Errorf("Expecting no err to get writer actual %#v", nilWriterErr)
		}
		if nilWriter == nil {
			t.Errorf("Expecting writer, Actual nil writer")
		} else {
			writeByteFromBuffer := []byte("12345")
			writtenBytes, writeErr := nilWriter.Write(writeByteFromBuffer)
			if writtenBytes != 0 {
				t.Errorf("Expecting to write 0, actual %d", writtenBytes)
			}
			if _, isErrLarge := ErrTooLarge.AsError(writeErr); !isErrLarge {
				t.Errorf("Expecting ErrTooLarge to be returned on write, Actual %#v", writeErr)
			}
		}
		if asCloseableWriter, isCloseable := nilWriter.(io.Closer); !isCloseable {
			t.Errorf("Expected writer to implement Closer, Actual it does not")
		} else if closeErr := asCloseableWriter.Close(); closeErr != nil {
			t.Errorf("Expecting no error while closing writer, Actual %#v", closeErr)
		}
	})

}

func TestMemoryReaderWriter_Write(t *testing.T) {
	var writeData1 = []byte("12345")
	var writeData2 = []byte("67890")
	var writeData3 = []byte("abcde")
	var writeData4 = []byte("ghijk")

	t.Run("ValidSourceWithSkipStrategy", func(t *testing.T) {
		memorySource := ioutils.Resolve[*MemorySource, *MemorySourceConfig](nil, MemorySourceTypeName, WithSizeForTopResults(10))
		writer, writerErr := memorySource.GetWriter(nil)
		if writerErr != nil {
			t.Errorf("Expected no error to get writer, Actual %#v", writerErr)
		}
		write(t, writer, writeData1, true, len(writeData1), nil, nil)
		write(t, writer, writeData2, true, len(writeData2), nil, nil)
		write(t, writer, writeData3, true, 0, nil, nil)
	})
	t.Run("ValidSourceWithExtendStrategy", func(t *testing.T) {
		memorySource := ioutils.Resolve[*MemorySource, *MemorySourceConfig](nil, MemorySourceTypeName, WithExtendStrategy(5, 10, 3))
		writer, _ := getReaderWriter(t, memorySource)
		write(t, writer, writeData1, true, len(writeData1), nil, nil)
		write(t, writer, writeData2, true, len(writeData2), nil, nil)
		write(t, writer, writeData3, false, 0, nil, ErrTooLarge)
	})
	t.Run("ValidSourceWithExtendStrategyWithIntermittentRead", func(t *testing.T) {
		memorySource := ioutils.Resolve[*MemorySource, *MemorySourceConfig](nil, MemorySourceTypeName, WithExtendStrategy(5, 10, 3))
		writer, reader := getReaderWriter(t, memorySource) // write 5
		write(t, writer, writeData1, true, len(writeData1), nil, nil)
		read(t, reader, []byte("123"), true, 3, nil, nil)
		write(t, writer, writeData2, true, len(writeData2), nil, nil)
		read(t, reader, []byte("456"), true, 3, nil, nil)
		write(t, writer, writeData3, true, 5, nil, nil)
		write(t, writer, writeData4, false, 1, nil, ErrTooLarge)
	})

}

func getReaderWriter(t *testing.T, memorySource *MemorySource) (io.Writer, io.Reader) {
	writer, writerErr := memorySource.GetWriter(nil)
	if writerErr != nil {
		t.Errorf("Expected no error to get writer, Actual %#v", writerErr)
	}
	reader, readErr := memorySource.GetReader(nil)
	if readErr != nil {
		t.Errorf("Expected no error to get reader, Actual %#v", readErr)
	}
	return writer, reader
}
func write(t *testing.T, writer io.Writer, writeData []byte, expectingSuccess bool, expectingWriteSize int, expectedErr error, expectedErrCode errext.ErrorCode) {
	writtenBytes, writeErr := writer.Write(writeData)
	if writtenBytes != expectingWriteSize {
		t.Error("Expected to write ", string(writeData), " of size", expectingWriteSize, "Actual", writtenBytes)
	}
	if expectingSuccess && writeErr != nil {
		t.Errorf("Expected no error while writing %s Actual %#v", string(writeData), writeErr)
	}
	if !expectingSuccess && writeErr == nil {
		t.Errorf("Expected an error while writing %s but no error was returned", writeData)
	}
	if expectedErrCode != nil {
		if _, isErrorCode := expectedErrCode.AsError(writeErr); !isErrorCode {
			t.Errorf("Expected Error Code while writing %s of type %#v, Actual %#v", writeData, expectedErrCode, writeErr)
		}
	}
	if expectedErr != nil {
		if !errors.Is(writeErr, expectedErr) {
			t.Errorf("Expected Error while writing %s of type %#v, Actual %#v", writeData, expectedErr, writeErr)
		}
	}
}

func read(t *testing.T, reader io.Reader, expectedReadData []byte, expectingSuccess bool, expectingReadSize int, expectedErr error, expectedErrCode errext.ErrorCode) {
	var readData = make([]byte, expectingReadSize)
	readBytes, readErr := reader.Read(readData)
	readString := string(readData)
	expectedReadString := string(expectedReadData)
	if readString != expectedReadString {
		t.Error("Expected to read ", expectedReadString, "Actual", readString)
	}
	if readBytes != expectingReadSize {
		t.Error("Expected to read ", expectedReadString, " of size ", expectingReadSize, "Actual", readBytes)
	}
	if expectingSuccess && readErr != nil {
		t.Errorf("Expected no error while reading %s Actual %#v", readString, readErr)
	}
	if !expectingSuccess && readErr == nil {
		t.Errorf("Expected an error while reading %s but no error was returned", readString)
	}
	if expectedErrCode != nil {
		if _, isErrorCode := expectedErrCode.AsError(readErr); !isErrorCode {
			t.Errorf("Expected Error Code while reading %s of type %#v, Actual %#v", readString, expectedErrCode, readErr)
		}
	}
	if expectedErr != nil {
		if !errors.Is(readErr, expectedErr) {
			t.Errorf("Expected Error while reading %s of type %#v, Actual %#v", readData, expectedErr, readErr)
		}
	}
}

func TestMemorySourceType_String(t *testing.T) {
	t.Run("DefaultObject", func(t *testing.T) {
		defaultObject := &MemorySourceType{}
		if defaultObject.String() != "" {
			t.Error("Expected empty string, actual", defaultObject.String())
		}
	})
	t.Run("InitObject", func(t *testing.T) {
		defaultObject := &MemorySourceType{name: "TestingId"}
		if defaultObject.String() != "TestingId" {
			t.Error("Expected TestingId string, actual", defaultObject.String())
		}
	})
}

type nilSourceConfig struct {
	counter int
}

func (config *nilSourceConfig) Supports(context context.Context, source ioutils.Source) bool {
	return false
}

type nilSource struct {
	config *nilSourceConfig
}

func (source *nilSource) Supports(context context.Context, capability ioutils.SourceCapability) bool {
	return false
}

func TestMemorySourceType_NewSource(t *testing.T) {
	t.Run("NilConfig", func(t *testing.T) {
		source, err := ioutils.Default().Resolve(nil, MemorySourceTypeName, nil)
		if err == nil {
			t.Errorf("Expected error due to nil config being passed. Actual nil")
		}
		if _, isErrCode := MemorySourceInvalidConfiguration.AsError(errors.Unwrap(err)); !isErrCode {
			t.Errorf("Expected error of type MemorySourceInvalidConfiguration, Actual %#v", err)
		}
		if source != nil {
			t.Errorf("Expected nil source, Actual %#v", source)
		}
	})
	t.Run("MismatchingConfig", func(t *testing.T) {
		source, err := ioutils.Default().Resolve(nil, MemorySourceTypeName, &nilSourceConfig{})
		if err == nil {
			t.Errorf("Expected error due to mismatching config being passed. Actual nil")
		}
		if _, isErrCode := MemorySourceInvalidConfiguration.AsError(errors.Unwrap(err)); !isErrCode {
			t.Errorf("Expected error of type MemorySourceInvalidConfiguration, Actual %#v", err)
		}
		if source != nil {
			t.Errorf("Expected nil source, Actual %#v", source)
		}
	})
}

func TestMemorySourceConfig_Supports(t *testing.T) {
	t.Run("MemorySourceConfigAndMemorySource", func(t *testing.T) {
		source, resolveErr := ioutils.ResolveE(nil, MemorySourceTypeName, WithSize(10))
		if resolveErr != nil {
			t.Errorf("Expected no error while creating memory source. Actual %#v", resolveErr)
		}
		sourceConfig, configErr := ioutils.Default().NewSourceConfig(nil, MemorySourceTypeName)
		if configErr != nil {
			t.Errorf("Expected no error while creating memory source config. Actual %#v", configErr)
		}
		if !sourceConfig.Supports(nil, source) {
			t.Error("Expected memory source config to support memory source. Actual false")
		}
	})
	t.Run("MemorySourceConfigAndNilSource", func(t *testing.T) {
		sourceConfig, configErr := ioutils.Default().NewSourceConfig(nil, MemorySourceTypeName)
		if configErr != nil {
			t.Errorf("Expected no error while creating memory source config. Actual %#v", configErr)
		}
		if sourceConfig.Supports(nil, &nilSource{config: &nilSourceConfig{counter: 2}}) {
			t.Error("Expected memory source config to not support nilsource source. Actual true")
		}
	})
}
