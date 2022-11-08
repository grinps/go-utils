package memory_source

import (
	"github.com/grinps/go-utils/errext"
	"io"
	"strings"
	"testing"
)

func TestBuffer_Default(t *testing.T) {
	var buffer = &Buffer{}
	if length := buffer.Len(); length != 0 {
		t.Error("Expected default buffer Len to be 0, Actual", buffer.Len())
	}
	buffer.Truncate(5)
	readBuffer := make([]byte, 10, 10)
	readBytes, readErr := buffer.Read(readBuffer)
	if readBytes > 0 {
		t.Error("Expected read bytes on empty as 0, Actual:", readBytes)
	}
	if readErr != io.EOF {
		t.Error("Expected read on empty to return EOF, actual", readErr)
	}
	writeBuffer := []byte("0123456")
	writeBytes, writeErr := buffer.Write(writeBuffer)
	if writeBytes != 0 {
		t.Error("Trying to write to empty buffer, expected 0, Actual", writeBytes)
	}
	if writeErr != nil {
		t.Errorf("Expected no error, Actual %#v", writeErr)
	}
	readAnotherBuffer := make([]byte, 10, 10)
	readBytes, readErr = buffer.Read(readAnotherBuffer)
	if readBytes != 0 {
		t.Error("Expected reading 0 bytes, actual", readBytes)
	}
	if readErr != io.EOF {
		t.Errorf("Expecting no error, actual %#v", readErr)
	}
}

