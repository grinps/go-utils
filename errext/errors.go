package errext

import (
	"errors"
	"fmt"
	"io"
	"runtime"
)

// EnableStackTrace controls whether stack traces are captured when errors are created.
// This is disabled by default to avoid performance impact.
var EnableStackTrace = false

// Error represents an implementation of [error] interface that supports [errext.ErrorCode] and [errors.Unwrap] capability.
//
// [error]: https://pkg.go.dev/builtin#error
type Error struct {
	errorCode ErrorCode
	text      string
	err       error
	args      []interface{}
	stack     []uintptr
}

// Unwrap returns the error associated with this error instance.
// This function implementation aligns with support for [errors.Unwrap]
func (wrappedError *Error) Unwrap() error {
	return wrappedError.err
}

// Error returns the error description associated with error.
// This function is required for implementation of [error] interface
//
// [error]: https://pkg.go.dev/builtin#error
func (wrappedError *Error) Error() string {
	msg := wrappedError.text
	if len(wrappedError.args) > 0 {
		msg += " " + formatAttributes(wrappedError.args)
	}
	return msg
}

// Format implements the fmt.Formatter interface to support custom printing.
// It supports %+v to print the error along with its stack trace if available.
func (wrappedError *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			io.WriteString(s, wrappedError.Error())
			if wrappedError.err != nil {
				fmt.Fprintf(s, ": %+v", wrappedError.err)
			}
			wrappedError.printStack(s)
			return
		}
		fallthrough
	case 's':
		io.WriteString(s, wrappedError.Error())
	case 'q':
		fmt.Fprintf(s, "%q", wrappedError.Error())
	}
}

func (wrappedError *Error) printStack(w io.Writer) {
	if len(wrappedError.stack) > 0 {
		frames := runtime.CallersFrames(wrappedError.stack)
		for {
			frame, more := frames.Next()
			fmt.Fprintf(w, "\n%+v", frame.Function)
			fmt.Fprintf(w, "\n\t%s:%d", frame.File, frame.Line)
			if !more {
				break
			}
		}
	}
}

// As implements the interface required by errors.As.
// It allows matching against the embedded ErrorCode or specific ErrorCode implementations.
func (wrappedError *Error) As(target any) bool {
	if targetVal, ok := target.(*ErrorCode); ok {
		*targetVal = wrappedError.errorCode
		return true
	}
	return false
}

// ErrorCode defines an interface to create errors based on pre-defined standards and capture specific information needed.
type ErrorCode interface {
	// New returns an instance of error created using the passed text and optional key-value attributes.
	New(text string, args ...interface{}) error

	// NewWithError create a wrapped error with the given string as error detail and optional key-value attributes.
	// Various methods like [errors.Unwrap] can be used in the created error object.
	NewWithError(text string, err error, args ...interface{}) error

	// AsError returns whether the given err is an instance of [Error] and the value as [Error]
	// If given object is not an instance nil is returned.
	AsError(err error) (*Error, bool)
}

// ErrorCodeValue represents the code part of the error.
type ErrorCodeValue int

// ErrorType represents the type of error being created.
type ErrorType string

const (
	// DefaultErrorCodeType represents the default type of errors.
	DefaultErrorCodeType string = ""
	// DefaultErrorCodeTypeObject represents the default type of errors.
	DefaultErrorCodeTypeObject ErrorType = ""
	// ErrorCodeNotSet is default value that should be returned if ErrorCodeValue is not set
	ErrorCodeNotSet ErrorCodeValue = -1
	// ErrorCodeUnknown represents code for unknown error
	ErrorCodeUnknown ErrorCodeValue = -2
	// ErrorCodeSuccess represent successful calls.
	ErrorCodeSuccess ErrorCodeValue = 0
)

// ErrorCodeImpl implements the [errext.ErrorCode] interface.
type ErrorCodeImpl struct {
	errorCode     ErrorCodeValue // ErrorCodeValue associated with this [errext.ErrorCode]
	errorCodeSet  bool           // Whether ErrorCodeValue has been set or not
	errorCodeType ErrorType      //Type of error.
	defaultArgs   []interface{}  // Default attributes
}

