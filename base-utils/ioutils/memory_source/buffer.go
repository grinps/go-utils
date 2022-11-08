package memory_source

import (
	logger "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
	"io"
)

type BufferFullType uint8

const (
	BufferFullDefault           BufferFullType = iota
	BufferFullDropOnEnd                        = 0b1
	BufferFullExpandToMax                      = 0b10
	BufferFullContinueFromStart                = 0b100
	BufferFullStopOnEnd                        = 0b1000
)

type BufferEndOfFileType uint8

const (
	BufferEndOfFileDefault         BufferEndOfFileType = iota
	BufferEndOfFileIfNothingToRead                     = 0b1
	BufferEndOfFileNever                               = 0b10
	BufferEndOfFileAlways                              = 0b100
	// BufferEndOfFileOnMultipleEmptyBufferCalls - "multiple" is configurable
)

type BufferConfig struct {
	InitialSize   int
	MaxSize       int
	ExtendBy      int
	MaxUnreadSize int
	OnBufferFull  BufferFullType
	OnEndOfFile   BufferEndOfFileType
	// NumberOfEndOfFileFailures int
	//IsChunked        bool
	//MaxNumberOfChunk int
}

// Buffer is a variable-sized buffer of bytes with Read and Write methods.
//
// It heavily borrows from [bytes.Buffer] for implementation details while providing configurability and extensibility.
// It provides support for separate read and write offset tracking to support ring buffers.
// This implementation is not thread-safe.
// The implementation provides very limited set of APIs compared to [bytes.Buffer] to support different
type Buffer struct {
	buf         []byte // content are bytes[readOff:writeOff]
	readOff     int    // Offset to read from buffer
	writeOff    int    // Offset to write to buffer
	ringEngaged bool   // Whether writing from start of buffer has started. This is needed to track if write == read due to LEN() == 0 or due to ring buffer filling up.
	// TODO: Implement chunking
	// chunkLocators []int  // location of each chunk
	// chunkOffset   int    // Track the chunk being processed.
	config BufferConfig
}

const MemoryBufferErrors string = "MemoryBufferErrors"

// ErrTooLarge if memory cannot be allocated to store data in a buffer.
var ErrTooLarge = errext.NewErrorCodeOfType(1, MemoryBufferErrors)
var ErrInconsistentState = errext.NewErrorCodeOfType(2, MemoryBufferErrors)

const BufferMinSize = 64                 // smallest allocation for buffer storage
const BufferMaxSize = int(^uint(0) >> 1) // maximum size for storage

// Len returns number of bytes of unread buffer.
func (b *Buffer) Len() int {
	if b.writeOff > b.readOff && !b.ringEngaged { // linear & sparse ring buffer
		return b.writeOff - b.readOff
	} else if b.writeOff < b.readOff && b.ringEngaged { // ring buffer
		return (len(b.buf) - b.readOff) + b.writeOff
	} else if b.ringEngaged { // readOff == writeOff but ring buffer is full
		return len(b.buf)
	}
	return 0
}

// Truncate discards all but the first n unread bytes from buffer while maintaining
// same underlying storage
// It ignores calls with negative size or larger than unread buffer size.
func (b *Buffer) Truncate(n int) {
	if n == 0 {
		b.buf = b.buf[:0]
		b.readOff = 0
		b.writeOff = 0
	}
	if n > 0 && n < b.Len() {
		if b.writeOff > b.readOff && !b.ringEngaged { // linear & sparse ring buffer
			locationOfEndOfUnreadDataAfterTruncate := b.readOff + n
			b.buf = b.buf[:locationOfEndOfUnreadDataAfterTruncate]
			logger.Log("Truncated data on linear/sparse ring buffer", b, "resulting in loss of ", b.writeOff-locationOfEndOfUnreadDataAfterTruncate, " bytes")
			b.writeOff = locationOfEndOfUnreadDataAfterTruncate
		} else if b.writeOff <= b.readOff && b.ringEngaged { // ring buffer
			if n <= len(b.buf)-b.readOff {
				b.buf = b.buf[:b.readOff+n]
				logger.Log("Truncated data on buffer", b, "resulting in loss. Truncate size", n, "Read offset", b.readOff, "Length", len(b.buf), "Write offset", b.writeOff)
				b.writeOff = b.readOff + n
			} else { // n extends to start of ring buffer
				lengthOfUnReadBufferInEnd := len(b.buf) - b.readOff
				lengthOfDataAtStartAfterTruncate := n - lengthOfUnReadBufferInEnd
				if lengthOfDataAtStartAfterTruncate < b.writeOff {
					logger.Log("Truncated data on buffer", b, "resulting in loss of ", b.writeOff-lengthOfDataAtStartAfterTruncate, " bytes")
					b.writeOff = lengthOfDataAtStartAfterTruncate
				} else { // should never happen since we validate Len above
					logger.Warn("Logic Bug: Reached situation while truncating data on buffer", b, " with lengthOfDataAtStartAfterTruncate ", lengthOfDataAtStartAfterTruncate, ">", b.writeOff, ". Doing nothing")
				}
			}
		}
	}
}

