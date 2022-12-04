package errext

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

func TestWithTemplate(t *testing.T) {
	t.Run("WithTemplateAndNoValue", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate())
		if aErrCode == nil {
			t.Errorf("Expected Error Code not nil, Actual nil")
		}
		aErrWithText := aErrCode.New("AnErrorText")
		if aErrWithText == nil {
			t.Errorf("Expected Error text for AnErrorText as not nil, Actual nil")
		}
		if aErrWithText.Error() != "AnErrorText" {
			t.Errorf("Expected Error text as AnErrorText. Actual %#v", aErrWithText)
		}
		aErrWithTwoText := aErrCode.NewF("KEY1", "VALUE1")
		if aErrWithTwoText == nil {
			t.Errorf("Expected Error text for KEY1 VALUE1 as not nil, Actual nil")
		}
		if aErrWithTwoText.Error() != "Parameter0 = KEY1 Parameter1 = VALUE1" {
			t.Errorf("Expected Error text as Parameter0 = KEY1 Parameter1 = VALUE1. Actual %#v", aErrWithTwoText)
		}
		someErrToBubble := errors.New("AnError")
		aErrWithErrAndText := aErrCode.NewWithError("TEXT", someErrToBubble)
		if aErrWithErrAndText == nil {
			t.Errorf("Expected Error text for TEXT & AnError as not nil, Actual nil")
		}
		if errTxt := aErrWithErrAndText.Error(); errTxt != "TEXT" {
			t.Errorf("Expected Error text as TEXT. Actual %#v", errTxt)
		}
		if unwrappedErr := errors.Unwrap(aErrWithErrAndText); unwrappedErr != someErrToBubble {
			t.Errorf("Expected Error %#v. Actual %#v", someErrToBubble, unwrappedErr)

		}
		aErrWithErrAndArgs := aErrCode.NewWithErrorF(someErrToBubble, "KEY2", "VALUE2")
		if aErrWithErrAndArgs == nil {
			t.Errorf("Expected Error text for Err & KEY2 VALUE2 as not nil, Actual nil")
		}
		if aErrWithErrAndArgs.Error() != "Parameter0 = KEY2 Parameter1 = VALUE2" {
			t.Errorf("Expected Error text as Parameter0 = KEY2 Parameter1 = VALUE2. Actual %#v", aErrWithErrAndArgs)
		}
	})

	t.Run("WithTemplateOf1String", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("TEXT1"))
		if aErrCode == nil {
			t.Errorf("Expected Error Code not nil, Actual nil")
		}
		aErrWithText := aErrCode.New("AnErrorText")
		if aErrWithText == nil {
			t.Errorf("Expected Error text for AnErrorText as not nil, Actual nil")
		}
		if aErrWithText.Error() != "AnErrorText" {
			t.Errorf("Expected Error text as AnErrorText. Actual %#v", aErrWithText)
		}
		aErrWithTwoText := aErrCode.NewF("KEY1", "VALUE1")
		if aErrWithTwoText == nil {
			t.Errorf("Expected Error text for KEY1 VALUE1 as not nil, Actual nil")
		}
		if aErrWithTwoText.Error() != "TEXT1 Parameter0 = KEY1 Parameter1 = VALUE1" {
			t.Errorf("Expected Error text as TEXT1 Parameter0 = KEY1 Parameter1 = VALUE1. Actual %#v", aErrWithTwoText)
		}
		aErrWithTwoMixedText := aErrCode.NewF(9, "VALUE1")
		if aErrWithTwoMixedText == nil {
			t.Errorf("Expected Error text for 9 VALUE1 as not nil, Actual nil")
		}
		if aErrWithTwoMixedText.Error() != "TEXT1 Parameter0 = 9 Parameter1 = VALUE1" {
			t.Errorf("Expected Error text as TEXT1 Parameter0 = 9 Parameter1 = VALUE1. Actual %#v", aErrWithTwoMixedText)
		}
		someErrToBubble := errors.New("AnError")
		aErrWithErrAndText := aErrCode.NewWithError("TEXT", someErrToBubble)
		if aErrWithErrAndText == nil {
			t.Errorf("Expected Error text for TEXT & AnError as not nil, Actual nil")
		}
		if errTxt := aErrWithErrAndText.Error(); errTxt != "TEXT" {
			t.Errorf("Expected Error text as TEXT. Actual %#v", errTxt)
		}
		if unwrappedErr := errors.Unwrap(aErrWithErrAndText); unwrappedErr != someErrToBubble {
			t.Errorf("Expected Error %#v. Actual %#v", someErrToBubble, unwrappedErr)

		}
		aErrWithErrAndArgs := aErrCode.NewWithErrorF(someErrToBubble, "KEY2", "VALUE2")
		if aErrWithErrAndArgs == nil {
			t.Errorf("Expected Error text for Err & KEY2 VALUE2 as not nil, Actual nil")
		}
		if aErrWithErrAndArgs.Error() != "TEXT1 Parameter0 = KEY2 Parameter1 = VALUE2" {
			t.Errorf("Expected Error text as TEXT1 Parameter0 = KEY2 Parameter1 = VALUE2. Actual %#v", aErrWithErrAndArgs)
		}
	})

	t.Run("WithTemplateOf2String1Params", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("TEXT1", "[KEY1]"))
		if aErrCode == nil {
			t.Errorf("Expected Error Code not nil, Actual nil")
		}
		aErrWithText := aErrCode.New("AnErrorText")
		if aErrWithText == nil {
			t.Errorf("Expected Error text for AnErrorText as not nil, Actual nil")
		}
		if aErrWithText.Error() != "AnErrorText" {
			t.Errorf("Expected Error text as AnErrorText. Actual %#v", aErrWithText)
		}
		aErrWithTwoText := aErrCode.NewF("KEY1", "VALUE1")
		if aErrWithTwoText == nil {
			t.Errorf("Expected Error text for KEY1 VALUE1 as not nil, Actual nil")
		}
		if aErrWithTwoText.Error() != "TEXT1 VALUE1" {
			t.Errorf("Expected Error text as TEXT1 VALUE1. Actual %#v", aErrWithTwoText)
		}
		someErrToBubble := errors.New("AnError")
		aErrWithErrAndText := aErrCode.NewWithError("TEXT", someErrToBubble)
		if aErrWithErrAndText == nil {
			t.Errorf("Expected Error text for TEXT & AnError as not nil, Actual nil")
		}
		if errTxt := aErrWithErrAndText.Error(); errTxt != "TEXT" {
			t.Errorf("Expected Error text as TEXT. Actual %#v", errTxt)
		}
		if unwrappedErr := errors.Unwrap(aErrWithErrAndText); unwrappedErr != someErrToBubble {
			t.Errorf("Expected Error %#v. Actual %#v", someErrToBubble, unwrappedErr)

		}
		aErrWithErrAndArgs := aErrCode.NewWithErrorF(someErrToBubble, "KEY1", "VALUE2")
		if aErrWithErrAndArgs == nil {
			t.Errorf("Expected Error text for Err & KEY2 VALUE2 as not nil, Actual nil")
		}
		if aErrWithErrAndArgs.Error() != "TEXT1 VALUE2" {
			t.Errorf("Expected Error text as TEXT1 VALUE2. Actual %#v", aErrWithErrAndArgs)
		}
	})
	t.Run("WithTemplateOf3String1Params&ExtraParamValues", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("Error due to KEY1 =", "[KEY1]", " and other issues."))
		if aErrCode == nil {
			t.Errorf("Expected Error Code not nil, Actual nil")
		}
		aErrWithText := aErrCode.New("AnErrorText")
		if aErrWithText == nil {
			t.Errorf("Expected Error text for AnErrorText as not nil, Actual nil")
		}
		if aErrWithText.Error() != "AnErrorText" {
			t.Errorf("Expected Error text as AnErrorText. Actual %#v", aErrWithText)
		}
		aErrWithTwoText := aErrCode.NewF("KEY1", "VALUE1", "AnotherValue")
		if aErrWithTwoText == nil {
			t.Errorf("Expected Error text for KEY1 VALUE1 as not nil, Actual nil")
		}
		err1Txt := "Error due to KEY1 = VALUE1  and other issues. Parameter2 = AnotherValue"
		if aErrWithTwoText.Error() != err1Txt {
			t.Errorf("Expected Error text as %s. Actual %#v", err1Txt, aErrWithTwoText)
		}
		someErrToBubble := errors.New("AnError")
		aErrWithErrAndText := aErrCode.NewWithError("TEXT", someErrToBubble)
		if aErrWithErrAndText == nil {
			t.Errorf("Expected Error text for TEXT & AnError as not nil, Actual nil")
		}
		if errTxt := aErrWithErrAndText.Error(); errTxt != "TEXT" {
			t.Errorf("Expected Error text as TEXT. Actual %#v", errTxt)
		}
		if unwrappedErr := errors.Unwrap(aErrWithErrAndText); unwrappedErr != someErrToBubble {
			t.Errorf("Expected Error %#v. Actual %#v", someErrToBubble, unwrappedErr)

		}
		aErrWithErrAndArgs := aErrCode.NewWithErrorF(someErrToBubble, "KEY1", "VALUE2", "KEY2", "VALUE2")
		if aErrWithErrAndArgs == nil {
			t.Errorf("Expected Error text for Err & KEY2 VALUE2 as not nil, Actual nil")
		}
		errTxt2 := "Error due to KEY1 = VALUE2  and other issues. Parameter2 = KEY2 Parameter3 = VALUE2"
		if aErrWithErrAndArgs.Error() != errTxt2 {
			t.Errorf("Expected Error text as %s. Actual %#v", errTxt2, aErrWithErrAndArgs)
		}
	})
}

