package system

import (
	"context"
	logger "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
	"reflect"
)

func GetServiceFromSystem[T any](context context.Context, system System, name string, serviceType ServiceType[T], options ...GetOption) (service Service[T], err error) {
	logger.Log("Entering GetServiceFromSystem(", context, ",", system, ",", name, ",", serviceType, ",", options, ")")
	defer func() { logger.Log("Exiting GetServiceFromSystem()", service, ",", err) }()
	if serviceType == nil {
		err = WithGetErrorGenerator(context, system, name, "NilServiceType", options)(ErrSystemReasonServiceTypeNil, nil)
		return
	}
	prefilledErr := WithGetErrorGenerator(context, system, name, serviceType.String(), options)
	applicableServiceTypeName := serviceType.String()
	if asComparable, isComparable := serviceType.(Comparable[string]); isComparable {
		applicableServiceTypeName = asComparable.Key()
	}
	applicableGetOptions := append(GetSystemGetOptions(system), options...)
	if asRegistrySystem, isRegistrySystem := system.(RegistrySystem); isRegistrySystem {
		returnedService, returnedErr := asRegistrySystem.GetService(context, name, applicableServiceTypeName, applicableGetOptions...)
		if returnedErr != nil {
			err = returnedErr
			return
		}
		if returnedService != nil {
			transformedService, transformErr := serviceType.As(returnedService)
			if transformErr != nil {
				err = prefilledErr(ErrSystemReasonServiceTypeTransformationFailed, transformErr)
			} else {
				return NewNamedService(name, transformedService), nil
			}
		}
	} else {
		err = prefilledErr(ErrSystemReasonSystemTypeNotSupported, nil,
			errext.NewField("SystemType", reflect.TypeOf(system)))
	}
	return
}

func (system *systemImpl) GetService(context context.Context, name string, serviceTypeName string, getOptions ...GetOption) (serviceInstance any, err error) {
	logger.Log("Entering GetServiceWithOptions(", context, ",", system, ",", name, ",", serviceTypeName, ",", getOptions, ")")
	defer func() { logger.Log("Exiting GetServiceWithOptions()", serviceInstance, ",", err) }()
	prefilledErr := WithGetErrorGenerator(context, system, name, serviceTypeName, getOptions)
	defer errext.PanicHandler("GetServiceWithOptions", errext.Panic, prefilledErr, &err)
	if isInitialized, initErr := system.isInitialized(); isInitialized {
		retrievedService := system.registry.Get(name)
		var applicableService = retrievedService
		applicableGetOptions := system.defaultGetOptions
		if len(getOptions) > 0 {
			applicableGetOptions = getOptions
		}
		if len(applicableGetOptions) > 0 {
			for _, getOption := range applicableGetOptions {
				logger.Log("Processing get option ", getOption)
				applicableService = getOption(context, system, name, serviceTypeName, getOptions, retrievedService, applicableService, &err)
			}
		}
		if err == nil {
			serviceInstance = applicableService
		} else {
			logger.Log("GetServiceWithOptions: skipping setting service instance due to errors", err)
		}
		return
	} else {
		err = prefilledErr(ErrSystemReasonSystemNotInitialized, initErr)
	}
	return
}

var defaultGetOptions = []GetOption{WithMatchingServiceType(errext.GenerateError|errext.Panic, nil)}

func ChangeInitializationGetOptions(getOptions ...GetOption) {
	defaultGetOptions = getOptions
}

func ChangeSystemGetOptions(system System, getOptions ...GetOption) (err error) {
	logger.Log("Entering ChangeDefaultGetOptions(", system, ",", getOptions, ")")
	defer func() { logger.Log("Exiting GetServiceFromSystem()", err) }()
	if asServiceMgr, isServiceMgr := system.(*systemImpl); isServiceMgr && asServiceMgr != nil {
		asServiceMgr.defaultGetOptions = getOptions
		return
	} else {
		err = ErrChangeSystemGetOptions.NewF(ErrSystemParamSystem, system,
			ErrSystemParamOptions, getOptions, ErrSystemParamReason, ErrSystemReasonSystemTypeNotSupported)
		return
	}
}

func GetSystemGetOptions(system System) (getOptions []GetOption) {
	logger.Log("Entering GetSystemGetOptions(", system, ")")
	defer func() { logger.Log("Exiting GetSystemGetOptions()", getOptions) }()
	if asServiceMgr, isServiceMgr := system.(*systemImpl); isServiceMgr && asServiceMgr != nil {
		getOptions = asServiceMgr.defaultGetOptions
		return
	}
	return
}

func WithMatchingServiceType(errorHandlingMode errext.HandleErrorMode, panicError any) GetOption {
	return func(context context.Context, system System, serviceIdentifier string, serviceTypeIdentifier string,
		options []GetOption, retrievedService any, previousApplicableValue any, err *error) (applicableValue any) {
		reason := ""
		if registeredServiceTypes := serviceTypeMapping.Get(serviceIdentifier); registeredServiceTypes != nil {
			if _, hasServiceType := registeredServiceTypes[serviceTypeIdentifier]; !hasServiceType {
				reason = ErrSystemReasonServiceTypeNotSupported
			}
		} else {
			reason = ErrSystemReasonServiceTypeNotSet
		}
		errext.HandleOptionError("WithMatchingServiceType", errorHandlingMode, err,
			WithGetErrorGenerator(context, system, serviceIdentifier, serviceTypeIdentifier, options),
			reason, panicError)
		return
	}
}

func GetOptionDoNothing() GetOption {
	return func(context context.Context, system System, serviceIdentifier string, serviceTypeIdentifier string, options []GetOption, retrievedService any, previousApplicableValue any, err *error) (applicableValue any) {
		applicableValue = previousApplicableValue
		return
	}
}

func WithGetErrorGenerator(context context.Context, system System, name string, serviceTypeName string,
	getOptions []GetOption, addFields ...interface{}) errext.ErrorGenerator {
	return func(reason string, err error, additionalFields ...interface{}) error {
		totalFields := append([]interface{}{errext.NewField(ErrSystemParamContext, context),
			errext.NewField(ErrSystemParamSystem, system), errext.NewField(ErrSystemParamServiceName, name),
			errext.NewField(ErrSystemParamServiceType, serviceTypeName), errext.NewField(ErrSystemParamOptions, getOptions),
		}, additionalFields...)
		totalFields = append(totalFields, addFields...)
		if err != nil {
			return ErrSystemGetServiceFailed.NewWithErrorF(err, totalFields...)
		} else {
			return ErrSystemGetServiceFailed.NewF(totalFields...)
		}
	}
}
