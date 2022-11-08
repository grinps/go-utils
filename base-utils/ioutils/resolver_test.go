package ioutils

import (
	"context"
	"github.com/grinps/go-utils/errext"
	"reflect"
	"testing"
)

var nilSourceTypeName SourceTypeName = "nilSourceType"

type notNilSource struct {
	config *nilSourceConfig
}

func (source *notNilSource) Supports(context context.Context, capability SourceCapability) bool {
	return false
}

type nilSource struct {
	config *nilSourceConfig
}

func (source *nilSource) Supports(context context.Context, capability SourceCapability) bool {
	return false
}

type nilSourceConfig struct {
	counter int
}

func WithCounterIncremented() SourceConfigOpts[*nilSource, *nilSourceConfig] {
	return func(config *nilSourceConfig) {
		config.counter += 1
	}
}

func (config *nilSourceConfig) Supports(context context.Context, source Source) bool {
	if _, isNilSource := source.(*nilSource); isNilSource {
		return true
	}
	return false
}

type nilSourceType struct {
}

func (sourceType *nilSourceType) String() string { return "nilSourceType" }

type NilSourceContextValue string

var nilSourceErrorContext NilSourceContextValue = "NilSourceErrorContext"
var nilContextError = errext.NewErrorCode(0)

func (sourceType *nilSourceType) NewSource(context context.Context, config SourceConfig) (Source, error) {
	if context != nil && context.Value(nilSourceErrorContext) == "SrcErr" {
		return nil, nilContextError.NewF("Raising a new Source error")
	}
	if context != nil && context.Value(nilSourceErrorContext) == "SrcWrong" {
		return &notNilSource{}, nil
	}
	if context != nil && context.Value(nilSourceErrorContext) == "SrcNil" {
		return nil, nil
	}
	if asNilSourceConfig, isNilSourceConfig := config.(*nilSourceConfig); isNilSourceConfig {
		return &nilSource{config: asNilSourceConfig}, nil
	}
	return nil, ResolverInvalidValue.NewF("Invalid source config type. Expected *nilSourceConfig, Actual", reflect.TypeOf(config))
}

func (sourceType *nilSourceType) NewConfig(context context.Context) (SourceConfig, error) {
	if context != nil && context.Value(nilSourceErrorContext) == "CfgErr" {
		return nil, nilContextError.NewF("Raising a new config error")
	}
	return &nilSourceConfig{}, nil
}

func TestNewResolverSystem(t *testing.T) {
	resolverSystem := NewResolverSystem(nil)
	if resolverSystem == nil {
		t.Error("Expected not nil Actual nil")
	}
}

var nilResolverSystem *ResolverSystem = nil