func TestNewParameter(t *testing.T) {
	t.Run("EmptyKeyNilValue", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "KEY1"))
		aErr := aErrCode.NewF(NewField("", nil))
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText KEY1 Parameter0 = <nil>"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
	t.Run("NotMatchingKeyNilValue", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "KEY1"))
		aErr := aErrCode.NewF(NewField("KEY2", nil))
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText KEY1 KEY2 = <nil>"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
	t.Run("NotMatchingKeyPrimitiveValue", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "KEY1"))
		aErr := aErrCode.NewF(NewField("KEY2", 8))
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText KEY1 KEY2 = 8"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
	t.Run("MatchingKeyPrimitiveValue", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "[KEY1]"))
		aErr := aErrCode.NewF(NewField("KEY1", 8))
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText 8"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
	t.Run("MatchingKeysAndMultiValue", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "[KEY1]", "MiddleText", "[KEY2]", "EndText"))
		aErr := aErrCode.NewF(NewField("KEY1", 8), "KEY2", []int{8, 9, 10}, "KEY3", 10)
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText 8 MiddleText [8 9 10] EndText Parameter3 = KEY3 Parameter4 = 10"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
	t.Run("MatchingKeysAndMultiValue", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "[KEY1]", "MiddleText", "[KEY2]", "EndText"))
		aErr := aErrCode.NewF(NewField("KEY1", 8), "KEY2", []int{8, 9, 10}, 9, 10)
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText 8 MiddleText [8 9 10] EndText Parameter3 = 9 Parameter4 = 10"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
	t.Run("MissingParamValue", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "[KEY1]", "MiddleText", "[KEY2]", "EndText"))
		aErr := aErrCode.NewF(NewField("KEY1", 8))
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText 8 MiddleText KEY2 EndText"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
	t.Run("DuplicateParams", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "[KEY1]", "MiddleText", "[KEY1]", "EndText"))
		aErr := aErrCode.NewF(NewField("KEY1", 8))
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText 8 MiddleText 8 EndText"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
	t.Run("DuplicateValuesForParam", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "[KEY1]", "MiddleText", "[KEY2]", "EndText"))
		aErr := aErrCode.NewF(NewField("KEY2", 15), "KEY1", 8, []int{10}, NewField("KEY1", 10))
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText 10 MiddleText 15 EndText Parameter3 = [10]"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
	t.Run("DuplicateValuesForParamReverse", func(t *testing.T) {
		aErrCode := NewErrorCodeWithOptions(WithTemplate("StartText", "[KEY1]", "MiddleText", "[KEY2]", "EndText"))
		aErr := aErrCode.NewF(NewField("KEY1", 15), "KEY2", 8, NewField("KEY1", 10))
		if aErr == nil {
			t.Errorf("Expected not nil err, Actual nil")
		}
		expectedErrMsg := "StartText 10 MiddleText 8 EndText"
		if errMsg := aErr.Error(); errMsg != expectedErrMsg {
			t.Errorf("Expected error message %s, Actual %s", expectedErrMsg, errMsg)
		}
	})
}