func TestBuffer_Write(t *testing.T) {
	t.Run("ValidExtensibleBuffer", func(t *testing.T) {
		buffer := NewBuffer(AsExtensibleBuffer(10, 20, 5))
		unreadLength := buffer.Len()
		if unreadLength != 0 {
			t.Error("Expected unread length in default buffer as 0, Actual", unreadLength)
		}
		buffer.Truncate(10)
		readBuffer := make([]byte, 5)
		testRead(t, buffer, &readBuffer, false, make([]byte, 5), 0, true, nil)
		writeBuffer := []byte("01234") // Grow to 10 (cap 16) based on initial size (and then reset to 5)
		testWrite(t, buffer, &writeBuffer, true, 5, nil)
		writeBuffer = []byte("56789") // Grow by reslice to 10 since enough space is left
		testWrite(t, buffer, &writeBuffer, true, 5, nil)
		readBuffer = make([]byte, 1)
		testRead(t, buffer, &readBuffer, true, []byte("0"), 1, false, nil)
		writeBuffer = []byte("0123456") // Grow by 10 (Extend by x 2 (1+ 7/ExtendBy) ) & then reduce by 3 since only 7 needed
		testWrite(t, buffer, &writeBuffer, true, 7, nil)
		writeBuffer = []byte("78901") // Grow by reslice by 4 (since extend by calc of 10 exceeded max size) upto 20 max size.
		testWrite(t, buffer, &writeBuffer, false, 4, ErrTooLarge)
		readBuffer = make([]byte, 4)
		testRead(t, buffer, &readBuffer, true, []byte("1234"), 4, false, nil)
		writeBuffer = []byte("12345") // Grow by reslice by 4 (since extend by calc of 10 exceeds max size) to 24 (since 4 readoff)
		testWrite(t, buffer, &writeBuffer, false, 4, ErrTooLarge)
		writeBuffer = []byte("5") // Grow by reslice by 1 but then reduced to 24 since 25 > MaxSize + readOff
		testWrite(t, buffer, &writeBuffer, false, 0, ErrTooLarge)
		readBuffer = make([]byte, 20, 20)
		testRead(t, buffer, &readBuffer, true, []byte("56789012345678901234"), 20, false, nil)
		// readBuffer = make([]byte, 1, 1) // commented this to avoid truncate and recover space.
		// testRead(t, buffer, &readBuffer, false, make([]byte, 1, 1), 0, true, nil)
		writeBuffer = []byte("1234567890") // Truncate & Grow by slice to 10 and then reduce to 5
		testWrite(t, buffer, &writeBuffer, true, 10, nil)
		readBuffer = make([]byte, 5, 5)
		testRead(t, buffer, &readBuffer, true, []byte("12345"), 5, false, nil)
	})
	t.Run("ExtensibleBufferNegativeLengths", func(t *testing.T) {
		buffer := NewBuffer(AsExtensibleBuffer(-1, -1, -1))
		writeBuffer := []byte("01234") // Grow to 10 (cap 16) based on initial size (and then reset to 5)
		testWrite(t, buffer, &writeBuffer, true, 5, nil)
		writeBuffer = []byte("56789") // Grow by reslice to 10 since enough space is left
		testWrite(t, buffer, &writeBuffer, true, 5, nil)
		readBuffer := make([]byte, 10, 10)
		testRead(t, buffer, &readBuffer, true, []byte("0123456789"), 10, false, nil)
		readBuffer = make([]byte, 1, 1)
		testRead(t, buffer, &readBuffer, false, make([]byte, 1, 1), 0, true, nil)
		writeBuffer = []byte("01234") // Grow to 10 (cap 16) based on initial size (and then reset to 5)
		testWrite(t, buffer, &writeBuffer, true, 5, nil)
		readBuffer = make([]byte, 10, 10)
		testRead(t, buffer, &readBuffer, false, []byte("01234\x00\x00\u0000\u0000\u0000"), 5, true, nil)
	})
	t.Run("StaticBufferTopResult0Size", func(t *testing.T) {
		buffer := NewBuffer(AsStaticBufferForTopResults(-1))
		readBuffer := make([]byte, 3, 3)
		testRead(t, buffer, &readBuffer, false, []byte("\u0000\u0000\u0000"), 0, true, nil)
		writeBuffer := []byte("01234")
		testWrite(t, buffer, &writeBuffer, false, 0, nil)
	})
	t.Run("StaticBuffer0Size", func(t *testing.T) {
		buffer := NewBuffer(AsStaticBuffer(-1))
		readBuffer := make([]byte, 3, 3)
		testRead(t, buffer, &readBuffer, false, []byte("\u0000\u0000\u0000"), 0, true, nil)
		readBuffer = make([]byte, 0, 0)
		testRead(t, buffer, &readBuffer, true, []byte{}, 0, false, nil)
		writeBuffer := []byte("01234")
		testWrite(t, buffer, &writeBuffer, false, 0, ErrTooLarge)
	})
	t.Run("RingBufferNegativeSize", func(t *testing.T) {
		buffer := NewBuffer(AsRingBuffer(-1))
		readBuffer := make([]byte, 3, 3)
		testRead(t, buffer, &readBuffer, false, []byte("\u0000\u0000\u0000"), 0, true, nil)
		writeBuffer := []byte("01234")
		testWrite(t, buffer, &writeBuffer, false, 0, ErrTooLarge)

	})
	t.Run("RingBufferSize10", func(t *testing.T) {
		buffer := NewBuffer(AsRingBuffer(10))
		readBuffer := make([]byte, 3, 3)
		testRead(t, buffer, &readBuffer, false, []byte("\u0000\u0000\u0000"), 0, true, nil)
		writeBuffer := []byte("01234")
		testWrite(t, buffer, &writeBuffer, true, 5, nil)
		testLen(t, buffer, 5)
		writeBuffer = []byte("56789")
		testWrite(t, buffer, &writeBuffer, true, 5, nil)
		testLen(t, buffer, 10)
		writeBuffer = []byte("abcde")
		testWrite(t, buffer, &writeBuffer, true, 5, nil)
		testLen(t, buffer, 10)
		readBuffer = make([]byte, 5, 5)
		testRead(t, buffer, &readBuffer, true, []byte("56789"), 5, false, nil)
		testLen(t, buffer, 5)
		readBuffer = make([]byte, 5, 5)
		testRead(t, buffer, &readBuffer, true, []byte("abcde"), 5, false, nil)
		testLen(t, buffer, 0)
		readBuffer = make([]byte, 1, 1)
		testRead(t, buffer, &readBuffer, false, []byte("\u0000"), 0, true, nil)
		writeBuffer = []byte("ghijklmno")
		testWrite(t, buffer, &writeBuffer, true, 9, nil)
		testLen(t, buffer, 9)
		readBuffer = make([]byte, 2, 2)
		testRead(t, buffer, &readBuffer, true, []byte("gh"), 2, false, nil)
		readBuffer = make([]byte, 8, 8)
		testRead(t, buffer, &readBuffer, false, []byte("ijklmno\u0000"), 7, true, nil)
		readBuffer = make([]byte, 2, 2)
		testRead(t, buffer, &readBuffer, false, []byte("\u0000\u0000"), 0, true, nil)
		testLen(t, buffer, 0)
		writeBuffer = []byte("0123456789abc") // Write fills buffer and then overrides first 3 bytes
		testWrite(t, buffer, &writeBuffer, true, 13, nil)
		testLen(t, buffer, 10)
		readBuffer = make([]byte, 3, 3)
		testRead(t, buffer, &readBuffer, true, []byte("345"), 3, false, nil)
		testLen(t, buffer, 7)
		writeBuffer = []byte("defghi") // Writing passes the read location
		testWrite(t, buffer, &writeBuffer, true, 6, nil)
		testLen(t, buffer, 10)
		readBuffer = make([]byte, 1, 1)
		testRead(t, buffer, &readBuffer, true, []byte("9"), 1, false, nil)
		testLen(t, buffer, 9)
		readBuffer = make([]byte, 5, 5)
		testRead(t, buffer, &readBuffer, true, []byte("abcde"), 5, false, nil)
		testLen(t, buffer, 4)
		writeBuffer = []byte("j01")
		testWrite(t, buffer, &writeBuffer, true, 3, nil)
		testLen(t, buffer, 7)
		writeBuffer = []byte("23")
		testWrite(t, buffer, &writeBuffer, true, 2, nil)
		testLen(t, buffer, 9)
		writeBuffer = []byte(strings.Repeat("1234567890", 4))
		testWrite(t, buffer, &writeBuffer, true, 40, nil)
		testLen(t, buffer, 10)
		readBuffer = make([]byte, 9, 9)
		testRead(t, buffer, &readBuffer, true, []byte("123456789"), 9, false, nil)

	})
}

