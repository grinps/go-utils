package system

import (
	logger "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
	"reflect"
)

const ErrSystemInputParamType = "inputType"
const ErrSystemOutputParamType = "outputType"
const ErrSystemReasonInputNilValue = "input value is nil"
const ErrSystemReasonOutputNilValue = "output value is nil"
const ErrSystemReasonTargetInvalid = "invalid target (expected not-nil pointer)"
const ErrSystemReasonTargetNotInterface = "invalid target (*target must be interface)"
const ErrSystemReasonTransformationNotAllowed = "target is not supported transformation (not AssignableTo, no As)"

var ErrAsFailedInvalidInput = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to transform input of type",
	"["+ErrSystemInputParamType+"]", "to type", "["+ErrSystemOutputParamType+"]", "due to error", "["+ErrSystemParamReason+"]"))
var ErrAsFailedTransformation = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to transform input of type",
	"["+ErrSystemInputParamType+"]", "to type", "["+ErrSystemOutputParamType+"]", "due to error", "["+ErrSystemParamReason+"]"))

func As[IN any, OUT any](input IN) (output OUT) {
	var _ = AsType(input, &output)
	return

}

func AsP[IN any, OUT any](input IN) (output OUT) {
	outcome := AsType(input, &output)
	if outcome != nil {
		panic(outcome)
	}
	return
}

func AsE[IN any, OUT any](input IN) (output OUT, err error) {
	err = AsType(input, &output)
	return
}

func AsB[IN any, OUT any](input IN) (output OUT, ok bool) {
	outcome := AsType(input, &output)
	if outcome != nil {
		ok = false
	} else {
		ok = true
	}
	return
}

func AsType(input any, output any) (err error) {
	logger.Log("Entering AsType(", input, ",", output, ")")
	defer func() { logger.Log("Exiting  AsType(", input, ",", output, ")", err) }()
	err = nil
	if input == nil {
		err = ErrAsFailedInvalidInput.NewF(errext.NewField(ErrSystemInputParamType, reflect.TypeOf(input)),
			errext.NewField(ErrSystemOutputParamType, reflect.TypeOf(output)),
			errext.NewField(ErrSystemParamReason, ErrSystemReasonInputNilValue))
		return
	}
	if output == nil {
		err = ErrAsFailedInvalidInput.NewF(errext.NewField(ErrSystemInputParamType, reflect.TypeOf(input)),
			errext.NewField(ErrSystemOutputParamType, reflect.TypeOf(output)),
			errext.NewField(ErrSystemParamReason, ErrSystemReasonOutputNilValue))
		return
	}
	val := reflect.ValueOf(output)
	logger.Log("AsType: Value of output", val)
	typ := val.Type()
	logger.Log("AsType: Type of output", typ)
	if typ.Kind() != reflect.Ptr || val.IsNil() {
		err = ErrAsFailedInvalidInput.NewF(errext.NewField(ErrSystemInputParamType, reflect.TypeOf(input)),
			errext.NewField(ErrSystemOutputParamType, reflect.TypeOf(output)),
			errext.NewField(ErrSystemParamReason, ErrSystemReasonTargetInvalid))
		return
	}
	targetType := typ.Elem()
	logger.Log("AsType: Element type of output", targetType)
	if reflect.TypeOf(input).AssignableTo(targetType) {
		logger.Log("AsType: Input is assignable")
		val.Elem().Set(reflect.ValueOf(input))
		return
	}
	if reflect.TypeOf(input).ConvertibleTo(targetType) {
		logger.Log("AsType: Input is convertible")
		val.Elem().Set(reflect.ValueOf(input).Convert(targetType))
	}
	/*
		Need to find use-case
		if targetType.Kind() == reflect.Interface && reflect.TypeOf(input).Implements(targetType) {
			logger.Log("AsType: Input implements target")
			val.Elem().Set(reflect.ValueOf(input))
		}*/
	if x, ok := input.(interface{ As(any) bool }); ok && x.As(output) {
		return
	}
	err = ErrAsFailedTransformation.NewF(errext.NewField(ErrSystemInputParamType, reflect.TypeOf(input)),
		errext.NewField(ErrSystemOutputParamType, reflect.TypeOf(output)),
		errext.NewField(ErrSystemParamReason, ErrSystemReasonTransformationNotAllowed))
	return
}
