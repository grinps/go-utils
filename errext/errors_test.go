package errext

import (
	"errors"
	"testing"
)

func TestErrorCodeImpl(t *testing.T) {
	t.Run("NilInstance", func(t *testing.T) {
		var nilInstance *ErrorCodeImpl = nil
		nilErrorTest := nilInstance.New("Test")
		if errValue := nilErrorTest.Error(); errValue != "Test" {
			t.Error("For test error created, Expected Test, Actual", errValue)
		}
		if asErr, isErr := defaultErrorCode.AsError(nilErrorTest); !isErr {
			t.Errorf("For test error created, Expected instance of %#v, Actual %#v", defaultErrorCode, nilErrorTest)
		} else if asErr.errorCode != defaultErrorCode {
			t.Errorf("For test error created, Expected errorCode to be %#v, Actual %#v", defaultErrorCode, asErr.errorCode)
		} else if asErr.err != nil {
			t.Errorf("For test error created, Expected wrapped error to be nil, Actual %#v", asErr.err)
		}
		testWithArgErr := nilInstance.NewF("Test1", "test2")
		if errValue := testWithArgErr.Error(); errValue != "Test1 test2" {
			t.Error("For test error created, Expected Test1 test2, Actual>", errValue, "<")
		}
		if asErr, isErr := defaultErrorCode.AsError(nilErrorTest); !isErr {
			t.Errorf("For test error created, Expected instance of %#v, Actual %#v", defaultErrorCode, nilErrorTest)
		} else if asErr.errorCode != defaultErrorCode {
			t.Errorf("For test error created, Expected errorCode to be %#v, Actual %#v", defaultErrorCode, asErr.errorCode)
		} else if asErr.err != nil {
			t.Errorf("For test error created, Expected wrapped error to be nil, Actual %#v", asErr.err)
		}
	})
	t.Run("DefaultObject", func(t *testing.T) {
		var nilInstance *ErrorCodeImpl = &ErrorCodeImpl{}
		nilErrorTest := nilInstance.New("Test")
		if errValue := nilErrorTest.Error(); errValue != errorCodeNotSetMessage {
			t.Errorf("For test error created, Expected %s, Actual %s", errorCodeNotSetMessage, errValue)
		}
		if asErr, isErr := defaultErrorCode.AsError(nilErrorTest); isErr {
			t.Errorf("For test error created, Expected not an instance of error %s, Actual %#v", errorCodeNotSetMessage, asErr)
		}
		testWithArgErr := nilInstance.NewF("Test1", "test2")
		if errValue := testWithArgErr.Error(); errValue != errorCodeNotSetMessage {
			t.Errorf("For test error created, Expected %s, Actual %s", errorCodeNotSetMessage, errValue)
		}
	})
}

