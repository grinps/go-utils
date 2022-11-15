package errext

import "testing"

func TestNewErrorCodeWithOptions(t *testing.T) {
	t.Run("EmptyOptions", func(t *testing.T) {
		anErrCode := NewErrorCodeWithOptions()
		if anErrCode == nil {
			t.Error("Expected not nil, actual nil")
		}
	})
}

func TestWithErrorCode(t *testing.T) {
	t.Run("ErrorType-1", func(t *testing.T) {
		anErrCode := NewErrorCodeWithOptions(WithErrorCode(ErrorCodeNotSet))
		if anErrCode == nil {
			t.Error("Expected not nil, actual nil")
		}
	})
	t.Run("ErrorType100", func(t *testing.T) {
		anErrCode := NewErrorCodeWithOptions(WithErrorCode(100))
		if anErrCode == nil {
			t.Error("Expected not nil, actual nil")
		}
	})
	t.Run("ErrorType128", func(t *testing.T) {
		defer func() {
			if panicErr := recover(); panicErr == nil {
				t.Error("Expected not nil panic, actual nil")
			} else if asErr, isErr := panicErr.(error); !isErr {
				t.Errorf("Expected an error, actual %#v", panicErr)
			} else if _, isInvalid := ErrCodeInvalidErrorCode.AsError(asErr); !isInvalid {
				t.Errorf("Expected instance of ErrCodeInvalidErrorCode, Actual %#v", panicErr)
			}
		}()
		anErrCode := NewErrorCodeWithOptions(WithErrorCode(ErrorCodeValueStartValueForGeneration))
		if anErrCode == nil {
			t.Error("Expected not nil, actual nil")
		}
	})
}

func TestWithErrorType(t *testing.T) {
	t.Run("ValidValue", func(t *testing.T) {
		errCode := NewErrorCodeWithOptions(WithErrorType("ANewErrType"))
		if errCode == nil {
			t.Error("Expected not nil, actual nil")
		}
	})
	t.Run("MultipleOptions", func(t *testing.T) {
		errCode := NewErrorCodeWithOptions(WithErrorType("ANewErrType1"), WithErrorCode(10))
		if errCode == nil {
			t.Error("Expected not nil, actual nil")
		}

	})
}
