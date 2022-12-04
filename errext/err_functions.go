package errext

import "sync"

// ErrorCodeOptions specifies function that can be used to customize the ErrorCodeImpl value
type ErrorCodeOptions func(errorCode *ErrorCodeImpl) *ErrorCodeImpl

// NewErrorCodeWithOptions returns an ErrorCode customized using given options.
func NewErrorCodeWithOptions(options ...ErrorCodeOptions) ErrorCode {
	newErrorCode := &ErrorCodeImpl{
		errorCode:     ErrorCodeNotSet,
		errorCodeSet:  false,
		errorCodeType: DefaultErrorCodeTypeObject,
	}
	for _, errCodeOption := range options {
		newErrorCode = errCodeOption(newErrorCode)
	}
	if !newErrorCode.errorCodeSet || newErrorCode.errorCode == ErrorCodeNotSet {
		newErrorCode.errorCode = nextErrorCodeForType(newErrorCode.errorCodeType)
		newErrorCode.errorCodeSet = true
	}
	return newErrorCode
}

// ErrorCodeErrors represents type of errors created during ErrorCode related operations.
const ErrorCodeErrors ErrorType = "ErrorCodeErrors"

// ErrCodeInvalidErrorCode can be used to create errors due to invalid error code values.
var ErrCodeInvalidErrorCode = &ErrorCodeImpl{
	errorCode:     1,
	errorCodeSet:  true,
	errorCodeType: ErrorCodeErrors,
}

// WithErrorCode provides option to create [errext.ErrorCode] for given [errext.ErrorCodeValue] of  [errext.DefaultErrorCodeTypeObject]
// type.
//
// This option is available for backward compatibility and option with generated error code values (for example [errext.WithErrorType]) should be used.
func WithErrorCode(errorCode ErrorCodeValue) ErrorCodeOptions {
	return WithUniqueCodeAndType(false, errorCode, DefaultErrorCodeTypeObject)
}

// WithErrorType provides option to create [errext.ErrorCode] for given [errext.ErrorType] with generated error code.
func WithErrorType(errorType ErrorType) ErrorCodeOptions {
	return WithUniqueCodeAndType(false, ErrorCodeNotSet, errorType)
}

// WithErrorCodeAndType specifies option to create unique [errext.ErrorCode] for given error code and [errext.ErrorType]
func WithErrorCodeAndType(unique bool, errorCode int, errorType string) ErrorCodeOptions {
	return WithUniqueCodeAndType(unique, ErrorCodeValue(errorCode), ErrorType(errorType))
}

var uniqueErrorCodeMap = map[ErrorType]map[ErrorCodeValue]*ErrorCodeImpl{}
var uniqueErrorCodeMutex = &sync.Mutex{}

// WithUniqueCodeAndType defines option to create [errext.ErrorCode] with given [errext.ErrorCodeValue] and [errext.ErrorType]
//
// If unique is true then an existing instance of [errext.ErrorCode] matching given errorCode and errorType is returned
// if already created otherwise a new instance is returned.
// In case [errext.ErrorCodeNotSet] is passed, a new [errext.ErrorCodeValue] is generated and used. If value is passed,
// given value must be below [errext.ErrorCodeValueStartValueForGeneration] to avoid collision.
func WithUniqueCodeAndType(unique bool, errorCode ErrorCodeValue, errorType ErrorType) ErrorCodeOptions {
	applicableErrorType := errorType
	if errorType == "" {
		applicableErrorType = DefaultErrorCodeTypeObject
	}
	applicableErrorCode := errorCode
	if errorCode == ErrorCodeNotSet {
		applicableErrorCode = nextErrorCodeForType(applicableErrorType)
	} else if errorCode >= ErrorCodeValueStartValueForGeneration {
		panic(ErrCodeInvalidErrorCode.NewF("Error code ", errorCode, " which is part of generated range can not be set for type", errorType, ". Please set value below ", ErrorCodeValueStartValueForGeneration))
	} else if errorCode < 0 {
		panic(ErrCodeInvalidErrorCode.NewF("Error code ", errorCode, " can not be set < 0. Please set value between 0 and ", ErrorCodeValueStartValueForGeneration))
	}
	return func(errorCodeImpl *ErrorCodeImpl) *ErrorCodeImpl {
		applicableErrorCodeImpl := errorCodeImpl
		if unique {
			uniqueErrorCodeMutex.Lock()
			defer uniqueErrorCodeMutex.Unlock()
			errorCodeMap, errorCodeMapExists := uniqueErrorCodeMap[applicableErrorType]
			if !errorCodeMapExists {
				errorCodeMap = map[ErrorCodeValue]*ErrorCodeImpl{}
				uniqueErrorCodeMap[applicableErrorType] = errorCodeMap
			}
			errorCodeImpl, errorCodeExists := errorCodeMap[applicableErrorCode]
			if errorCodeExists {
				return errorCodeImpl
			} else {
				errorCodeMap[applicableErrorCode] = applicableErrorCodeImpl
			}
		}
		applicableErrorCodeImpl.errorCode = applicableErrorCode
		applicableErrorCodeImpl.errorCodeSet = true
		applicableErrorCodeImpl.errorCodeType = applicableErrorType
		return applicableErrorCodeImpl
	}
}

var errorCodeGenerationTracker = map[ErrorType]ErrorCodeValue{}
var errorCodeGenerationMutex = &sync.Mutex{}

const ErrorCodeValueStartValueForGeneration ErrorCodeValue = 0b10000000 // 128 to support customizable error codes .

func nextErrorCodeForType(errorCodeType ErrorType) ErrorCodeValue {
	errorCodeGenerationMutex.Lock()
	defer errorCodeGenerationMutex.Unlock()
	applicableCodeValue := ErrorCodeNotSet
	if currentCodeValue, hasCodeValue := errorCodeGenerationTracker[errorCodeType]; hasCodeValue {
		applicableCodeValue = currentCodeValue
	} else {
		applicableCodeValue = ErrorCodeValueStartValueForGeneration
	}
	var nextCodeValue = ErrorCodeValue(int(applicableCodeValue) + 1)
	errorCodeGenerationTracker[errorCodeType] = nextCodeValue
	return nextCodeValue
}