func TestPhraseImpl_Equals(t *testing.T) {
	aErrCode := NewErrorCodeWithOptions(WithTemplate("Phrase1", "[Param1]", "[Phrase1]"))
	if asErrCodeImpl, isErrCodeImpl := aErrCode.(*ErrorCodeImpl); isErrCodeImpl {
		if !asErrCodeImpl.template.phrases[0].Equals(phraseImpl("Phrase1")) {
			t.Errorf("Expected phrase %#v to be match a phrase Phrase1", asErrCodeImpl.template.phrases[0])
		}
		if !asErrCodeImpl.template.phrases[1].Equals(paramDef("Param1")) {
			t.Errorf("Expected parameter %#v to be match a parameter Param1", asErrCodeImpl.template.phrases[1])
		}
		if !asErrCodeImpl.template.phrases[2].Equals(asErrCodeImpl.template.phrases[0]) {
			t.Errorf("Expected phrase %#v to match a parameter %#v", asErrCodeImpl.template.phrases[2], asErrCodeImpl.template.phrases[0])
		}
		if asErrCodeImpl.template.phrases[1].Equals(asErrCodeImpl.template.phrases[0]) {
			t.Errorf("Expected phrase %#v to not match a parameter %#v", asErrCodeImpl.template.phrases[1], asErrCodeImpl.template.phrases[0])
		}
	} else {
		t.Errorf("Expected type *ErrorCodeImpl, Actual %#v", aErrCode)
	}
}

func Example() {
	var aFileProcessingErrorCode = NewErrorCodeWithOptions(WithErrorType("FileProcessing"), WithTemplate("Failure to read file", "[FILENAME]", "due to error", "[READERR]"))
	var AFunctionReturningError = func(fileName string) (returnFile *os.File, returnErr error) {
		var fileOpenErr error
		if returnFile, fileOpenErr = os.Open(fileName); fileOpenErr != nil {
			returnErr = aFileProcessingErrorCode.NewWithErrorF(fileOpenErr, NewField("FILENAME", fileName), "READERR", fileOpenErr)
		}
		return
	}
	_, err := AFunctionReturningError("./SomeInvalidFileName")
	if _, isErr := aFileProcessingErrorCode.AsError(err); isErr {
		fmt.Printf("Error while opening file. %s\n", err.Error())
	}
	// Output: Error while opening file. Failure to read file ./SomeInvalidFileName due to error open ./SomeInvalidFileName: no such file or directory
}
