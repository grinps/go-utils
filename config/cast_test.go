package config

import (
	"context"
	"errors"
	"testing"
)

func TestGetAsMap(t *testing.T) {
	t.Run("AllNil", func(t *testing.T) {
		output, outErr := GetAsMap(nil, nil, nil)
		if outErr != nil {
			t.Errorf("expected no error, actual error.")
		}
		if output != nil {
			t.Errorf("expected no output, actual %#v", output)
		}
	})
	var defaultConfigMap = map[string]any{}
	var defaultConfig = NewConfig[SimpleConfigDriver, SimpleConfig](context.Background(), OptionSimpleConfigWithMap(defaultConfigMap))
	extErr := ExtensibleErr(context.Background(), defaultConfig, "RandomKey")
	t.Run("InputNil", func(t *testing.T) {
		output, outErr := GetAsMap(context.Background(), nil, extErr)
		if outErr != nil {
			t.Errorf("expected no error, actual error.")
		}
		if output != nil {
			t.Errorf("expected no output, actual %#v", output)
		}
	})
	t.Run("EmptyJSON", func(t *testing.T) {
		output, outErr := GetAsMap(context.Background(), "{}", extErr)
		if outErr != nil {
			t.Errorf("expected no error actual %s", outErr.Error())
		}
		if output == nil {
			t.Errorf("expected a config actual nil")
		}
	})
	t.Run("ValidJSON", func(t *testing.T) {
		output, outErr := GetAsMap(context.Background(), "{\"key1\": \"val1\", \"key2\": {\"Key21\": true} }", extErr)
		if outErr != nil {
			t.Errorf("expected no error actual %s", outErr.Error())
		}
		if output == nil {
			t.Errorf("expected a config actual nil")
		}
	})
	t.Run("InvalidJSON", func(t *testing.T) {
		output, outErr := GetAsMap(context.Background(), "{\"test\": {}", extErr)
		if outErr == nil {
			t.Errorf("expected error actual no error")
		}
		if output != nil {
			t.Errorf("expected no config actual %#v", output)
		}
	})
	t.Run("map[string]stringValid", func(t *testing.T) {
		var t1 map[string]string = map[string]string{"K1": "V1", "K2": "v2"}
		output, outErr := GetAsMap(context.Background(), t1, extErr)
		if outErr != nil {
			t.Errorf("expected no error actual %s", outErr)
		}
		if output == nil {
			t.Errorf("expected valid output, actual nil")
		}
		if k1Val := output["K1"]; k1Val != "V1" {
			t.Errorf("Expected K1=V1, actual %#v", k1Val)
		}
	})
	t.Run("map[string]stringEmpty", func(t *testing.T) {
		var t1 map[string]string = map[string]string{}
		output, outErr := GetAsMap(context.Background(), t1, extErr)
		if outErr != nil {
			t.Errorf("expected no error actual %s", outErr)
		}
		if output == nil {
			t.Errorf("expected valid output, actual nil")
		}
	})
	t.Run("map[any]anyEmpty", func(t *testing.T) {
		var t1 map[any]any = map[any]any{}
		output, outErr := GetAsMap(context.Background(), t1, extErr)
		if outErr != nil {
			t.Errorf("expected no error actual %s", outErr)
		}
		if output == nil {
			t.Errorf("expected valid output, actual nil")
		}
		if len(output) != 0 {
			t.Errorf("expected empty map actual %d", len(output))
		}
	})
	t.Run("map[any]anyValid", func(t *testing.T) {
		var stringKey = "K4"
		var ptrToStringKey = &stringKey
		var t1 map[any]any = map[any]any{"K1": "V1", "K2": map[string]any{"K21": 1, "K22": true},
			"K3": nil, ptrToStringKey: "V4"}
		output, outErr := GetAsMap(context.Background(), t1, extErr)
		if outErr != nil {
			t.Errorf("expected no error actual %s", outErr)
		}
		if output == nil {
			t.Errorf("expected valid output, actual nil")
		}
		if len(output) != 4 {
			t.Errorf("expected 4 actual %d", len(output))
		}
	})
	t.Run("map[any]anyInvalidNil&InvalidKeys", func(t *testing.T) {
		var t1 map[any]any = map[any]any{"K1": "V1", nil: map[string]any{"K21": 1, "K22": true}, 1: true, "K3": nil}
		output, outErr := GetAsMap(context.Background(), t1, extErr)
		if outErr == nil {
			t.Errorf("expected error actual no error")
		} else {
			ValidateErr(t, outErr, "RandomKey", ErrConfigInvalidValue,
				"unable to cast 1 of type int to string",
				"failed to add key <nil> with value map[K21:1 K22:true] to map since key is nil")
		}
		if output != nil {
			t.Errorf("expected no output, actual %#v", output)
		}
	})
	t.Run("map[string]anyNilValue", func(t *testing.T) {
		var t1 map[string]any
		output, outErr := GetAsMap(context.Background(), t1, extErr)
		if outErr != nil {
			t.Errorf("expected no error actual error %s", outErr.Error())
		}
		if output != nil {
			t.Errorf("expected no output, actual %#v", output)
		}
	})
	t.Run("map[string]anyEmptyValue", func(t *testing.T) {
		var t1 map[string]any = map[string]any{}
		output, outErr := GetAsMap(context.Background(), t1, extErr)
		if outErr != nil {
			t.Errorf("expected no error actual error %s", outErr.Error())
		}
		if output == nil {
			t.Errorf("expected output, actual nil")
		}
		if len(output) != 0 {
			t.Errorf("expected 0 actual %d", len(output))
		}
	})
	t.Run("map[string]anyValidValue", func(t *testing.T) {
		var t1 map[string]any = map[string]any{"K1": "V1", "K2": map[string]any{"K21": "V21"}}
		output, outErr := GetAsMap(context.Background(), t1, extErr)
		if outErr != nil {
			t.Errorf("expected no error actual error %s", outErr.Error())
		}
		if output == nil {
			t.Errorf("expected output, actual no output")
		}
		if len(output) != 2 {
			t.Errorf("expected 2 actual %d", len(output))
		}
	})
	t.Run("nonMapValue", func(t *testing.T) {
		output, outErr := GetAsMap(context.Background(), 1, extErr)
		if outErr == nil {
			t.Errorf("expected error actual no error")
		} else {
			ValidateErr(t, outErr, "RandomKey", ErrConfigInvalidValue,
				"conversion of configuration 1 of type int to map[string]any not supported")
		}
		if output != nil {
			t.Errorf("expected no output, actual %#v", output)
		}
	})
}