func TestResolverSystem_Register(t *testing.T) {
	t.Run("NilResolverSystemValidValues", func(t *testing.T) {
		oldResolver, err := nilResolverSystem.Register(nil, nilSourceTypeName, &nilSourceType{})
		if err == nil {
			t.Error("Expected error, actual no error")
		}
		if oldResolver != nil {
			t.Error("Expected nil resolver, Actual ", oldResolver)
		}
		if _, isErr := ResolverNotInitialized.AsError(err); !isErr {
			t.Errorf("Expected error of type ResolverNotInitialized. Actual %#v", err)
		}
		retrievedResolver, retrievalErr := nilResolverSystem.getResolver(nil, nilSourceTypeName)
		if retrievalErr == nil {
			t.Error("Expected error, Actual no error")
		}
		if _, isErr := ResolverNotInitialized.AsError(retrievalErr); !isErr {
			t.Errorf("Expected error of type ResolverNotInitialized. Actual %#v", err)
		}
		if retrievedResolver != nil {
			t.Error("Expected nil source type to be retrieved, actual ", retrievedResolver)
		}
	})
	t.Run("NilResolverSystemInValidValues", func(t *testing.T) {
		oldResolver, err := nilResolverSystem.Register(nil, "", nil)
		if err == nil {
			t.Error("Expected error, actual no error")
		}
		if oldResolver != nil {
			t.Error("Expected nil, Actual ", oldResolver)
		}
		if _, isErr := ResolverNotInitialized.AsError(err); !isErr {
			t.Errorf("Expected error of type ResolverNotInitialized. Actual %#v", err)
		}
	})
	var resolverSystem = NewResolverSystem(nil)
	t.Run("RegisterEmptySourceTypeNameWithValidSourceType", func(t *testing.T) {
		resolver, err := resolverSystem.Register(nil, "", &nilSourceType{})
		if err == nil {
			t.Error("Expected error, actual no error")
		}
		if _, isErr := ResolverInvalidValue.AsError(err); !isErr {
			t.Errorf("Expected error of type ResolverInvalidValue. Actual %#v", err)
		}
		if resolver != nil {
			t.Error("Expected resolver returned to be nil")
		}
	})
	t.Run("RegisterEmptySourceTypeNameWithNilSourceType", func(t *testing.T) {
		resolver, err := resolverSystem.Register(nil, "", nil)
		if err == nil {
			t.Error("Expected error, actual no error")
		}
		if _, isErr := ResolverInvalidValue.AsError(err); !isErr {
			t.Errorf("Expected error of type ResolverInvalidValue. Actual %#v", err)
		}
		if resolver != nil {
			t.Error("Expected resolver returned to be nil")
		}
	})
	t.Run("RegisterValidSourceTypeNameWithNilSourceType", func(t *testing.T) {
		resolver, err := resolverSystem.Register(nil, nilSourceTypeName, nil)
		if err == nil {
			t.Error("Expected error, actual no error")
		}
		if _, isErr := ResolverInvalidValue.AsError(err); !isErr {
			t.Errorf("Expected error of type ResolverNotInitialized. Actual %#v", err)
		}
		if resolver != nil {
			t.Error("Expected resolver returned to be nil")
		}
		retrievedResolver, retrievalErr := resolverSystem.getResolver(nil, nilSourceTypeName)
		if retrievalErr == nil {
			t.Error("Expected error, Actual no error")
		}
		if _, isErr := ResolverMissingResolver.AsError(retrievalErr); !isErr {
			t.Error("Expected error of type ResolverMissingResolver. Actual ", err)
		}
		if retrievedResolver != nil {
			t.Error("Expected nil source type to be retrieved, actual ", retrievedResolver)
		}
	})
	var initialNilSourceType = &nilSourceType{}
	t.Run("RegisterValidSourceTypeNameWithValueSourceType", func(t *testing.T) {
		resolver, err := resolverSystem.Register(nil, nilSourceTypeName, initialNilSourceType)
		if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
		if resolver != nil {
			t.Error("Expected resolver returned to be nil")
		}
		retrievedResolver, retrievalErr := resolverSystem.getResolver(nil, nilSourceTypeName)
		if retrievalErr != nil {
			t.Errorf("Expected no error, Actual %#v", retrievalErr)
		}
		if retrievedResolver != initialNilSourceType {
			t.Error("Expected source type to be retrieved as ", initialNilSourceType, ", actual ", retrievedResolver)
		}
	})
	t.Run("ReRegisterValidSourceTypeNameWithSameValueSourceType", func(t *testing.T) {
		resolver, err := resolverSystem.Register(nil, nilSourceTypeName, initialNilSourceType)
		if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
		if resolver != initialNilSourceType {
			t.Error("Expected resolver returned to be same as original source type")
		}
		retrievedResolver, retrievalErr := resolverSystem.getResolver(nil, nilSourceTypeName)
		if retrievalErr != nil {
			t.Errorf("Expected no error, Actual %#v", retrievalErr)
		}
		if retrievedResolver != initialNilSourceType {
			t.Error("Expected source type to be retrieved as ", initialNilSourceType, ", actual ", retrievedResolver)
		}
	})
	var newSourceType = &nilSourceType{}
	t.Run("ReRegisterValidSourceTypeNameWithNewValueSourceType", func(t *testing.T) {
		resolver, err := resolverSystem.Register(nil, nilSourceTypeName, newSourceType)
		if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
		if resolver != initialNilSourceType {
			t.Error("Expected resolver returned to be same as original source type since this is first registration of new source type")
		}
		retrievedResolver, retrievalErr := resolverSystem.getResolver(nil, nilSourceTypeName)
		if retrievalErr != nil {
			t.Errorf("Expected no error, Actual %#v", retrievalErr)
		}
		if retrievedResolver != newSourceType {
			t.Error("Expected source type to be retrieved as ", newSourceType, ", actual ", retrievedResolver)
		}
	})
	t.Run("ReRegisterValidSourceTypeNameWithAnotherNewValueSourceType", func(t *testing.T) {
		resolver, err := resolverSystem.Register(nil, nilSourceTypeName, &nilSourceType{})
		if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
		if resolver != newSourceType {
			t.Error("Expected resolver returned to be same as new source type since this follows registration of a new source type")
		}
	})
}

