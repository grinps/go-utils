package errext

// HandleErrorMode defines how to handle an error condition
type HandleErrorMode int

const (
	// NotSpecified means the way to handle error condition is not specified.
	NotSpecified  HandleErrorMode = iota
	// IgnoreError implies that error should be ignored and cleared.
	IgnoreError                   = 0b001
	// GenerateError generates the error using the given error input & ErrorGenerator
	GenerateError                 = 0b010
	// Panic triggers panic with the appropriate value.
	Panic                         = 0b100
)

const (
	// ErrReasonUnknown defines that reason for error is not known
	ErrReasonUnknown = "Unknown error occurred"
)

// ErrorGenerator defines a way to enable caller to generate error for the given reason and cause.
type ErrorGenerator func(reason string, err error, additionalFields ...interface{}) error

// PanicHandler is a function that can be used to define how the panic should be handled in an API.
// It should typically be used with defer to ensure execution in case method panics.
func PanicHandler(contextName string, errorHandleMode HandleErrorMode, errorCodeGen ErrorGenerator, returnedError *error) {
	if panicOutcome := recover(); panicOutcome != nil {
		HandleOptionError(contextName, errorHandleMode, returnedError, errorCodeGen, ErrReasonUnknown, panicOutcome)
	}
}

// HandleOptionError provides a standard way to handle error at a method level by reducing boiler plate code.
// Depending on the mode, it can generate error, generate panic or reset current error
func HandleOptionError(contextDetail string, errorHandleMode HandleErrorMode, errRef *error, errGen ErrorGenerator, reasonForError string, panicErr any) {
	if int(errorHandleMode)&IgnoreError == IgnoreError {
		if errRef != nil {
			*errRef = nil
		}
		return
	}
	var generatedErr error
	if int(errorHandleMode)&GenerateError == GenerateError &&
		reasonForError != "" {
		var inputErr error = nil
		if errRef != nil {
			inputErr = *errRef
		}
		if errGen != nil {
			generatedErr = errGen(reasonForError, inputErr)
		} else {
			generatedErr = &simpleError{
				errorString: reasonForError,
				wrappedErr:  inputErr,
			}
		}
		if errRef != nil {
			*errRef = generatedErr
		} else {
			panic(generatedErr)
		}
	}
	if int(errorHandleMode)&Panic == Panic {
		if generatedErr != nil {
			panic(generatedErr)
		} else if panicErr != nil {
			panic(panicErr)
		}
	}
}

// simpleError is an implemented of error that supports wrappedErr
type simpleError struct {
	errorString string
	wrappedErr  error
}

func (errWrapper *simpleError) Error() string {
	return errWrapper.errorString
}
func (errWrapper *simpleError) Unwrap() error {
	return errWrapper.wrappedErr
}
