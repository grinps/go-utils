package errext

import (
	"errors"
	"fmt"
	"strings"
)

// Error represents an implementation of [error] interface that supports [errext.ErrorCode] and [errors.Unwrap] capability.
//
// [error]: https://pkg.go.dev/builtin#error
type Error struct {
	errorCode ErrorCode
	text      string
	err       error
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
	return wrappedError.text
}

// ErrorCode defines an interface to create errors based on pre-defined standards and capture specific information needed.
// For example in case a set of invalid input errors need to be created, a standard [errext.ErrorCode] can be defined and then
// on detection of error, specific parameter values can be provided to create the actual error and return. This error can then
// be used by calling function to verify whether Error is of type Input Validation ( [errext.ErrorCodeImpl.AsError] ) or validate the parameter
// responsible for error.
type ErrorCode interface {
	// New returns an instance of error created using the passed text.
	New(text string) error
	// NewF creates an error string using provided phrases by concatenating them similar to [fmt.Sprintln].
	NewF(arguments ...interface{}) error
	// NewWithError create a wrapped error with the given string as error detail.
	// Various methods like [errors.Unwrap] can be used in the created error object.
	NewWithError(text string, err error) error
	// NewWithErrorF creates a wrapped error using the given phrases.
	NewWithErrorF(err error, arguments ...interface{}) error
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
	template      *errTemplate   // Template to create new errors using parameters
}

// New returns instance of error using given error string.
// Refer to [errext.ErrorCodeImpl.NewF] for additional details.
func (errorCode *ErrorCodeImpl) New(text string) error {
	return errorCode.NewWithError(text, nil)
}

// NewF returns instance of error by using the given arguments.
func (errorCode *ErrorCodeImpl) NewF(arguments ...interface{}) error {
	return errorCode.NewWithErrorF(nil, arguments...)
}

const errorCodeNotSetMessage = "error code has not been set. Please use function NewErrorCode(int) to create new ErrorCode"

// NewWithError returns instance of error by using the given error string and causing err.
func (errorCode *ErrorCodeImpl) NewWithError(text string, err error) error {
	if errorCode == nil {
		return defaultErrorCode.NewWithError(text, err)
	}
	if !errorCode.errorCodeSet {
		return errors.New(errorCodeNotSetMessage)
	}
	return &Error{
		errorCode: errorCode,
		text:      text,
		err:       err,
	}
}

// NewWithErrorF returns instance of error by using the given arguments and causing err.
func (errorCode *ErrorCodeImpl) NewWithErrorF(err error, arguments ...interface{}) error {
	if errorCode != nil {
		return errorCode.NewWithError(printArguments(generateFromTemplate(errorCode.template, arguments...)...), err)
	} else {
		return errorCode.NewWithError(printArguments(arguments...), err)
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

func printArguments(args ...interface{}) (returnValue string) {
	returnValue = "Error message could not be generated."
	printValue := fmt.Sprintln(args...)
	returnValue = strings.Trim(printValue, "\n")
	return
}

var defaultErrorCode = &ErrorCodeImpl{
	errorCode:     ErrorCodeUnknown,
	errorCodeSet:  true,
	errorCodeType: DefaultErrorCodeTypeObject,
}

// NewErrorCode returns an implementation of new [errext.ErrorCode] with given errorId and [DefaultErrorCodeType]
// Refer to [WithErrorCode] for specific validations.
func NewErrorCode(errorId int) ErrorCode {
	return NewErrorCodeWithOptions(WithErrorCodeAndType(false, errorId, DefaultErrorCodeType))
}

// NewErrorCodeOfType returns an implementation of new [errext.ErrorCode] for the given errorId and given type.
//
// The given errorId and errorCodeType is converted to [errext.ErrorCodeValue] and [errext.ErrorType] respectively.
func NewErrorCodeOfType(errorId int, errorCodeType string) ErrorCode {
	return NewErrorCodeWithOptions(WithErrorCodeAndType(false, errorId, errorCodeType))
}

// NewUniqueErrorCode returns a unique version of [errext.ErrorCode] for the given errorId.
//
// If an existing errorId has already been created, the same is returned otherwise a new instance is returned.
func NewUniqueErrorCode(errorId int) ErrorCode {
	return NewErrorCodeWithOptions(WithErrorCodeAndType(true, errorId, DefaultErrorCodeType))
}

// NewUniqueErrorCodeOfType returns a unique version of [errext.ErrorCode] for the given errorId and error type
//
// If an existing errorId has already been created, the same is returned otherwise a new instance is returned.
func NewUniqueErrorCodeOfType(errorId int, codeType string) ErrorCode {
	return NewErrorCodeWithOptions(WithErrorCodeAndType(true, errorId, codeType))
}
