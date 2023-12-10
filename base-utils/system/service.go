package system

import (
	"fmt"
	"github.com/grinps/go-utils/errext"
	"reflect"
	"sync"
)

const InvalidServiceName = "InvalidServiceName"

type ServiceType[T any] interface {
	fmt.Stringer
	As(input any) (T, error)
}

type Service[T any] interface {
	fmt.Stringer
	AsService() T
}

type Comparable[V comparable] interface {
	Key() V
}

type namedService[T any] struct {
	name    string
	service T
}

func (name *namedService[T]) String() string {
	if name == nil || name.name == "" {
		return InvalidServiceName
	}
	return name.name
}

func (name *namedService[T]) As(input any) (outVal T, err error) {
	outVal, err = AsE[any, T](input)
	return
}

func (name *namedService[T]) AsService() T {
	var nilValue T
	if name == nil {
		return nilValue
	}
	return name.service
}

func (name *namedService[T]) Key() string {
	return name.String()
}

func NewNamedService[T any](name string, service T) Service[T] {
	newServiceName, _ := NewNamedServiceE[T](name, service)
	return newServiceName
}

func NewNamedServiceP[T any](name string, service T) Service[T] {
	newServiceName, err := NewNamedServiceE[T](name, service)
	if err != nil {
		panic(err)
	}
	return newServiceName
}

func NewNamedServiceE[T any](name string, service T) (Service[T], error) {
	if name == "" {
		return nil, ErrServiceCreationFailed.NewF(errext.NewField(ErrSystemParamServiceName, name),
			errext.NewField(ErrSystemParamServiceInstance, service),
			errext.NewField(ErrSystemParamReason, ErrSystemReasonServiceNameEmpty))
	}
	if reflect.ValueOf(service).IsNil() { //TODO: Check kind to avoid panic
		return nil, ErrServiceCreationFailed.NewF(errext.NewField(ErrSystemParamServiceName, name),
			errext.NewField(ErrSystemParamServiceInstance, service),
			errext.NewField(ErrSystemParamReason, ErrSystemReasonServiceInstanceNil))
	}
	return &namedService[T]{
		name:    name,
		service: service,
	}, nil
}

func NewServiceType[T any](name string) ServiceType[T] {
	newServiceName, _ := NewServiceTypeE[T](name)
	return newServiceName
}

func NewServiceTypeP[T any](name string) ServiceType[T] {
	newServiceType, err := NewServiceTypeE[T](name)
	if err != nil {
		panic(err)
	}
	return newServiceType
}

var serviceTypeMutex = &sync.Mutex{}
var serviceTypeSet = map[string]struct{}{}

func NewServiceTypeE[T any](name string) (ServiceType[T], error) {
	if name == "" {
		return nil, ErrServiceTypeCreationFailed.NewF(errext.NewField(ErrSystemParamServiceType, name), errext.NewField(ErrSystemParamReason, ErrSystemReasonServiceTypeEmpty))
	}
	serviceTypeMutex.Lock()
	defer serviceTypeMutex.Unlock()
	if _, hasValue := serviceTypeSet[name]; hasValue {
		return nil, ErrServiceTypeCreationFailed.NewF(errext.NewField(ErrSystemParamServiceType, name),
			errext.NewField(ErrSystemParamReason, ErrSystemReasonServiceTypeAlreadyRegistered))
	} else {
		serviceTypeSet[name] = struct{}{}
	}
	return &namedService[T]{name: name}, nil
}