func TestResolverSystem_NewSourceConfig(t *testing.T) {
	t.Run("NilResolverSystemValidValues", func(t *testing.T) {
		config, err := nilResolverSystem.NewSourceConfig(nil, nilSourceTypeName)
		if err == nil {
			t.Error("Expected error, actual no error")
		}
		if _, isErr := ResolverNotInitialized.AsError(err); !isErr {
			t.Errorf("Expected ResolverNotInitialized, actual %#v", err)
		}
		if config != nil {
			t.Error("Expected nil config, actual ", config)
		}
	})
	t.Run("NilResolverSystemInValidValues", func(t *testing.T) {
		config, err := nilResolverSystem.NewSourceConfig(nil, "")
		if err == nil {
			t.Error("Expected error, actual no error")
		}
		if _, isErr := ResolverNotInitialized.AsError(err); !isErr {
			t.Errorf("Expected ResolverNotInitialized, actual %#v", err)
		}
		if config != nil {
			t.Error("Expected nil config, actual ", config)
		}
	})
	var resolverSystem = NewResolverSystem(nil)
	t.Run("ResolverSystemInValidName", func(t *testing.T) {
		config, err := resolverSystem.NewSourceConfig(nil, "")
		if err == nil {
			t.Error("Expected error, actual no error")
		}
		if _, isErr := ResolverInvalidValue.AsError(err); !isErr {
			t.Errorf("Expected ResolverInvalidValue, actual %#v", err)
		}
		if config != nil {
			t.Error("Expected nil config, actual ", config)
		}
	})
	t.Run("ResolverSystemValidNameMissingRegistration", func(t *testing.T) {
		config, err := resolverSystem.NewSourceConfig(nil, nilSourceTypeName)
		if err == nil {
			t.Error("Expected error, actual no error")
		}
		if _, isErr := ResolverMissingResolver.AsError(err); !isErr {
			t.Errorf("Expected ResolverMissingResolver, actual %#v", err)
		}
		if config != nil {
			t.Error("Expected nil config, actual ", config)
		}
	})
	t.Run("ResolverSystemValidNameWithRegistration", func(t *testing.T) {
		resolverSystem.Register(nil, nilSourceTypeName, &nilSourceType{})
		config, err := resolverSystem.NewSourceConfig(nil, nilSourceTypeName)
		if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
		if config == nil {
			t.Error("Expected not nil config, actual nil")
		}
		if nilCfg, isNilCfg := config.(*nilSourceConfig); !isNilCfg {
			t.Errorf("Expected config type nilSourceConfig, actual %#v", config)
		} else if !nilCfg.Supports(nil, &nilSource{}) {
			t.Error("Expected nilCfg config to support nilSource, Actual false")
		}
	})
}

