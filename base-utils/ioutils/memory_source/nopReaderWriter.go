package memory_source

import (
	"io"
)

type nopReaderWriterCloser struct{}

func (reader *nopReaderWriterCloser) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (reader *nopReaderWriterCloser) Close() error {
	return nil
}

func (reader *nopReaderWriterCloser) Write(p []byte) (n int, err error) {
	return 0, ErrTooLarge.New("No space to write to NopReaderWriter")
}
