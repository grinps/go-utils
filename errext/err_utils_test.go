package errext

import (
	"errors"
	"testing"
)

const ErrorReasonPrefix = "Error due to "

var simpleErrCodeGenerator ErrorGenerator = func(reason string, err error, additionalFields ...interface{}) error {
	internalErr := &simpleError{
		errorString: ErrorReasonPrefix + reason,
	}
	if err != nil {
		internalErr.wrappedErr = err
	}
	return internalErr
}

func TestPanicHandler(t *testing.T) {
	t.Run("No panic", func(t *testing.T) {
		var returnedErr error = nil
		defer func() {
			if panicOutcome := recover(); panicOutcome != nil {
				t.Errorf("Expected not panic error, actual %#v", panicOutcome)
			}
		}()
		defer PanicHandler("", Panic, nil, &returnedErr)
		//nothing to generate panic
	})
	t.Run("PanicWithContinuePanic", func(t *testing.T) {
		var returnedErr error = nil
		var panicValue = struct{}{}
		defer func() {
			if panicOutcome := recover(); panicOutcome != panicValue {
				t.Errorf("Expected panic %#v, actual %#v", panicValue, panicOutcome)
			}
		}()
		defer PanicHandler("PanicWithContinuePanic", Panic, nil, &returnedErr)
		panic(panicValue)
	})
	t.Run("PanicToError", func(t *testing.T) {
		var returnedErr error = nil
		var panicValue = struct{}{}
		defer func() {
			if panicOutcome := recover(); panicOutcome != nil {
				t.Errorf("Expected no panic, actual %#v", panicOutcome)
			}
			if returnedErr == nil {
				t.Errorf("Expected error, actual nil error")
			} else if returnedErr.Error() != ErrorReasonPrefix+ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrorReasonPrefix+ErrReasonUnknown, returnedErr.Error())
			}
		}()
		defer PanicHandler("PanicToError", GenerateError, simpleErrCodeGenerator, &returnedErr)
		defer func() {
			if panicOutcome := recover(); panicOutcome == nil {
				t.Errorf("Expected panic, actual no panic")
			} else if panicOutcome != panicValue {
				t.Errorf("Expected panic value %#v, actual %#v", panicValue, panicOutcome)
			} else {
				panic(panicOutcome)
			}
		}()
		panic(panicValue)
	})
	t.Run("PanicToErrorWithNoCodeGenerator", func(t *testing.T) {
		var returnedErr error = nil
		var panicValue = struct{}{}
		defer func() {
			if panicOutcome := recover(); panicOutcome != nil {
				t.Errorf("Expected no panic, actual %#v", panicOutcome)
			} else if returnedErr.Error() != ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrReasonUnknown, returnedErr.Error())
			}
		}()
		defer PanicHandler("PanicToErrorWithNoCodeGenerator", GenerateError, nil, &returnedErr)
		panic(panicValue)
	})
	t.Run("PanicToErrorWithNilError", func(t *testing.T) {
		var panicValue = struct{}{}
		defer func() {
			if panicOutcome := recover(); panicOutcome == nil {
				t.Errorf("Expected panic, actual no panic")
			} else if asErr, isErr := panicOutcome.(error); !isErr {
				t.Errorf("Expected an error, actual %#v", panicOutcome)
			} else if asErr.Error() != ErrorReasonPrefix+ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrorReasonPrefix+ErrReasonUnknown, asErr.Error())
			}
		}()
		defer PanicHandler("PanicToErrorWithNilError", GenerateError, simpleErrCodeGenerator, nil)
		panic(panicValue)
	})
	t.Run("PanicToErrorWithNilErrorAndNilGenerator", func(t *testing.T) {
		var panicValue = struct{}{}
		defer func() {
			if panicOutcome := recover(); panicOutcome == nil {
				t.Errorf("Expected panic, actual no panic")
			} else if asErr, isErr := panicOutcome.(error); !isErr {
				t.Errorf("Expected an error, actual %#v", panicOutcome)
			} else if asErr.Error() != ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrReasonUnknown, asErr.Error())
			}
		}()
		defer PanicHandler("PanicToErrorWithNoCodeGenerator", GenerateError, nil, nil)
		panic(panicValue)
	})
	t.Run("PanicToError+Panic", func(t *testing.T) {
		var returnedErr error = nil
		var panicValue = struct{}{}
		defer func() {
			if panicOutcome := recover(); panicOutcome == nil {
				t.Errorf("Expected panic, actual no panic")
			} else if asErr, isErr := panicOutcome.(error); !isErr {
				t.Errorf("Expected an error, actual %#v", panicOutcome)
			} else if asErr.Error() != ErrorReasonPrefix+ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrorReasonPrefix+ErrReasonUnknown, asErr.Error())
			}
			if returnedErr == nil {
				t.Errorf("Expected error, actual no error")
			} else if returnedErr.Error() != ErrorReasonPrefix+ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrorReasonPrefix+ErrReasonUnknown, returnedErr.Error())
			}
		}()
		defer PanicHandler("PanicToError+Panic", GenerateError|Panic, simpleErrCodeGenerator, &returnedErr)
		panic(panicValue)
	})
	t.Run("PanicToError+PanicWithNoCodeGenerator", func(t *testing.T) {
		var returnedErr error = nil
		var panicValue = struct{}{}
		defer func() {
			if panicOutcome := recover(); panicOutcome == nil {
				t.Errorf("Expected panic, actual no panic")
			} else if asErr, isErr := panicOutcome.(error); !isErr {
				t.Errorf("Expected an error, actual %#v", panicOutcome)
			} else if asErr.Error() != ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrReasonUnknown, asErr.Error())
			}
			if returnedErr == nil {
				t.Errorf("Expected error, actual no error")
			} else if returnedErr.Error() != ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrReasonUnknown, returnedErr.Error())
			}
		}()
		defer PanicHandler("PanicToError+PanicWithNoCodeGenerator", GenerateError|Panic, nil, &returnedErr)
		panic(panicValue)
	})
	t.Run("PanicToError+PanicWithNilError", func(t *testing.T) {
		var panicValue = struct{}{}
		defer func() {
			if panicOutcome := recover(); panicOutcome == nil {
				t.Errorf("Expected panic, actual no panic")
			} else if asErr, isErr := panicOutcome.(error); !isErr {
				t.Errorf("Expected an error, actual %#v", panicOutcome)
			} else if asErr.Error() != ErrorReasonPrefix+ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrorReasonPrefix+ErrReasonUnknown, asErr.Error())
			}
		}()
		defer PanicHandler("PanicToError+PanicWithNilError", GenerateError|Panic, simpleErrCodeGenerator, nil)
		panic(panicValue)
	})
	t.Run("PanicToError+PanicWithNilErrorAndNilGenerator", func(t *testing.T) {
		var panicValue = struct{}{}
		defer func() {
			if panicOutcome := recover(); panicOutcome == nil {
				t.Errorf("Expected panic, actual no panic")
			} else if asErr, isErr := panicOutcome.(error); !isErr {
				t.Errorf("Expected an error, actual %#v", panicOutcome)
			} else if asErr.Error() != ErrReasonUnknown {
				t.Errorf("Expected error as %s actual %s", ErrReasonUnknown, asErr.Error())
			}
		}()
		defer PanicHandler("PanicToError+PanicWithNilErrorAndNilGenerator", GenerateError, nil, nil)
		panic(panicValue)
	})
	t.Run("IgnoreError", func(t *testing.T) {
		var panicValue = struct{}{}
		var initialErr = errors.New("InitErr")
		defer func() {
			if panicOutcome := recover(); panicOutcome != nil {
				t.Errorf("Expected no panic, actual %#v", panicOutcome)
			}
			if initialErr != nil {
				t.Errorf("Expected nil initial err, actual %#v", initialErr)
			}
		}()
		defer PanicHandler("IgnoreError", IgnoreError, simpleErrCodeGenerator, &initialErr)
		panic(panicValue)
	})
	t.Run("IgnoreError+GenerateErr", func(t *testing.T) {
		var panicValue = struct{}{}
		var initialErr = errors.New("InitErr")
		defer func() {
			if panicOutcome := recover(); panicOutcome != nil {
				t.Errorf("Expected no panic, actual %#v", panicOutcome)
			}
			if initialErr != nil {
				t.Errorf("Expected nil initial err, actual %#v", initialErr)
			}
		}()
		defer PanicHandler("IgnoreError+GenerateErr", IgnoreError|GenerateError, simpleErrCodeGenerator, &initialErr)
		panic(panicValue)
	})
	t.Run("IgnoreError|Panic", func(t *testing.T) {
		var panicValue = struct{}{}
		var initialErr = errors.New("InitErr")
		defer func() {
			if panicOutcome := recover(); panicOutcome != nil {
				t.Errorf("Expected no panic, actual %#v", panicOutcome)
			}
			if initialErr != nil {
				t.Errorf("Expected nil initial err, actual %#v", initialErr)
			}
		}()
		defer PanicHandler("IgnoreError|Panic", IgnoreError|Panic, simpleErrCodeGenerator, &initialErr)
		panic(panicValue)
	})
	t.Run("IgnoreError|Panic|GenerateErr", func(t *testing.T) {
		var panicValue = struct{}{}
		var initialErr = errors.New("InitErr")
		defer func() {
			if panicOutcome := recover(); panicOutcome != nil {
				t.Errorf("Expected no panic, actual %#v", panicOutcome)
			}
			if initialErr != nil {
				t.Errorf("Expected nil initial err, actual %#v", initialErr)
			}
		}()
		defer PanicHandler("IgnoreError|Panic|GenerateErr", IgnoreError|Panic|GenerateError, simpleErrCodeGenerator, &initialErr)
		panic(panicValue)
	})
	t.Run("NotNilInitialErr+NilGenerator", func(t *testing.T) {
		var panicValue = struct{}{}
		basicErr := errors.New("BasicErr")
		var initialErr = basicErr
		defer func() {
			if panicOutcome := recover(); panicOutcome != nil {
				t.Errorf("Expected no panic, actual %#v", panicOutcome)
			}
			if initialErr == nil {
				t.Errorf("Expected not nil err, actual nil")
			} else if unwrappedErr := errors.Unwrap(initialErr); !errors.Is(unwrappedErr, basicErr) {
				t.Errorf("Expected %#v, actual %#v", basicErr, unwrappedErr)
			}
		}()
		defer PanicHandler("NotNilInitialErr+NilGenerator", GenerateError, nil, &initialErr)
		panic(panicValue)
	})
	t.Run("NotNilInitialErr+CodeGen", func(t *testing.T) {
		var panicValue = struct{}{}
		var basicErr = errors.New("InitErr")
		var initialErr = basicErr
		defer func() {
			if panicOutcome := recover(); panicOutcome != nil {
				t.Errorf("Expected no panic, actual %#v", panicOutcome)
			}
			if initialErr == nil {
				t.Errorf("Expected not nil initial err, actual nil")
			} else if unwrappedErr := errors.Unwrap(initialErr); !errors.Is(unwrappedErr, basicErr) {
				t.Errorf("Expected %#v, actual %#v", basicErr, unwrappedErr)
			}
		}()
		defer PanicHandler("NotNilInitialErr+CodeGen", GenerateError, simpleErrCodeGenerator, &initialErr)
		panic(panicValue)
	})
}