func TestRegister(t *testing.T) {
	t.Run("RegisterInvalidNameAndNilValue", func(t *testing.T) {
		previousResolver := Register(nil, "", nil)
		if previousResolver != nil {
			t.Errorf("Expected nil, Actual %#v", previousResolver)
		}
	})
	t.Run("ReRegisterInvalidNameAndNilValue", func(t *testing.T) {
		previousResolver := Register(nil, nilSourceTypeName, nil)
		if previousResolver != nil {
			t.Errorf("Expected nil, Actual %#v", previousResolver)
		}
	})
	t.Run("RegisterValidNameAndInValidValue", func(t *testing.T) {
		previousResolver := Register(nil, nilSourceTypeName, nil)
		if previousResolver != nil {
			t.Errorf("Expected nil, Actual %#v", previousResolver)
		}
		source, err := ResolveE[*nilSource, *nilSourceConfig](nil, nilSourceTypeName)
		if err == nil {
			t.Error("Expected error, Actual no error")
		}
		if _, isErr := ResolverSourceResolutionFailed.AsError(err); !isErr {
			t.Errorf("Expected ResolverSourceResolutionFailed, Actual %#v", err)
		}
		if source != nil {
			t.Errorf("Expected nil Actual %#v", source)
		}
	})
	t.Run("RegisterValidNameAndValidValue", func(t *testing.T) {
		previousResolver := Register(nil, nilSourceTypeName, &nilSourceType{})
		if previousResolver != nil {
			t.Errorf("Expected nil, Actual %#v", previousResolver)
		}
		source, err := ResolveE[*nilSource, *nilSourceConfig](nil, nilSourceTypeName, WithCounterIncremented(), WithCounterIncremented())
		if err != nil {
			t.Errorf("Expected no error, Actual %#v", err)
		}
		if source == nil {
			t.Error("Expected not nil Actual nil")
		}
		if source.config.counter != 2 {
			t.Error("Expected counter to be 2 since 2 WithCounterIncremented have been passed, Actual", source.config.counter)
		}
	})
}

func TestRegisterP(t *testing.T) {
	t.Run("RegisterPWithInvalidValues", func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil {
				if asError, isError := err.(error); !isError {
					t.Errorf("Expected error in panic, Actual %#v", err)
				} else if _, isResolverErr := ResolverInvalidValue.AsError(asError); !isResolverErr {
					t.Errorf("Expected ResolverInvalidValue, Actual %#v", err)
				}
			} else {
				t.Error("Expected value in panic, Actual no value but panic")
			}
		}()
		RegisterP(nil, "", nil)
		t.Error("Expected panic due to error, Actual no panic")
	})
	t.Run("RegisterPWithValidValues", func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("No value was expected in recovery since this call should be successful, actual %#v", err)
			}
		}()
		RegisterP(nil, nilSourceTypeName, &nilSourceType{})
	})
}

func TestResolve(t *testing.T) {
	var testResolveNilSourceType SourceTypeName = "testResolveNilSourceType"
	Register(nil, testResolveNilSourceType, &nilSourceType{})
	source := Resolve[*nilSource, *nilSourceConfig](nil, testResolveNilSourceType, WithCounterIncremented(), WithCounterIncremented())
	if source == nil {
		t.Error("Expected not nil Actual nil")
	}
	if source.config.counter != 2 {
		t.Error("Expected counter to be 2 since 2 WithCounterIncremented have been passed, Actual", source.config.counter)
	}
}