// New returns instance of error using given error string.
func (errorCode *ErrorCodeImpl) New(text string, args ...interface{}) error {
	return errorCode.NewWithError(text, nil, args...)
}

const errorCodeNotSetMessage = "error code has not been set. Please use function NewErrorCode(int) to create new ErrorCode"

// NewWithError returns instance of error by using the given error string and causing err.
func (errorCode *ErrorCodeImpl) NewWithError(text string, err error, args ...interface{}) error {
	if errorCode == nil {
		return defaultErrorCode.NewWithError(text, err, args...)
	}
	if !errorCode.errorCodeSet {
		return errors.New(errorCodeNotSetMessage)
	}

	var allArgs []interface{}
	if len(errorCode.defaultArgs) > 0 || len(args) > 0 {
		allArgs = make([]interface{}, 0, len(errorCode.defaultArgs)+len(args))
		allArgs = append(allArgs, errorCode.defaultArgs...)
		allArgs = append(allArgs, args...)
	}

	var stack []uintptr
	if EnableStackTrace {
		const depth = 32
		var pcs [depth]uintptr
		n := runtime.Callers(2, pcs[:])
		stack = pcs[0:n]
	}
	return &Error{
		errorCode: errorCode,
		text:      text,
		err:       err,
		args:      allArgs,
		stack:     stack,
	}
}

// AsError validates whether the given error is an instance of [errext.Error] and returns its reference if available nil otherwise.
func (errorCode *ErrorCodeImpl) AsError(err error) (*Error, bool) {
	if errAsError, ok := err.(*Error); ok {
		if errAsError.errorCode != nil {
			if errorCodeImpl, codeImplOk := errAsError.errorCode.(*ErrorCodeImpl); codeImplOk {
				if errorCodeImpl == errorCode {
					return errAsError, true
				}
			}
		}
	}
	return nil, false
}

var defaultErrorCode = &ErrorCodeImpl{
	errorCode:     ErrorCodeUnknown,
	errorCodeSet:  true,
	errorCodeType: DefaultErrorCodeTypeObject,
}

// NewErrorCode returns an implementation of new [errext.ErrorCode] with given errorId and [DefaultErrorCodeType]
// Refer to [WithErrorCode] for specific validations.
// This function should be used during package initialization to create new error codes to avoid any memory leaks
func NewErrorCode(errorId int) ErrorCode {
	return NewErrorCodeWithOptions(WithErrorCodeAndType(false, errorId, DefaultErrorCodeType))
}

// NewErrorCodeOfType returns an implementation of new [errext.ErrorCode] for the given errorId and given type.
//
// The given errorId and errorCodeType is converted to [errext.ErrorCodeValue] and [errext.ErrorType] respectively.
// This function should be used during package initialization to create new error codes to avoid any memory leaks
func NewErrorCodeOfType(errorId int, errorCodeType string) ErrorCode {
	return NewErrorCodeWithOptions(WithErrorCodeAndType(false, errorId, errorCodeType))
}

// NewUniqueErrorCode returns a unique version of [errext.ErrorCode] for the given errorId.
//
// If an existing errorId has already been created, the same is returned otherwise a new instance is returned.
// This function should be used during package initialization to create new error codes to avoid any memory leaks
func NewUniqueErrorCode(errorId int) ErrorCode {
	return NewErrorCodeWithOptions(WithErrorCodeAndType(true, errorId, DefaultErrorCodeType))
}

// NewUniqueErrorCodeOfType returns a unique version of [errext.ErrorCode] for the given errorId and error type
//
// If an existing errorId has already been created, the same is returned otherwise a new instance is returned.
// This function should be used during package initialization to create new error codes to avoid any memory leaks
func NewUniqueErrorCodeOfType(errorId int, codeType string) ErrorCode {
	return NewErrorCodeWithOptions(WithErrorCodeAndType(true, errorId, codeType))
}
