package errext

import (
	"errors"
	"fmt"
)

type Error struct {
	errorCode ErrorCode
	text      string
	err       error
}

func (wrappedError *Error) Unwrap() error {
	return wrappedError.err
}

func (wrappedError *Error) Error() string {
	return wrappedError.text
}

type ErrorCode interface {
	New(text string) error
	NewF(arguments ...interface{}) error
	NewWithError(text string, err error) error
	NewWithErrorF(err error, arguments ...interface{}) error
	AsError(err error) (*Error, bool)
}

type ErrorCodeImpl struct {
	errorCode     int
	errorCodeSet  bool
	errorCodeType string
}

func (errorCode *ErrorCodeImpl) New(text string) error {
	return errorCode.NewWithError(text, nil)
}

func (errorCode *ErrorCodeImpl) NewF(arguments ...interface{}) error {
	return errorCode.New(fmt.Sprintln(arguments...))
}

func (errorCode *ErrorCodeImpl) NewWithError(text string, err error) error {
	if !errorCode.errorCodeSet {
		return errors.New("error code has not been set. Please use function NewErrorCode(int) to create new ErrorCode")
	}
	return &Error{
		errorCode: errorCode,
		text:      text,
		err:       err,
	}
}

func (errorCode *ErrorCodeImpl) NewWithErrorF(err error, arguments ...interface{}) error {
	return errorCode.NewWithError(fmt.Sprintln(arguments...), err)
}

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

func NewErrorCode(errorId int) ErrorCode {
	return NewErrorCodeOfType(errorId, DefaultErrorCodeType)
}

func NewErrorCodeOfType(errorId int, errorCodeType string) ErrorCode {
	return &ErrorCodeImpl{
		errorCode:     errorId,
		errorCodeSet:  true,
		errorCodeType: errorCodeType,
	}
}

//TODO: Mutex
var uniqueErrorCodeMap = map[string]map[int]ErrorCode{}

const (
	DefaultErrorCodeType string = ""
)

func NewUniqueErrorCode(errorId int) ErrorCode {
	return NewUniqueErrorCodeOfType(errorId, DefaultErrorCodeType)
}

func NewUniqueErrorCodeOfType(errorId int, codeType string) ErrorCode {
	errorCodeMap, errorCodeMapExists := uniqueErrorCodeMap[codeType]
	if !errorCodeMapExists {
		errorCodeMap = map[int]ErrorCode{}
		uniqueErrorCodeMap[codeType] = errorCodeMap
	}
	errorCode, errorCodeExists := errorCodeMap[errorId]
	if !errorCodeExists {
		errorCodeMap[errorId] = NewErrorCodeOfType(errorId, codeType)
		errorCode = errorCodeMap[errorId]
	}
	return errorCode
}