func TestResolveP(t *testing.T) {
	t.Run("ResolvePWithInvalidValue", func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil {
				if asError, isError := err.(error); !isError {
					t.Errorf("Expected error in panic, Actual %#v", err)
				} else if _, isResolverErr := ResolverSourceResolutionFailed.AsError(asError); !isResolverErr {
					t.Errorf("Expected ResolverSourceResolutionFailed, Actual %#v", err)
				}
			} else {
				t.Error("Expected value in panic, Actual no value but panic")
			}
		}()
		var testResolvePNilSourceType SourceTypeName = "testResolvePNilSourceType"
		Register(nil, testResolvePNilSourceType, &nilSourceType{})
		var errContext = context.WithValue(context.TODO(), nilSourceErrorContext, "CfgErr")
		source := ResolveP[*nilSource, *nilSourceConfig](errContext, testResolvePNilSourceType, WithCounterIncremented(), WithCounterIncremented())
		if source != nil {
			t.Errorf("Expected nil Actual %#v", source)
		}
	})
	t.Run("ResolvePWithValidValue", func(t *testing.T) {
		defer func() {
			if err := recover(); err != nil {
				t.Errorf("Expected no value in panic, Actual %#v", err)
			}
		}()
		var testResolvePNilSourceType SourceTypeName = "testResolvePNilSourceTypeValidValue"
		Register(nil, testResolvePNilSourceType, &nilSourceType{})
		source := ResolveP[*nilSource, *nilSourceConfig](nil, testResolvePNilSourceType, WithCounterIncremented(), WithCounterIncremented())
		if source == nil {
			t.Error("Expected not nil Actual nil")
		}
		if source.config.counter != 2 {
			t.Error("Expected counter to be 2 since 2 WithCounterIncremented have been passed, Actual", source.config.counter)
		}
	})
}

func TestResolveEWithResolverSystem(t *testing.T) {
	t.Run("NilSystem", func(t *testing.T) {
		source, err := ResolveEWithResolverSystem[*nilSource, *nilSourceConfig](nil, nil, nilSourceTypeName)
		if source != nil {
			t.Errorf("Expected nil actual %#v", source)
		}
		if err == nil {
			t.Error("Expected not nil error, actual nil error")
		} else if _, isErr := ResolverInvalidValue.AsError(err); !isErr {
			t.Errorf("Expected ResolverInvalidValue, Actual %#v", err)
		}
	})
	var resolverSystem = NewResolverSystem(nil)
	resolverSystem.Register(nil, nilSourceTypeName, &nilSourceType{})
	t.Run("ValidSystemWithErrorGeneratingSource", func(t *testing.T) {
		var errContext = context.WithValue(context.TODO(), nilSourceErrorContext, "SrcErr")
		source, err := ResolveEWithResolverSystem[*nilSource, *nilSourceConfig](errContext, resolverSystem, nilSourceTypeName)
		if source != nil {
			t.Errorf("Expected nil actual %#v", source)
		}
		if err == nil {
			t.Error("Expected not nil error, actual nil error")
		} else if _, isErr := ResolverSourceResolutionFailed.AsError(err); !isErr {
			t.Errorf("Expected ResolverSourceResolutionFailed, Actual %#v", err)
		}
	})
	t.Run("ValidSystemWithWrongSource", func(t *testing.T) {
		var errContext = context.WithValue(context.TODO(), nilSourceErrorContext, "SrcWrong")
		source, err := ResolveEWithResolverSystem[*nilSource, *nilSourceConfig](errContext, resolverSystem, nilSourceTypeName)
		if source != nil {
			t.Errorf("Expected nil actual %#v", source)
		}
		if err == nil {
			t.Error("Expected not nil error, actual nil error")
		} else if _, isErr := ResolverSourceResolutionFailed.AsError(err); !isErr {
			t.Errorf("Expected ResolverSourceResolutionFailed, Actual %#v", err)
		}
	})
	t.Run("ValidSystemWithNilSource", func(t *testing.T) {
		var errContext = context.WithValue(context.TODO(), nilSourceErrorContext, "SrcNil")
		source, err := ResolveEWithResolverSystem[*nilSource, *nilSourceConfig](errContext, resolverSystem, nilSourceTypeName)
		if source != nil {
			t.Errorf("Expected nil actual %#v", source)
		}
		if err == nil {
			t.Error("Expected not nil error, actual nil error")
		} else if _, isErr := ResolverSourceResolutionFailed.AsError(err); !isErr {
			t.Errorf("Expected ResolverSourceResolutionFailed, Actual %#v", err)
		}
	})
}

func TestDefault(t *testing.T) {
	if Default() != defaultResolverSystem {
		t.Errorf("Expected Default resolve system %#v Actual %#v", defaultResolverSystem, Default())
	}
}