func TestNewErrorCode(t *testing.T) {
	t.Run("BigNegativeValue", func(t *testing.T) {
		defer func() {
			if panicRecover := recover(); panicRecover == nil {
				t.Errorf("Expecting a panic, Actually no panic")
			} else if asErr, isErr := panicRecover.(error); !isErr {
				t.Errorf("Expected an error object, actual %T (%#v)", asErr, asErr)
			} else if _, isErr := ErrCodeInvalidErrorCode.AsError(asErr); !isErr {
				t.Errorf("Expected Error Code ErrCodeInvalidErrorCode, Actual %T (%#v)", asErr, asErr)
			}
		}()
		errorCode := NewErrorCode(-10)
		text1Err := errorCode.New("Text1")
		if text1Err.Error() != "Text1" {
			t.Errorf("Expected Text1, Actual %v", text1Err)
		}
		textArgsErr := errorCode.NewF("Text1", "Text2")
		if textArgsErr.Error() != "Text1 Text2" {
			t.Errorf("Expected Text1 Text2, Actual %v", text1Err)
		}
		anErr := errors.New("AnErr")
		textWithErr := errorCode.NewWithErrorF(anErr, "Text2", "Text3")
		if textWithErr.Error() != "Text2 Text3" {
			t.Errorf("Expected Text2 Text3, Actual %v", textWithErr)
		}
		if unwrappedErr := errors.Unwrap(textWithErr); unwrappedErr != anErr {
			t.Errorf("Expected %v, Actual %v", anErr, unwrappedErr)
		}
	})
	t.Run("O-2", func(t *testing.T) {
		defer func() {
			if panicRecover := recover(); panicRecover == nil {
				t.Errorf("Expecting a panic, Actually no panic")
			} else if asErr, isErr := panicRecover.(error); !isErr {
				t.Errorf("Expected an error object, actual %T (%#v)", asErr, asErr)
			} else if _, isErr := ErrCodeInvalidErrorCode.AsError(asErr); !isErr {
				t.Errorf("Expected Error Code ErrCodeInvalidErrorCode, Actual %T (%#v)", asErr, asErr)
			}
		}()
		errorCode := NewErrorCode(-2)
		text1Err := errorCode.New("Text1")
		if text1Err.Error() != "Text1" {
			t.Errorf("Expected Text1, Actual %v", text1Err)
		}
		textArgsErr := errorCode.NewF("Text1", "Text2")
		if textArgsErr.Error() != "Text1 Text2" {
			t.Errorf("Expected Text1 Text2, Actual %v", text1Err)
		}
		anErr := errors.New("AnErr")
		textWithErr := errorCode.NewWithError("Text2", anErr)
		if textWithErr.Error() != "Text2" {
			t.Errorf("Expected Text2, Actual %v", textWithErr)
		}
		if unwrappedErr := errors.Unwrap(textWithErr); unwrappedErr != anErr {
			t.Errorf("Expected %v, Actual %v", anErr, unwrappedErr)
		}
	})
	t.Run("O-1", func(t *testing.T) {
		defer func() {
			if panicRecover := recover(); panicRecover != nil {
				t.Errorf("Expecting no panic, Actually panic %#v", panicRecover)
			}
		}()
		errorCode := NewErrorCode(-1)
		text1Err := errorCode.New("Text1")
		if text1Err.Error() != "Text1" {
			t.Errorf("Expected Text1, Actual %v", text1Err)
		}
		textArgsErr := errorCode.NewF("Text1", "Text2")
		if textArgsErr.Error() != "Text1 Text2" {
			t.Errorf("Expected Text1 Text2, Actual %v", text1Err)
		}
		anErr := errors.New("AnErr")
		textWithErr := errorCode.NewWithError("Text2", anErr)
		if textWithErr.Error() != "Text2" {
			t.Errorf("Expected Text2, Actual %v", textWithErr)
		}
		if unwrappedErr := errors.Unwrap(textWithErr); unwrappedErr != anErr {
			t.Errorf("Expected %v, Actual %v", anErr, unwrappedErr)
		}
	})
	t.Run("O", func(t *testing.T) {
		defer func() {
			if panicRecover := recover(); panicRecover != nil {
				t.Errorf("Expecting no panic, Actually panic %#v", panicRecover)
			}
		}()
		errorCode := NewErrorCode(0)
		text1Err := errorCode.New("Text1")
		if text1Err.Error() != "Text1" {
			t.Errorf("Expected Text1, Actual %v", text1Err)
		}
		textArgsErr := errorCode.NewF("Text1", "Text2")
		if textArgsErr.Error() != "Text1 Text2" {
			t.Errorf("Expected Text1 Text2, Actual %v", text1Err)
		}
		anErr := errors.New("AnErr")
		textWithErr := errorCode.NewWithError("Text2", anErr)
		if textWithErr.Error() != "Text2" {
			t.Errorf("Expected Text2, Actual %v", textWithErr)
		}
		if unwrappedErr := errors.Unwrap(textWithErr); unwrappedErr != anErr {
			t.Errorf("Expected %v, Actual %v", anErr, unwrappedErr)
		}
	})
	t.Run("1", func(t *testing.T) {
		defer func() {
			if panicRecover := recover(); panicRecover != nil {
				t.Errorf("Expecting no panic, Actually panic %#v", panicRecover)
			}
		}()
		errorCode := NewErrorCode(1)
		text1Err := errorCode.New("Text1")
		if text1Err.Error() != "Text1" {
			t.Errorf("Expected Text1, Actual %v", text1Err)
		}
		textArgsErr := errorCode.NewF("Text1", "Text2")
		if textArgsErr.Error() != "Text1 Text2" {
			t.Errorf("Expected Text1 Text2, Actual %v", text1Err)
		}
		anErr := errors.New("AnErr")
		textWithErr := errorCode.NewWithError("Text2", anErr)
		if textWithErr.Error() != "Text2" {
			t.Errorf("Expected Text2, Actual %v", textWithErr)
		}
		if unwrappedErr := errors.Unwrap(textWithErr); unwrappedErr != anErr {
			t.Errorf("Expected %v, Actual %v", anErr, unwrappedErr)
		}
	})
	t.Run("BaseValue", func(t *testing.T) {
		defer func() {
			if panicRecover := recover(); panicRecover == nil {
				t.Errorf("Expecting a panic, Actually no panic")
			} else if asErr, isErr := panicRecover.(error); !isErr {
				t.Errorf("Expected an error object, actual %T (%#v)", asErr, asErr)
			} else if _, isErr := ErrCodeInvalidErrorCode.AsError(asErr); !isErr {
				t.Errorf("Expected Error Code ErrCodeInvalidErrorCode, Actual %T (%#v)", asErr, asErr)
			}
		}()
		errorCode := NewErrorCode(int(ErrorCodeValueStartValueForGeneration))
		text1Err := errorCode.New("Text1")
		if text1Err.Error() != "Text1" {
			t.Errorf("Expected Text1, Actual %v", text1Err)
		}
		textArgsErr := errorCode.NewF("Text1", "Text2")
		if textArgsErr.Error() != "Text1 Text2" {
			t.Errorf("Expected Text1 Text2, Actual %v", text1Err)
		}
		anErr := errors.New("AnErr")
		textWithErr := errorCode.NewWithError("Text2", anErr)
		if textWithErr.Error() != "Text2" {
			t.Errorf("Expected Text2, Actual %v", textWithErr)
		}
		if unwrappedErr := errors.Unwrap(textWithErr); unwrappedErr != anErr {
			t.Errorf("Expected %v, Actual %v", anErr, unwrappedErr)
		}
	})
	t.Run("BaseValue-1", func(t *testing.T) {
		defer func() {
			if panicRecover := recover(); panicRecover != nil {
				t.Errorf("Expecting no panic, Actually panic %#v", panicRecover)
			}
		}()
		errorCode := NewErrorCode(int(ErrorCodeValueStartValueForGeneration - 1))
		text1Err := errorCode.New("Text1")
		if text1Err.Error() != "Text1" {
			t.Errorf("Expected Text1, Actual %v", text1Err)
		}
		textArgsErr := errorCode.NewF("Text1", "Text2")
		if textArgsErr.Error() != "Text1 Text2" {
			t.Errorf("Expected Text1 Text2, Actual %v", text1Err)
		}
		anErr := errors.New("AnErr")
		textWithErr := errorCode.NewWithError("Text2", anErr)
		if textWithErr.Error() != "Text2" {
			t.Errorf("Expected Text2, Actual %v", textWithErr)
		}
		if unwrappedErr := errors.Unwrap(textWithErr); unwrappedErr != anErr {
			t.Errorf("Expected %v, Actual %v", anErr, unwrappedErr)
		}
	})
}

