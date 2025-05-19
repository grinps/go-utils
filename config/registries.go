package config

import (
	"context"
	"github.com/grinps/go-utils/base-utils/registry"
)

type RegistrationOption interface {
	Register(ctx context.Context, r *registries)
}

type RegistrationOptionF func(ctx context.Context, r *registries)

func (opt RegistrationOptionF) Register(ctx context.Context, r *registries) {
	opt(ctx, r)
}

type CommonRegister interface {
	Register(ctx context.Context, options ...RegistrationOption) CommonRegister
}

type registries struct {
	DriverRegistry registry.Register[string, any]
	SourceRegistry registry.Register[string, any]
}

func (r *registries) Register(ctx context.Context, options ...RegistrationOption) CommonRegister {
	for _, option := range options {
		option.Register(ctx, r)
	}
	return r
}

var commonRegistries = registries{
	DriverRegistry: registry.NewRegister[string, any](),
	SourceRegistry: registry.NewRegister[string, any](),
}

func Register(ctx context.Context, options ...RegistrationOption) CommonRegister {
	return commonRegistries.Register(ctx, options...)
}