// tryGrowByReSlice is an inlineable version of grow for the fast-case where the
// internal buffer only needs to be re-sliced.
// It returns the index where bytes should be written and whether it succeeded.
// Note: borrowed from [bytes.Buffer] and should only be used for linear buffers
func (b *Buffer) tryGrowByReSlice(n int) (int, bool) {
	if l := len(b.buf); n <= cap(b.buf)-l {
		b.buf = b.buf[:l+n]
		return l, true
	}
	return 0, false
}

// grow extends the buffer to guarantee space for n more bytes.
// It returns the index where bytes should be written.
// If the buffer can't grow it will return error ErrTooLarge.
func (b *Buffer) grow(n int) (int, error) {
	m := b.Len()
	// If buffer is empty, truncate to 0 to recover space.
	if m == 0 && b.readOff != 0 {
		b.Truncate(0)
	}
	// Try to grow by means of a reslice.
	if i, ok := b.tryGrowByReSlice(n); ok {
		return i, nil
	}
	if b.buf == nil && n <= BufferMinSize {
		b.buf = make([]byte, n, BufferMinSize)
		return 0, nil
	}
	c := cap(b.buf)
	if c > BufferMaxSize-c-n {
		return -1, ErrTooLarge.NewF("Memory buffer too large", b, " allowed max size", BufferMaxSize, "current cap", c, "requested increase", n)
	}
	var growErr error = nil
	if n <= c/2-m {
		// We can slide things down instead of allocating a new
		// slice. We only need m+n <= c to slide, but
		// we instead let capacity get twice as large so we
		// don't spend all our time copying.
		copy(b.buf, b.buf[b.readOff:b.writeOff])
	} else {
		// Add b.off to account for b.buf[:b.off] being sliced off the front.
		b.buf, growErr = growSlice(b.buf[b.readOff:], b.readOff+n)
		if growErr != nil {
			return -1, growErr
		}
	}
	// Restore b.off and len(b.buf).
	b.readOff = 0
	b.writeOff = m
	b.buf = b.buf[:m+n]
	return m, nil
}

// growSlice grows b by n, preserving the original content of b.
// If the allocation fails, it panics with ErrTooLarge.
func growSlice(b []byte, n int) (returnBytes []byte, returnErr error) {
	defer func() {
		if recovered := recover(); recovered != nil {
			returnBytes = nil
			returnErr = ErrTooLarge.NewF("Memory buffer ", b, " failed to grow by ", n, " due to panic with result ", recovered)
		}
	}()
	// TODO(http://golang.org/issue/51462): We should rely on the append-make
	// pattern so that the compiler can call runtime.growslice. For example:
	//	return append(b, make([]byte, n)...)
	// This avoids unnecessary zero-ing of the first len(b) bytes of the
	// allocated slice, but this pattern causes b to escape onto the heap.
	//
	// Instead use the append-make pattern with a nil slice to ensure that
	// we allocate buffers rounded up to the closest size class.
	c := len(b) + n // ensure enough space for n elements
	if c < 2*cap(b) {
		// The growth rate has historically always been 2x. In the future,
		// we could rely purely on append to determine the growth rate.
		c = 2 * cap(b)
	}
	b2 := append([]byte(nil), make([]byte, c)...)
	copy(b2, b)
	return b2[:len(b)], nil
}

