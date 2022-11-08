package ioutils

import (
	"context"
	logger "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/base-utils/registry"
	"github.com/grinps/go-utils/errext"
	"reflect"
)

const ResolverErrorCodeType = "resolverErrorCodeType"

type ResolverSystem struct {
	registry    *registry.Register[SourceTypeName, SourceType]
	initialized bool
}

var ResolverNotInitialized = errext.NewErrorCodeOfType(1, ResolverErrorCodeType)
var ResolverInvalidValue = errext.NewErrorCodeOfType(2, ResolverErrorCodeType)
var ResolverMissingResolver = errext.NewErrorCodeOfType(3, ResolverErrorCodeType)
var ResolverSourceConfigCreationFailed = errext.NewErrorCodeOfType(4, ResolverErrorCodeType)
var ResolverSourceResolutionFailed = errext.NewErrorCodeOfType(5, ResolverErrorCodeType)

func (system *ResolverSystem) Register(context context.Context, sourceTypeName SourceTypeName, sourceType SourceType) (SourceType, error) {
	logger.Log("Registering source type ", sourceTypeName, " as ", sourceType)
	var returnErrorCode error = nil
	var returnSourceType SourceType = nil
	if system == nil || !system.initialized || system.registry == nil {
		returnErrorCode = ResolverNotInitialized.New("Resolver is not initialized properly.")
	} else if sourceTypeName == "" {
		returnErrorCode = ResolverInvalidValue.New("Source type name is empty string")
	} else if sourceType == nil {
		returnErrorCode = ResolverInvalidValue.New("Nil source type implementation was passed.")
	} else {
		logger.Log("Registering source type ", sourceTypeName)
		//TODO Add new API that allows registration only if not already registered.
		returnSourceType = system.registry.Register(sourceTypeName, sourceType)
		logger.Log("Current source type", returnSourceType)
	}
	logger.Log("Return values ", "value", returnSourceType, "error", returnErrorCode)
	return returnSourceType, returnErrorCode
}

func (system *ResolverSystem) NewSourceConfig(context context.Context, sourceType SourceTypeName) (SourceConfig, error) {
	logger.Log("New source config for type ", sourceType)
	var returnConfig SourceConfig
	var returnError error = nil
	var resolver SourceType = nil
	resolver, returnError = system.getResolver(context, sourceType)
	if resolver != nil {
		config, configErr := resolver.NewConfig(context)
		if configErr != nil {
			returnError = ResolverSourceConfigCreationFailed.NewWithErrorF(configErr, "Failed to create config for source type ", sourceType, " using impl ", resolver)
		} else {
			returnConfig = config
		}
	}
	logger.Log("Return values ", "value", returnConfig, "error", returnError)
	return returnConfig, returnError
}

func (system *ResolverSystem) Resolve(context context.Context, sourceType SourceTypeName, config SourceConfig) (Source, error) {
	logger.Log("Resolving source of type ", sourceType, " with config ", config)
	var returnValue Source
	var returnError error = nil
	var resolver SourceType = nil
	resolver, returnError = system.getResolver(context, sourceType)
	if resolver != nil {
		sourceCreated, sourceCreationFailure := resolver.NewSource(context, config)
		logger.Log("New source created as ", sourceCreated)
		if sourceCreationFailure != nil {
			returnError = ResolverSourceResolutionFailed.NewWithErrorF(sourceCreationFailure, "Failed to create source for ", sourceType, " with config ", config)
		} else {
			returnValue = sourceCreated
		}
	}
	logger.Log("Return values ", "value", returnValue, "error", returnError)
	return returnValue, returnError
}