func TestBuffer_Truncate(t *testing.T) {
	t.Run("StaticBuffer10", func(t *testing.T) {
		buffer := NewBuffer(AsStaticBuffer(10))
		writeBuffer := []byte("0")
		testWrite(t, buffer, &writeBuffer, true, 1, nil)
		testLen(t, buffer, 1)
		writeBuffer = []byte("123456789")
		testWrite(t, buffer, &writeBuffer, true, 9, nil)
		testLen(t, buffer, 10)
		buffer.Truncate(4)
		testLen(t, buffer, 4)
		readBuffer := make([]byte, 4, 4)
		testRead(t, buffer, &readBuffer, true, []byte("0123"), 4, false, nil)
	})

	t.Run("RingBuffer10", func(t *testing.T) {
		buffer := NewBuffer(AsRingBuffer(10))
		writeBuffer := []byte("0")
		testWrite(t, buffer, &writeBuffer, true, 1, nil)
		testLen(t, buffer, 1)
		writeBuffer = []byte("123456789")
		testWrite(t, buffer, &writeBuffer, true, 9, nil)
		testLen(t, buffer, 10)
		buffer.Truncate(4)
		testLen(t, buffer, 4)
		readBuffer := make([]byte, 4, 4)
		testRead(t, buffer, &readBuffer, true, []byte("0123"), 4, false, nil)
		testLen(t, buffer, 0)
		writeBuffer = []byte("01234")
		testWrite(t, buffer, &writeBuffer, true, 5, nil)
		testLen(t, buffer, 5)
		writeBuffer = []byte("abcdefghi")
		testWrite(t, buffer, &writeBuffer, true, 9, nil)
		testLen(t, buffer, 10) // bcdefghi4a
		buffer.Truncate(4)     // truncate to 8+4 = 12
		readBuffer = make([]byte, 5, 5)
		testRead(t, buffer, &readBuffer, false, []byte("4abc\u0000"), 4, true, nil)
		readBuffer = make([]byte, 7, 7)
		testRead(t, buffer, &readBuffer, false, []byte("\u0000\u0000\u0000\u0000\u0000\u0000\u0000"), 0, true, nil)

	})

}

func testRead(t *testing.T, buffer *Buffer, readBuffer *[]byte, expectedSuccess bool, expectedReadData []byte, expectedReadBytes int, expectingEOF bool, expectedErrCode errext.ErrorCode) {
	expectedString := string(expectedReadData)
	readBytes, readErr := buffer.Read(*readBuffer)
	if readBytes != expectedReadBytes {
		t.Errorf("Read case %s Expected %d Actual %d", expectedString, expectedReadBytes, readBytes)
	}
	if expectedSuccess && readErr != nil {
		t.Errorf("Read case %s Expected no error, Actual %#v", expectedString, readErr)
	}
	if expectedErrCode != nil {
		if _, isErrCode := expectedErrCode.AsError(readErr); isErrCode {
			t.Errorf("Read case %s Expected %#v, Actual %#v", expectedString, expectedErrCode, readErr)
		}
	}
	if expectingEOF && readErr != io.EOF {
		t.Errorf("Read case %s Expected EOF Actual %#v", expectedString, readErr)
	}
	if expectedString != string(*readBuffer) {
		t.Errorf("Read case %s Expected %s, Actual %s", expectedString, expectedString, string(*readBuffer))
	}
}

func testWrite(t *testing.T, buffer *Buffer, writeBuffer *[]byte, expectedSuccess bool, expectedWriteBytes int, expectedErrCode errext.ErrorCode) {
	expectedString := string(*writeBuffer)
	writeBytes, writeErr := buffer.Write(*writeBuffer)
	if writeBytes != expectedWriteBytes {
		t.Errorf("Write case %s expected write bytes size %d, Actual %d", expectedString, expectedWriteBytes, writeBytes)
	}
	if expectedSuccess && writeErr != nil {
		t.Errorf("Write case %s Expected no error, Actual %#v", expectedString, writeErr)
	}
	if expectedErrCode != nil {
		if _, isErrCode := expectedErrCode.AsError(writeErr); !isErrCode {
			t.Errorf("Write case %s Expected error code %#v, Actual %#v", expectedString, expectedErrCode, writeErr)
		}
	}

}

func testLen(t *testing.T, buffer *Buffer, expectedSize int) {
	unreadDataSize := buffer.Len()
	if expectedSize != unreadDataSize {
		t.Errorf("Expected unread data size %d, Actual %d", expectedSize, unreadDataSize)
	}
}