// Write adds the contents of p to the buffer at write offset, growing the buffer
// if and as needed. The return value n is the length of p; err is always nil.
// If the buffer becomes too large, Write will panic with ErrTooLarge.
func (b *Buffer) Write(p []byte) (writtenBytes int, writeErr error) {
	writtenBytes = 0
	writeErr = nil
	sizeOfDataToWrite := len(p)
	sizeOfBuffer := len(b.buf)
	if b.writeOff >= b.readOff && !b.ringEngaged { // linear or sparse ring
		if sizeOfDataToWrite <= sizeOfBuffer-b.writeOff { // space left in buffer to write
			writtenBytes = copy(b.buf[b.writeOff:], p)
			b.writeOff += writtenBytes
		} else {
			availableSpace := sizeOfBuffer - b.writeOff
			if (sizeOfBuffer < b.config.InitialSize) || (b.config.OnBufferFull&BufferFullExpandToMax > 0) {
				growSliceBy := sizeOfDataToWrite
				if sizeOfBuffer < b.config.InitialSize { // first try to grow to initial size
					growSliceBy = b.config.InitialSize - sizeOfBuffer
				}
				if growSliceBy <= sizeOfDataToWrite && (b.config.OnBufferFull&BufferFullExpandToMax > 0) {
					if b.config.ExtendBy > 0 {
						numberOfExtensionNeededForInput := (sizeOfDataToWrite - availableSpace) / b.config.ExtendBy
						additionalSpaceNeeded := (sizeOfDataToWrite - availableSpace) % b.config.ExtendBy
						if additionalSpaceNeeded == 0 { // edge case of if extension exactly matches the ask just upto extension
							numberOfExtensionNeededForInput -= 1 // this balances the +1 below since we don't need extra extension
						}
						growSliceBy = (numberOfExtensionNeededForInput+1)*b.config.ExtendBy - availableSpace
					}
				}
				if b.writeOff+availableSpace+growSliceBy > b.config.MaxSize+b.readOff {
					growSliceBy = b.config.MaxSize - b.writeOff - availableSpace + b.readOff
				}
				if growSliceBy > 0 {
					// Resize the linear buffer
					_, growSucceeded := b.tryGrowByReSlice(growSliceBy)
					if !growSucceeded {
						_, growSizeErr := b.grow(growSliceBy)
						if growSizeErr != nil {
							writeErr = growSizeErr
						}
					}
				}
			}
			newSizeOfBuffer := len(b.buf)
			if sizeOfDataToWrite <= newSizeOfBuffer-b.writeOff {
				writtenBytes = copy(b.buf[b.writeOff:], p)
				b.writeOff += writtenBytes
			} else {
				// write all the way to the end of buffer
				maxDataThatCanBeWritten := newSizeOfBuffer - b.writeOff
				writtenBytes = copy(b.buf[b.writeOff:], p[:maxDataThatCanBeWritten])
				b.writeOff += writtenBytes
				if (BufferFullStopOnEnd|BufferFullExpandToMax)&b.config.OnBufferFull > 0 {
					writeErr = ErrTooLarge.NewF("Buffer", b, "is full and can not be expanded.Configuration", b.config, "current size", len(b.buf))
				} else if BufferFullContinueFromStart&b.config.OnBufferFull > 0 {
					additionalBytesWritten := fillARingBuffer(b, p, writtenBytes)
					writtenBytes += additionalBytesWritten
					if writtenBytes < len(p) {
						writeErr = ErrTooLarge.NewF("Ring Buffer", b, "is full and can not be expanded. Configuration", b.config, "current size", len(b.buf))
					}
				} else { //BufferFullDropOnEnd
					logger.Log("Dropping ", sizeOfDataToWrite-maxDataThatCanBeWritten, " while writing to buffer", b)
				}
			}
		}
	} else if b.writeOff <= b.readOff && b.ringEngaged { // ring
		if BufferFullContinueFromStart&b.config.OnBufferFull > 0 {
			if sizeOfDataToWrite <= b.readOff-b.writeOff {
				writtenBytes = copy(b.buf[b.writeOff:], p)
				b.writeOff += writtenBytes
			} else {
				// write all the way to the end of buffer
				maxDataThatCanBeWritten := sizeOfBuffer - b.writeOff
				if maxDataThatCanBeWritten > len(p) {
					maxDataThatCanBeWritten = len(p)
				}
				writtenBytes = copy(b.buf[b.writeOff:], p[:maxDataThatCanBeWritten])
				b.writeOff += writtenBytes
				if b.readOff < b.writeOff {
					b.readOff = b.writeOff // catch up the read offset so that the setting does not break
					b.ringEngaged = true   // engage ring since now read is behind write.
				}
				// fill the ring buffer
				additionalBytesWritten := fillARingBuffer(b, p, writtenBytes)
				writtenBytes += additionalBytesWritten
			}
		} else {
			logger.Warn("Buffer ", b, " has write offset ", b.writeOff, " less than read ", b.readOff, " even though it is not a ring buffer (", b.config.OnBufferFull, ").")
			writeErr = ErrInconsistentState.NewF("Buffer", b, "is not ring buffer but write offset < read offset. Configuration", b.config, "size", len(b.buf), "read offset", b.readOff, "write offset", b.writeOff)
		}
	}
	return
}