func TestNewErrorCodeOfType(t *testing.T) {
	t.Run("ValidErrCodeEmptyType", func(t *testing.T) {
		errCode := NewErrorCodeOfType(1, "")
		if errCode == nil {
			t.Error("Expected an error code, but nil was returned.")
		} else {
			someErr := errors.New("AnError")
			anErr := errCode.NewWithErrorF(someErr, "Test1", "Test2")
			if anErr == nil {
				t.Error("Expected error object to be created, actual nil")
			} else if _, isErr := errCode.AsError(anErr); !isErr {
				t.Errorf("Expceted error object matching errCode, Actual %#v", anErr)
			}
		}
	})
	t.Run("NotSetErrCodeEmptyType", func(t *testing.T) {
		errCode := NewErrorCodeOfType(-1, "")
		if errCode == nil {
			t.Error("Expected an error code, but nil was returned.")
		} else {
			someErr := errors.New("AnError")
			anErr := errCode.NewWithErrorF(someErr, "Test1", "Test2")
			if anErr == nil {
				t.Error("Expected error object to be created, actual nil")
			} else if _, isErr := errCode.AsError(anErr); !isErr {
				t.Errorf("Expceted error object matching errCode, Actual %#v", anErr)
			}
		}
	})
	t.Run("ValidValues", func(t *testing.T) {
		errCode := NewErrorCodeOfType(1, "SomeRandomType")
		if errCode == nil {
			t.Error("Expected an error code, but nil was returned.")
		} else {
			someErr := errors.New("AnError")
			anErr := errCode.NewWithErrorF(someErr, "Test1", "Test2")
			if anErr == nil {
				t.Error("Expected error object to be created, actual nil")
			} else if _, isErr := errCode.AsError(anErr); !isErr {
				t.Errorf("Expceted error object matching errCode, Actual %#v", anErr)
			}
		}
	})
	t.Run("ValidNotUnique", func(t *testing.T) {
		errCode1 := NewErrorCodeOfType(1, "SomeRandomType")
		errCode1Err := errCode1.New("errCode1Err")
		errCode2 := NewErrorCodeOfType(1, "SomeRandomType")
		if errCode1 == errCode2 {
			t.Errorf("Expecting non-unique error codes, actual same %#v", errCode1)
		} else if _, isValid := errCode2.AsError(errCode1Err); isValid {
			t.Errorf("Expecting error created from error codes with same error id & type to not match Actual, matched %#v, %#v", errCode1Err, errCode2)
		}
	})
}

func TestNewUniqueErrorCode(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		uniqueErr1 := NewUniqueErrorCode(2)
		uniqueErr2 := NewUniqueErrorCode(2)
		if uniqueErr1 != uniqueErr2 {
			t.Errorf("Expecting two unique error of same error code to match, actual not. Err1 %#v, Err2 %#v", uniqueErr1, uniqueErr2)
		}
	})
}

func TestNewUniqueErrorCodeOfType(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		uniqueErr1 := NewUniqueErrorCodeOfType(1, "JunkType")
		uniqueErr2 := NewUniqueErrorCodeOfType(1, "JunkType")
		if uniqueErr1 != uniqueErr2 {
			t.Errorf("Expecting two unique error of same error code to match, actual not. Err1 %#v, Err2 %#v", uniqueErr1, uniqueErr2)
		}
	})
	t.Run("-1Values", func(t *testing.T) {
		uniqueErr1 := NewUniqueErrorCodeOfType(-1, "JunkType1")
		uniqueErr2 := NewUniqueErrorCodeOfType(-1, "JunkType1")
		if uniqueErr1 == uniqueErr2 {
			t.Errorf("Expecting two unique error of same error code to not match, actual it does. Err1 %#v, Err2 %#v", uniqueErr1, uniqueErr2)
		}
	})
}
