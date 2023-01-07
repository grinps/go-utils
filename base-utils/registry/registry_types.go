package registry

import (
	"github.com/grinps/go-utils/errext"
	"reflect"
	"sync"
)

type registrationRecord[Value any] struct {
	valueSet bool
	value    Value
	lock     *sync.RWMutex
}

type Register[Key any, Value any] interface {
	Get(key Key) Value
	Register(key Key, newValue Value) Value
	Unregister(key Key) Value
}

type registryComparable[T comparable, Key any, Value any] struct {
	register map[T]*registrationRecord[Value]
	lock     *sync.RWMutex
}

type CustomKey[T comparable] interface {
	Unique() T
}

type CustomKeyAsFunction[T comparable] func() T

func (f CustomKeyAsFunction[T]) Unique() T {
	return f()
}

var ErrComparableValueConvertFailed = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to convert key of type", "[type]", "due to", "reason"))

const (
	ErrComparableValueConvertFailedParamType          = "type"
	ErrComparableValueConvertFailedParamReason        = "reason"
	ErrComparableValueConvertFailedReasonIncompatible = "type is neither comparable nor CustomKey[comparable]"
)

func ComparableValueConverter[T comparable](a any) (CustomKey[T], error) {
	if asCustomKey, isCustomKey := a.(CustomKey[T]); isCustomKey {
		return asCustomKey, nil
	} else if asT, isT := a.(T); isT {
		return CustomKeyAsFunction[T](func() T {
			return asT
		}), nil
	} else {
		return nil, ErrComparableValueConvertFailed.NewF(ErrComparableValueConvertFailedParamType, reflect.TypeOf(a),
			ErrComparableValueConvertFailedParamReason, ErrComparableValueConvertFailedReasonIncompatible)
	}
}
