package config

import (
	"context"
	"reflect"
)

const DefaultKeyDelimiter = "."
const EmptyKeyName = "<empty key>"
const InvalidKeyName = "<invalid key %#v>"

type DriverOption interface {
	AsOption() DriverOption
}

type GetOption interface {
	AsGetOption() GetOption
}

type Driver[C Config] interface {
	// Name Must be able to work with nil instance of driver object.
	Name() string
	NewConfigMgr(ctx context.Context, opts ...DriverOption) (C, error)
}

type Key interface {
	Key() Key
}

var simpleKeyType = reflect.TypeOf((*SimpleKey)(nil)).Elem()

type SimpleKey interface {
	Key
	Name(ctx context.Context) string
}

var hierarchyKeyType = reflect.TypeOf((*HierarchyKey)(nil)).Elem()

type HierarchyKey interface {
	Key
	HasChild(ctx context.Context) bool
	Child(ctx context.Context) Key
}

type Resolver[I any, O any, R any] func(ctx context.Context, resolvers []R, input I, previousOutput O) (O, error)

type InitOption[C Config] func(ctx context.Context, mgr C) error

func (opt InitOption[C]) AsOption() DriverOption {
	return opt
}

var keyParserType = reflect.TypeOf((*KeyParser)(nil)).Elem()

type KeyParser Resolver[string, []Key, KeyParser]

func (opt KeyParser) AsOption() DriverOption { return opt }

func (opt KeyParser) AsGetOption() GetOption { return opt }

type KeyResolutionDetail [2]string

func (keyDet KeyResolutionDetail) ResolverName() string { return keyDet[0] }

func (keyDet KeyResolutionDetail) ResolveKey() Key { return simpleStringKey(keyDet[1]) }

type KeyNameResolver Resolver[KeyResolutionDetail, []Key, KeyNameResolver]

type ValueRetriever[C Config] func(ctx context.Context, config C, originalKey string, key []Key, previousValue any) (value any, getErr error)

func (opt ValueRetriever[C]) AsOption() DriverOption { return opt }

func (opt ValueRetriever[C]) AsGetOption() GetOption {
	return opt
}

type Config interface {
	GetValue(ctx context.Context, key string, options ...GetOption) (any, error)
	GetConfig(ctx context.Context, key string, options ...GetOption) (Config, error)
}

type CommonRegistries interface {
	RegisterDriver() CommonRegistries
	RegisterDriverInstance(name string, driver any) CommonRegistries
}