func fillARingBuffer(b *Buffer, p []byte, sizeOfDataAlreadyWritten int) (writtenBytes int) {
	writtenBytes = 0
	sizeOfBuffer := len(b.buf)
	if sizeOfBuffer == 0 {
		return 0
	}
	sizeOfDataToWrite := len(p)
	sizeOfRemainingDataAfterWritingToEndOfBuffer := sizeOfDataToWrite - sizeOfDataAlreadyWritten
	if sizeOfRemainingDataAfterWritingToEndOfBuffer <= 0 {
		return 0
	}
	// in case the data to write is very large then write will just fill the buffer multiple times
	// so we want to just skip to end to write just the relevant data
	numberOfPassThroughOfBufferNeededBeforeDataCanBeWritten := sizeOfRemainingDataAfterWritingToEndOfBuffer / sizeOfBuffer
	if numberOfPassThroughOfBufferNeededBeforeDataCanBeWritten > 0 {
		logger.Warn("Buffer ", b, " is being repassed ", numberOfPassThroughOfBufferNeededBeforeDataCanBeWritten, "times due to write string size resulting in significant loss of data being tracked. Please review the buffer size setting.")
		//TODO: Need additional reviews for logic correctness
		locationOnInputThatWillBeAtStart := sizeOfDataAlreadyWritten + (sizeOfBuffer * numberOfPassThroughOfBufferNeededBeforeDataCanBeWritten)
		sizeOfDataToWriteFromStart := len(p) - locationOnInputThatWillBeAtStart
		writtenBytes += (sizeOfBuffer * (numberOfPassThroughOfBufferNeededBeforeDataCanBeWritten - 1)) + sizeOfDataToWriteFromStart
		writtenBytes += copy(b.buf[(sizeOfDataToWrite-(locationOnInputThatWillBeAtStart)):], p[locationOnInputThatWillBeAtStart-(sizeOfBuffer-sizeOfDataToWriteFromStart):locationOnInputThatWillBeAtStart])
		finalBytesWritten := copy(b.buf[:(sizeOfDataToWrite-(locationOnInputThatWillBeAtStart))], p[locationOnInputThatWillBeAtStart:])
		writtenBytes += finalBytesWritten
		b.writeOff = finalBytesWritten
		b.readOff = b.writeOff // since buffer has been overwritten multiple times move read & write to same location
		b.ringEngaged = true   // since ring buffer has been filled
	} else if numberOfPassThroughOfBufferNeededBeforeDataCanBeWritten == 0 { // small data
		finalBytesWritten := copy(b.buf[0:], p[sizeOfDataAlreadyWritten:])
		writtenBytes += finalBytesWritten
		b.writeOff = finalBytesWritten
		if b.readOff < b.writeOff { // if write has overwritten previous unread data, start read from write location
			b.readOff = b.writeOff
		}
		b.ringEngaged = true // since ring buffer has been filled.
	} else { // Should never be reached
		logger.Warn("Logic Bug: Reached situation while writing data on buffer", b, " with numberOfPassThroughOfBufferNeededBeforeDataCanBeWritten ", numberOfPassThroughOfBufferNeededBeforeDataCanBeWritten, "<0. Doing nothing")
	}
	return
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is io.EOF (unless len(p) is zero);
// otherwise it is nil.
func (b *Buffer) Read(p []byte) (readBytes int, readErr error) {
	if b.Len() == 0 {
		// Buffer is empty, reset to recover space.
		b.Truncate(0)
		if len(p) == 0 {
			return 0, nil
		}
		if b.config.OnEndOfFile&(BufferEndOfFileNever) > 0 {
			return 0, nil
		} else {
			return 0, io.EOF
		}
	}
	if b.readOff < b.writeOff && !b.ringEngaged {
		readBytes = copy(p, b.buf[b.readOff:b.writeOff])
		b.readOff += readBytes
	} else if b.writeOff <= b.readOff && b.ringEngaged { //ring buffer
		outputBufferSize := len(p)
		bufferSize := len(b.buf)
		readToEndOfBuffer := outputBufferSize
		if b.readOff+outputBufferSize > bufferSize {
			readToEndOfBuffer = bufferSize - b.readOff
		}
		readBytes = copy(p, b.buf[b.readOff:b.readOff+readToEndOfBuffer])
		b.readOff += readBytes // next read should be from start
		if b.readOff == bufferSize {
			b.readOff = 0         // set the read to start of buffer since all the data has been read.
			b.ringEngaged = false // ring is disengaged since reading pointer < writing pointer
		}
		if outputBufferSize > readBytes { // additional data to be read from start of ring
			remainingBufferSize := outputBufferSize - readBytes
			if remainingBufferSize > b.writeOff { // only read upto the place where buffer has been updated.
				remainingBufferSize = b.writeOff
			}
			readAdditionalBytes := copy(p[readBytes:readBytes+remainingBufferSize], b.buf[b.readOff:b.readOff+remainingBufferSize])
			readBytes += readAdditionalBytes
			b.readOff = readAdditionalBytes
		}
	}
	if len(p)-readBytes > 0 {
		if b.config.OnEndOfFile&(BufferEndOfFileIfNothingToRead|BufferEndOfFileAlways) > 0 {
			readErr = io.EOF
		}
	}
	return readBytes, readErr
}

type BufferConfigOpt func(config *BufferConfig)

func NewBuffer(opts ...BufferConfigOpt) *Buffer {
	buffer := &Buffer{
		buf:      []byte{},
		readOff:  0,
		writeOff: 0,
		config:   BufferConfig{},
	}
	for _, bufferConfig := range opts {
		bufferConfig(&buffer.config)
	}
	return buffer
}

func AsRingBuffer(size int) BufferConfigOpt {
	if size < 0 {
		size = 0
	}
	return func(config *BufferConfig) {
		config.InitialSize = size
		config.MaxSize = size
		config.ExtendBy = 0
		config.OnEndOfFile = BufferEndOfFileIfNothingToRead
		config.OnBufferFull = BufferFullContinueFromStart
	}
}

func AsStaticBuffer(size int) BufferConfigOpt {
	if size < 0 {
		size = 0
	}
	return func(config *BufferConfig) {
		config.InitialSize = size
		config.MaxSize = size
		config.ExtendBy = 0
		config.OnEndOfFile = BufferEndOfFileIfNothingToRead
		config.OnBufferFull = BufferFullStopOnEnd
	}
}

func AsStaticBufferForTopResults(size int) BufferConfigOpt {
	if size < 0 {
		size = 0
	}
	return func(config *BufferConfig) {
		config.InitialSize = size
		config.MaxSize = size
		config.ExtendBy = 0
		config.OnEndOfFile = BufferEndOfFileIfNothingToRead
		config.OnBufferFull = BufferFullDropOnEnd
	}
}

func AsExtensibleBuffer(initialSize int, maxSize int, extensibleBy int) BufferConfigOpt {
	return func(config *BufferConfig) {
		if initialSize <= 0 {
			initialSize = BufferMinSize
		}
		config.InitialSize = initialSize
		if maxSize <= 0 {
			maxSize = BufferMaxSize
		}
		config.MaxSize = maxSize
		if extensibleBy <= 0 {
			extensibleBy = initialSize
		}
		config.ExtendBy = extensibleBy
		config.OnBufferFull = BufferFullExpandToMax
		config.OnEndOfFile = BufferEndOfFileIfNothingToRead
	}
}
