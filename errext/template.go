package errext

import (
	"fmt"
	"strconv"
	"strings"
)

// Phrase defines interface that any object that is defined as phrase in error template must implement.
//
// [fmt.Stringer] interface provides [fmt.Stringer].String method that returns the Phrase value as string.
// [errext.Phrase.Equals] returns true if the two phrases match. In case a [errext.Parameter] has same name
// as Phrase the two will match.
type Phrase interface {
	fmt.Stringer
	Equals(phrase Phrase) bool
}

// Parameter is a special type of [errext.Phrase] that describes additional detail/context to be added to error
// to provide runtime context of the error.
//
// [errext.Parameter].Process allows parameter implementation to handle how the particular type of Parameter
// will transform the value before adding to error text. For example a SecureParameter can hash the value to prevent
// printing sensitive data.
//
// Also see [errext.WithTemplate]
type Parameter interface {
	Phrase
	Process(value interface{}) interface{}
}

type phraseImpl string

func (p phraseImpl) String() string {
	return string(p)
}

func (p phraseImpl) Equals(phrase Phrase) bool {
	return phrase.String() == p.String()
}

type paramDef = phraseImpl

func (p paramDef) Process(value interface{}) interface{} {
	return value
}

type errTemplate struct {
	phrases    []Phrase
	parameters map[string]Parameter
}

// WithTemplate setups the [errext.ErrorCode] with error message template to reduce text duplication
// while creating error message.
//
// It expects the parameters to be highlighted by prefixing [ and suffixing ]. So that in an error phrase
// WithTemplate("Failed due to error", "[ERR]"), the "Failed due to error" will be identified as a Phrase
// while "ERR" will be identified as parameter to be provided while creating error (through [errext.ErrorCodeImpl.NewF] or
// [errext.ErrorCodeImpl.NewWithErrorF].
//
// The current version supports default [errext.Parameter] implementation that does not transform value.
func WithTemplate(phrases ...string) ErrorCodeOptions {
	errorTemplate := &errTemplate{
		phrases:    []Phrase{},
		parameters: map[string]Parameter{},
	}
	for _, parameter := range phrases {
		applicablePhrase := parameter
		var applicablePhraseImpl phraseImpl
		if strings.HasPrefix(parameter, "[") && strings.HasSuffix(parameter, "]") {
			//TODO: Add support for new prefixes and associated external parameter implementations e.g. secure value
			applicablePhrase = strings.Trim(parameter, "[]")
			applicablePhraseImpl = paramDef(applicablePhrase)
			errorTemplate.parameters[applicablePhraseImpl.String()] = applicablePhraseImpl
		} else {
			applicablePhraseImpl = phraseImpl(parameter)
		}
		errorTemplate.phrases = append(errorTemplate.phrases, applicablePhraseImpl)
	}
	return func(errorCode *ErrorCodeImpl) *ErrorCodeImpl {
		if errorCode != nil {
			errorCode.template = errorTemplate
		}
		return errorCode
	}
}

// Field provides a simple way to package key value pair to pass parameter value while creating error.
type Field struct {
	Key   string
	Value interface{}
}

// NewField creates a new Field to be passed to error creation calls like [errext.ErrorCodeImpl.NewF]
func NewField(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

func generateFromTemplate(template *errTemplate, args ...interface{}) []interface{} {
	var returnArguments []interface{}
	if template != nil {
		parameterValues := map[Parameter]interface{}{}
		var extraParameters []Parameter
		skipNext := false
		for argumentIndex, arg := range args {
			if skipNext {
				skipNext = false
				continue
			}
			var applicableKey Parameter
			var applicableValue interface{}
			if argAsString, isString := arg.(string); isString && argumentIndex < len(args)-1 {
				if asParameter, isValidParameter := template.parameters[argAsString]; isValidParameter { // name=value pairs
					applicableKey = asParameter
					applicableValue = args[argumentIndex+1]
					skipNext = true
				} else {
					applicableKey = paramDef("Parameter" + strconv.Itoa(argumentIndex))
					applicableValue = arg
					extraParameters = append(extraParameters, applicableKey)
				}
			} else if argAsField, isParameter := arg.(Field); isParameter { // Field
				if asParameter, isValidParameter := template.parameters[argAsField.Key]; isValidParameter {
					applicableKey = asParameter
				} else if argAsField.Key != "" {
					applicableKey = paramDef(argAsField.Key)
					extraParameters = append(extraParameters, applicableKey)
				} else {
					applicableKey = paramDef("Parameter" + strconv.Itoa(argumentIndex))
					extraParameters = append(extraParameters, applicableKey)
				}
				applicableValue = argAsField.Value
			} else {
				applicableKey = paramDef("Parameter" + strconv.Itoa(argumentIndex))
				extraParameters = append(extraParameters, applicableKey)
				applicableValue = arg
			}
			parameterValues[applicableKey] = applicableValue
		}
		for _, phrase := range template.phrases {
			var applicableValue interface{} = phrase
			if asParameter, isParameter := phrase.(Parameter); isParameter {
				if mappedValue, hasMappedValue := parameterValues[asParameter]; hasMappedValue {
					applicableValue = asParameter.Process(mappedValue)
				}
			}
			returnArguments = append(returnArguments, applicableValue)
		}
		for _, extraParameter := range extraParameters {
			returnArguments = append(returnArguments, extraParameter.String(), "=", extraParameter.Process(parameterValues[extraParameter]))
		}
	} else {
		returnArguments = args
	}
	return returnArguments
}
