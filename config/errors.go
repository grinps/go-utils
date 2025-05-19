package config

import (
	"context"
	"fmt"
	"html/template"
	"strconv"
	"strings"
)

type ErrCode int

func (err ErrCode) Error() string {
	return "ErrorCode: " + strconv.Itoa(int(err))
}

const (
	ErrConfigCodeUnknown ErrCode = iota
	ErrConfigMissingValue
	ErrConfigEmptyKey
	ErrConfigInvalidKey
	ErrConfigInvalidConfig
	ErrConfigNilConfig
	ErrConfigMessageParsingFailed
	ErrConfigInvalidValueType
	ErrConfigKeyParsingFailed
	ErrConfigInvalidValue
)

var templateRepository = template.New("configErrorTemplates")
var errorCodeStringMap = map[ErrCode]*template.Template{
	ErrConfigCodeUnknown:          registerNewError(ErrConfigCodeUnknown, "unknown error occurred during config operation"),
	ErrConfigMissingValue:         registerNewError(ErrConfigMissingValue, "missing value for given key {{.Key}}"),
	ErrConfigEmptyKey:             registerNewError(ErrConfigEmptyKey, "given key({{.Key}}) is empty string"),
	ErrConfigInvalidKey:           registerNewError(ErrConfigInvalidKey, "given key({{.Key}}) is not of supported type"),
	ErrConfigKeyParsingFailed:     registerNewError(ErrConfigKeyParsingFailed, "failed to parse given key({{.Key}}) with valid result"),
	ErrConfigInvalidConfig:        registerNewError(ErrConfigInvalidConfig, "given config({{.Config}})  is not of supported type"),
	ErrConfigNilConfig:            registerNewError(ErrConfigNilConfig, "no config is provided"),
	ErrConfigMessageParsingFailed: registerNewError(ErrConfigMessageParsingFailed, "failed to parse configuration message"),
	ErrConfigInvalidValueType:     registerNewError(ErrConfigInvalidValueType, "retrieved value {{.Value}} for key {{.Key}} does not match requested type"),
	ErrConfigInvalidValue:         registerNewError(ErrConfigInvalidValue, "failed to retrieved value {{.Value}} for key {{.Key}} due to error {{.Cause}}"),
}

func RegisterNewError(errorCode ErrCode, errorMessage string) error {
	errTemplate := registerNewError(errorCode, errorMessage)
	if errTemplate != nil {
		errorCodeStringMap[errorCode] = errTemplate
	}
	return nil
}

func registerNewError(errorCode ErrCode, errorMessage string) *template.Template {
	parsedTemplate, err := templateRepository.Parse(errorMessage)
	if err != nil {
		parsedTemplate, _ = templateRepository.Parse(fmt.Sprintf("template parsing failed for error code %d with message %s with error %#v", errorCode, errorMessage, err))
	}
	return parsedTemplate
}

type FunctionalErr func(callType ErrCode, parameters ...any) any

func (err FunctionalErr) Error() string {
	return err(1).(string)
}

func (err FunctionalErr) Key() string {
	return err(2).(string)
}

func (err FunctionalErr) Is(target error) bool {
	return err(3, target).(bool)
}

func (err FunctionalErr) Code() ErrCode {
	return err(0).(ErrCode)
}

func (err FunctionalErr) Unwrap() error {
	return err(4).(error)
}

type errArguments struct {
	Context   context.Context
	ErrorCode ErrCode
	Config    Config
	Key       string
	Value     any
	Cause     error
}

var FuncErr = func(input errArguments) FunctionalErr {
	return FunctionalErr(func(callType ErrCode, parameters ...any) any {
		var parameter = input
		switch callType {
		case 0:
			return parameter.ErrorCode
		case 1:
			errMessageBuilder := &strings.Builder{}
			messageErr := errorCodeStringMap[parameter.ErrorCode].Execute(errMessageBuilder, parameter)
			if messageErr != nil {
				return errorCodeStringMap[parameter.ErrorCode].DefinedTemplates()
			} else {
				return errMessageBuilder.String()
			}
		case 2:
			return parameter.Key
		case 3:
			if len(parameters) >= 1 {
				switch parameters[0].(type) {
				case FunctionalErr:
					if parameters[0].(FunctionalErr).Code() == parameter.ErrorCode {
						return true
					}
				case ErrCode:
					if parameters[0].(ErrCode) == parameter.ErrorCode {
						return true
					}
				}
			}
			return false
		case 4:
			if parameter.Cause != nil {
				return parameter.Cause
			}
		}
		return "<unsupported case>"
	})
}

func ErrWithParameters(ctx context.Context, errorCode ErrCode, config Config, key string, value any, cause error) FunctionalErr {
	var templateValues = errArguments{
		Context:   ctx,
		ErrorCode: errorCode,
		Config:    config,
		Key:       key,
		Value:     value,
		Cause:     cause,
	}
	return FuncErr(templateValues)
}

type ExtensibleErrFunc func(options ...ErrOption) FunctionalErr

var ExtensibleErr = func(ctx context.Context, config Config, key string) ExtensibleErrFunc {
	var templateValues = errArguments{
		Context: ctx,
		Config:  config,
		Key:     key,
	}
	return func(options ...ErrOption) FunctionalErr {
		var inputValue = templateValues
		for _, option := range options {
			option(&inputValue)
		}
		return FuncErr(inputValue)
	}
}

type ErrOption func(input *errArguments)

func WithErrorCode(errCode ErrCode) ErrOption {
	return func(input *errArguments) {
		input.ErrorCode = errCode
	}
}

func WithCause(cause error) ErrOption {
	return func(input *errArguments) {
		input.Cause = cause
	}
}

func WithValue(value any) ErrOption {
	return func(input *errArguments) {
		input.Value = value
	}
}
