package system

import (
	"context"
	"github.com/grinps/go-utils/base-utils/registry"
)

type systemImpl struct {
	registry                   registry.Register[string, any]
	defaultGetOptions          []GetOption
	defaultRegistrationOptions []RegistrationOption
}

func (system *systemImpl) System() System {
	return system
}

func (system *systemImpl) isInitialized() (bool, error) {
	if system == nil {
		return false, ErrSystemNotInitialized.NewF(ErrSystemParamReason, ErrSystemReasonSystemIsNil)
	}
	if system.registry == nil {
		return false, ErrSystemNotInitialized.NewF(ErrSystemParamReason, ErrSystemReasonMissingServiceRegistry)
	}
	return true, nil
}

func NewSystem() System {
	return newSystem()
}

func newSystem() *systemImpl {
	return &systemImpl{
		registry:                   registry.NewRegister[string, any](),
		defaultGetOptions:          defaultGetOptions,
		defaultRegistrationOptions: defaultRegistrationOptions,
	}
}

type Option func(context context.Context, system *systemImpl)

func NewSystemWithOptions(options ...Option) System {
	var basicSystem = newSystem()
	for _, option := range options {
		if option != nil {
			option(context.TODO(), basicSystem)
		}
	}
	return basicSystem
}

func WithGetOptions(options ...GetOption) Option {
	return func(context context.Context, system *systemImpl) {
		if system != nil && len(options) > 0 {
			system.defaultGetOptions = options
		}
	}
}

func WithRegistrationOptions(options ...RegistrationOption) Option {
	return func(context context.Context, system *systemImpl) {
		if system != nil && len(options) > 0 {
			system.defaultRegistrationOptions = options
		}
	}
}