func (system *ResolverSystem) getResolver(context context.Context, sourceType SourceTypeName) (SourceType, error) {
	var returnSourceType SourceType = nil
	var returnError error = nil
	if system == nil || !system.initialized || system.registry == nil {
		returnError = ResolverNotInitialized.New("Resolver is not initialized properly.")
	} else if sourceType == "" {
		returnError = ResolverInvalidValue.New("Source type name is empty string")
	} else {
		resolver := system.registry.Get(sourceType)
		logger.Log("Resolver found ", resolver)
		if resolver == nil {
			returnError = ResolverMissingResolver.NewF("No associated resolver for ", sourceType, " could be located")
		} else {
			returnSourceType = resolver
		}
	}
	return returnSourceType, returnError
}

var defaultResolverSystem = NewResolverSystem(context.TODO())

func Default() *ResolverSystem {
	return defaultResolverSystem
}
func NewResolverSystem(context context.Context) *ResolverSystem {
	logger.Log("Creating new Resolver System")
	return &ResolverSystem{
		registry:    registry.NewRegister[SourceTypeName, SourceType](),
		initialized: true,
	}
}

func Register(context context.Context, sourceTypeName SourceTypeName, sourceType SourceType) SourceType {
	result, _ := RegisterE(context, sourceTypeName, sourceType)
	return result
}

func RegisterP(context context.Context, sourceTypeName SourceTypeName, sourceType SourceType) SourceType {
	result, err := RegisterE(context, sourceTypeName, sourceType)
	if err != nil {
		panic(err)
	}
	return result
}

func RegisterE(context context.Context, sourceTypeName SourceTypeName, sourceType SourceType) (SourceType, error) {
	return defaultResolverSystem.Register(context, sourceTypeName, sourceType)
}

func Resolve[T Source, V SourceConfig](context context.Context, sourceTypeName SourceTypeName, opts ...SourceConfigOpts[T, V]) T {
	result, _ := ResolveE(context, sourceTypeName, opts...)
	return result
}

func ResolveP[T Source, V SourceConfig](context context.Context, sourceTypeName SourceTypeName, opts ...SourceConfigOpts[T, V]) T {
	result, err := ResolveE(context, sourceTypeName, opts...)
	if err != nil {
		panic(err)
	}
	return result
}

func ResolveE[T Source, V SourceConfig](context context.Context, sourceTypeName SourceTypeName, opts ...SourceConfigOpts[T, V]) (T, error) {
	return ResolveEWithResolverSystem(context, defaultResolverSystem, sourceTypeName, opts...)
}

func ResolveEWithResolverSystem[T Source, V SourceConfig](context context.Context, resolverSystem *ResolverSystem, sourceTypeName SourceTypeName, opts ...SourceConfigOpts[T, V]) (T, error) {
	var returnError error = nil
	var returnSource T
	var sourceConfigType V
	if resolverSystem == nil {
		return returnSource, ResolverInvalidValue.New("No resolver system was passed.")
	}
	sourceConfig, configErr := resolverSystem.NewSourceConfig(context, sourceTypeName)
	if configErr == nil {
		if sourceConfig != nil {
			if sourceConfigAsV, isV := sourceConfig.(V); isV {
				logger.Log("Updating the config using passed options", "opts", opts)
				for _, option := range opts {
					option(sourceConfigAsV)
				}
			} else {
				logger.Log("Skipping adding the passed options since the source config is not of expected type.Expected type ", reflect.TypeOf(sourceConfigType), "actual", reflect.TypeOf(sourceConfig))
			}
		}
		source, sourceResolveErr := resolverSystem.Resolve(context, sourceTypeName, sourceConfig)
		if sourceResolveErr != nil {
			returnError = sourceResolveErr
		} else if sourceAsT, isTypeT := source.(T); isTypeT {
			returnSource = sourceAsT
		} else {
			returnError = ResolverSourceResolutionFailed.NewF("Failed to resolve source of type ", sourceTypeName, " due to mismatched source. Expected type ", reflect.TypeOf(returnSource), "actual", reflect.TypeOf(source))
		}
	} else {
		returnError = ResolverSourceResolutionFailed.NewWithErrorF(configErr, "Failed to resolve for source name ", sourceTypeName, " due to config creation error.")
	}
	return returnSource, returnError
}
