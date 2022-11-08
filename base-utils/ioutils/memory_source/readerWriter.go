package memory_source

/*
import (
	"bytes"
	logger "github.com/grinps/go-utils/base-utils/logs"
	"io"
	"sync"
)

type memoryReaderWriter struct {
	source *MemorySource
	mutex  *sync.Mutex
}

func (rw *memoryReaderWriter) Read(p []byte) (n int, err error) {
	var returnError = io.EOF
	var returnReadBytes = 0
	if rw != nil && rw.source != nil && rw.source.memory != nil {
		switch rw.source.config.fullStrategy {
		case MemoryFullExtendStrategy:
			rw.mutex.Lock()
			defer rw.mutex.Unlock()
			readBytes, readError := rw.source.memory.Read(p)
			if readError == io.EOF {
				if rw.source.config.maxSize < 0 {
					returnError = nil
				} else if rw.source.memory.Cap() >= rw.source.config.maxSize {
					returnError = io.EOF
				} else {
					returnError = nil
				}
			} else {
				returnError = readError
			}
			returnReadBytes = readBytes

		case MemoryFullOverwriteFromStartStrategy:
			rw.mutex.Lock()
			defer rw.mutex.Unlock()
			readBytes, readError := rw.source.memory.Read(p)
			if readError == io.EOF {
				returnError = nil
			} else {
				returnError = readError
			}
			returnReadBytes = readBytes
		case MemoryFullSkipStrategy:
			rw.mutex.Lock()
			defer rw.mutex.Unlock()
			readBytes, readError := rw.source.memory.Read(p)
			if readError == io.EOF {
				if rw.source.memory.Cap() >= rw.source.config.initialSize {
					returnError = io.EOF
				} else {
					returnError = nil
				}
			} else {
				returnError = readError
			}
			returnReadBytes = readBytes
		default:
			readBytes, readError := rw.source.memory.Read(p)
			returnReadBytes = readBytes
			returnError = readError
		}
	}
	return returnReadBytes, returnError
}

func (rw *memoryReaderWriter) Close() error {
	if rw != nil && rw.source != nil && rw.source.memory != nil {
		rw.mutex.Lock()
		defer rw.mutex.Unlock()
		memorySourceRefresh(rw.source)
	}
	return nil
}

func (rw *memoryReaderWriter) Write(p []byte) (int, error) {
	var returnWrittenBytes = 0
	var returnError error = nil
	if rw != nil && rw.source != nil && rw.source.memory != nil {
		rw.mutex.Lock()
		defer rw.mutex.Unlock()
		toWriteBytes := len(p)
		currentSize := rw.source.memorySize
		filledUpTo := rw.source.memory.Len()
		if currentSize >= filledUpTo+toWriteBytes {
			logger.Log("Current size enough to write input")
			returnWrittenBytes, returnError = rw.source.memory.Write(p)
		} else { // don't have enough space to write.
			rw.source.Reset()
			totalAddedMemory := 0
			addedMemory, _ := rw.source.Resize(-1) // add based on configuration
			totalAddedMemory += addedMemory
			if totalAddedMemory < toWriteBytes {
				switch rw.source.config.fullStrategy {
				case MemoryFullExtendStrategy:
					logger.Log("Current size enough to write input")
					addedMemory, _ = rw.source.Resize(rw.source.memorySize + (toWriteBytes - addedMemory))
					totalAddedMemory += addedMemory
					if totalAddedMemory >= toWriteBytes {
						returnWrittenBytes, returnError = rw.source.memory.Write(p)
					} else {
						returnError = bytes.ErrTooLarge
					}
				case MemoryFullOverwriteFromStartStrategy:
					//TODO: rewrite to reduce the buffer copies.
					unreadDataLength := rw.source.memory.Len()
					unreadData := rw.source.memory.Next(unreadDataLength)
					rw.source.memory.Reset()
					rw.source.memorySize = 0
					overwriteMemorySize, _ := rw.source.Resize(unreadDataLength + toWriteBytes)
					if overwriteMemorySize >= unreadDataLength+toWriteBytes {
						returnWrittenBytes, returnError = rw.source.memory.Write(p)
						if returnError != nil {
							addedUnreadDataLength, unreadDataAddErr := rw.source.memory.Write(unreadData)
							logger.Log("Added full unread data", "addedUnreadDataLength", addedUnreadDataLength, "unreadDataAddErr", unreadDataAddErr)
						}
					} else if overwriteMemorySize >= toWriteBytes {
						returnWrittenBytes, returnError = rw.source.memory.Write(p)
						if returnError != nil { // since no error let's add unread data upto the possible value
							remainingSpaceForUnreadData := overwriteMemorySize - toWriteBytes
							locationOfUnreadDataToWriteFrom := unreadDataLength - remainingSpaceForUnreadData
							addedUnreadDataLength, unreadDataAddErr := rw.source.memory.Write(unreadData[locationOfUnreadDataToWriteFrom:])
							logger.Log("Added partial unread data", "addedUnreadDataLength", addedUnreadDataLength, "unreadDataAddErr", unreadDataAddErr)
						}
					} else { // total size is too small to write even the input so fill from start upto available size
						returnWrittenBytes, returnError = rw.source.memory.Write(p[:overwriteMemorySize])
					}
				case MemoryFullSkipStrategy, MemoryFullDefaultStrategy:
					if toWriteBytes-totalAddedMemory > 0 {
						returnWrittenBytes, returnError = rw.source.memory.Write(p[:totalAddedMemory])
					}
				default:
					if toWriteBytes-totalAddedMemory > 0 {
						returnWrittenBytes, returnError = rw.source.memory.Write(p[:totalAddedMemory])
					}
				}
			} else { // added memory enough to write input
				returnWrittenBytes, returnError = rw.source.memory.Write(p)
			}
		}
	}
	return returnWrittenBytes, returnError
}


*/
