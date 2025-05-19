package config

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

func GetAsMap(ctx context.Context, input any, extErr ExtensibleErrFunc) (map[string]any, error) {
	if extErr == nil {
		extErr = ExtensibleErr(ctx, nil, "")
	}
	applicableConfigMap := map[string]any{}
	if input == nil {
		return map[string]any(nil), nil
	}
	switch input.(type) {
	case string:
		//TODO: move this to cast code.
		unmarshalErr := json.Unmarshal([]byte(input.(string)), &applicableConfigMap)
		if unmarshalErr != nil {
			return nil, extErr(WithErrorCode(ErrConfigInvalidValue), WithCause(fmt.Errorf("failed to unmarshal string %s into map due to error %w", input, unmarshalErr)))
		}
	case map[string]string:
		for k, v := range input.(map[string]string) {
			applicableConfigMap[k] = v
		}
	case map[any]any:
		var transformationErrors []error
		for k, v := range input.(map[any]any) {
			addErr := addToMapOfStringKeys(&applicableConfigMap, k, v)
			if addErr != nil {
				transformationErrors = append(transformationErrors, addErr)
			}
		}
		if len(transformationErrors) > 0 {
			return nil, extErr(WithErrorCode(ErrConfigInvalidValue), WithCause(fmt.Errorf("failed to transform %v (%T) of type map[any][any] to map[string]any due to errors %w", input, input, errors.Join(transformationErrors...))))
		}
	case map[string]any:
		applicableConfigMap = input.(map[string]any)
	default:
		return nil, extErr(WithErrorCode(ErrConfigInvalidValue), WithCause(fmt.Errorf("conversion of configuration %v of type %T to map[string]any not supported", input, input)))
	}
	return applicableConfigMap, nil
}

func addToMapOfStringKeys(ptrMap *map[string]any, k any, v any) error {
	if k == nil {
		return fmt.Errorf("failed to add key %#v with value %v to map since key is nil", k, v)
	}
	kAsString, transformErr := toString(k)
	if transformErr != nil {
		return transformErr
	}
	(*ptrMap)[kAsString] = v
	return nil
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()
var fmtStringerType = reflect.TypeOf((*fmt.Stringer)(nil)).Elem()

// Reference: https://github.com/spf13/cast/blob/master/caste.go#L915
// TODO: Checkout https://github.com/spf13/cast and create this as a module in base
func toString(i any) (string, error) {
	if i == nil {
		return "", fmt.Errorf("unable to cast nil value to string")
	}
	// resolve pointer to element except if it is not fmt.Stringer & error interfaces
	v := reflect.ValueOf(i)
	for !v.Type().Implements(fmtStringerType) && !v.Type().Implements(errorType) && v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	i = v.Interface()
	switch s := i.(type) {
	case string:
		return s, nil
	/*case bool:
		return strconv.FormatBool(s), nil
	case float64:
		return strconv.FormatFloat(s, 'f', -1, 64), nil
	case float32:
		return strconv.FormatFloat(float64(s), 'f', -1, 32), nil
	case int:
		return strconv.Itoa(s), nil
	case int64:
		return strconv.FormatInt(s, 10), nil
	case int32:
		return strconv.Itoa(int(s)), nil
	case int16:
		return strconv.FormatInt(int64(s), 10), nil
	case int8:
		return strconv.FormatInt(int64(s), 10), nil
	case uint:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint64:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(s), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(s), 10), nil
	case json.Number:
		return s.String(), nil
	case []byte:
		return string(s), nil
	case template.HTML:
		return string(s), nil
	case template.URL:
		return string(s), nil
	case template.JS:
		return string(s), nil
	case template.CSS:
		return string(s), nil
	case template.HTMLAttr:
		return string(s), nil
	case nil:
		return "", nil
	case fmt.Stringer:
		return s.String(), nil
	case error:
		return s.Error(), nil
	*/
	default:
		return "", fmt.Errorf("unable to cast %#v of type %T to string", i, i)
	}
}