func TestToString(t *testing.T) {
	asString, strErr := toString(nil)
	if strErr == nil {
		t.Errorf("expected error actual no error")
	}
	if asString != "" {
		t.Errorf("Expected empty string, actual %s", asString)
	}
}

func ValidateErr(t *testing.T, outErr error, keyName string, errCode ErrCode, messages ...string) {
	var functionalErr FunctionalErr
	switch {
	case errors.As(outErr, &functionalErr):
		if keyName != "" {
			if functionalErr.Key() != keyName {
				t.Errorf("Expected key %s actual %s", keyName, functionalErr.Key())
			}
		}
		if !errors.Is(errCode, ErrConfigCodeUnknown) {
			if !errors.Is(functionalErr, errCode) {
				t.Errorf("Expected error code %v actual %v", errCode, functionalErr.Code())
			}
		}
		if len(messages) > 0 {
			var applErr = functionalErr
			var msgToMatch = len(messages)
			allMatched := make([]bool, msgToMatch)
			matchWrappedMessage(applErr, messages, &allMatched)
			for matchIndex, matchResult := range allMatched {
				if !matchResult {
					t.Errorf("Expected error message %s could not be located, error %s", messages[matchIndex], applErr.Error())
				}
			}
		}
	default:
		t.Errorf("Expected error of type FunctionalErr actual %T", outErr)
	}
}

func matchWrappedMessage(inputErr error, messages []string, allMatched *[]bool) {
	if inputErr != nil {
		inputErrMsg := inputErr.Error()
		for msgIndex, msg := range messages {
			if msg == inputErrMsg {
				(*allMatched)[msgIndex] = true
			}
		}
	}
	switch x := inputErr.(type) {
	case interface{ Unwrap() error }:
		{
			processErr := x.Unwrap()
			if processErr != nil {
				matchWrappedMessage(processErr, messages, allMatched)
			}
		}
	case interface{ Unwrap() []error }:
		{
			processErrs := x.Unwrap()
			for _, processErr := range processErrs {
				matchWrappedMessage(processErr, messages, allMatched)
			}
		}
	}
}
