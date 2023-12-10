package system

import (
	"context"
	logger "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/base-utils/registry"
	"github.com/grinps/go-utils/errext"
	"reflect"
)

func RegisterServiceWithSystem[T any](context context.Context, system System, service Service[T],
	options ...RegistrationOption) (err error) {
	logger.Log("Entering RegisterServiceWithSystem(", context, ",", system, ",", service, ",", options, ")")
	defer func() { logger.Log("Exiting RegisterServiceWithSystem()", err) }()
	if service == nil {
		return WithRegistrationErrorGenerator(context, system,
			"NilService", "NilService", options)(ErrSystemReasonServiceNil, nil)
	}
	applicableServiceName := service.String()
	if asComparable, isComparable := service.(Comparable[string]); isComparable {
		applicableServiceName = asComparable.Key()
	}
	applicableRegistrationOptions := append(GetSystemRegistrationOptions(system), options...)
	if asRegistrySystem, isRegistrySystem := system.(RegistrySystem); isRegistrySystem {
		err = asRegistrySystem.RegisterService(context, applicableServiceName, service.AsService(), applicableRegistrationOptions...)
	} else {
		err = WithRegistrationErrorGenerator(context, system,
			service.String(), service.AsService(), options)(ErrSystemReasonSystemTypeNotSupported, nil,
			errext.NewField("SystemType", reflect.TypeOf(system)))
	}
	return
}

func (system *systemImpl) RegisterService(context context.Context, serviceIdentifier string, newService any, registrationOptions ...RegistrationOption) (err error) {
	logger.Log("Entering RegisterServiceWithOptions(", context, ",", system, ",", serviceIdentifier, ",", newService, ",", registrationOptions, ")")
	defer func() { logger.Log("Exiting RegisterServiceWithOptions()", err) }()
	prefilledErr := WithRegistrationErrorGenerator(context, system, serviceIdentifier, newService, registrationOptions)
	defer errext.PanicHandler("RegisterServiceWithOptions", errext.Panic, prefilledErr, &err)
	if isInitialized, initErr := system.isInitialized(); isInitialized {
		applicableValue := newService
		applicableRegistrationOptions := system.defaultRegistrationOptions
		if len(registrationOptions) > 0 {
			applicableRegistrationOptions = registrationOptions
		}
		if len(applicableRegistrationOptions) > 0 {
			oldValue := system.registry.Get(serviceIdentifier)
			for _, registrationOption := range applicableRegistrationOptions {
				logger.Log("Processing registration option ", registrationOption)
				returnedValue := registrationOption(context, system, serviceIdentifier, newService, registrationOptions, oldValue, applicableValue, &err)
				if err != nil {
					return
				}
				applicableValue = returnedValue
			}
		}
		_ = system.registry.Register(serviceIdentifier, applicableValue)
	} else {
		return prefilledErr(ErrSystemReasonSystemNotInitialized, initErr)
	}
	return
}

var defaultRegistrationOptions = []RegistrationOption{
	RegistrationOptionAllowOnetimeRegistration(errext.GenerateError, nil),
}

func ChangeInitializationRegistrationOptions(registrationOptions ...RegistrationOption) {
	defaultRegistrationOptions = registrationOptions
}

func ChangeSystemRegistrationOptions(system System, registrationOptions ...RegistrationOption) (err error) {
	logger.Log("Entering ChangeSystemRegistrationOptions(", system, ",", registrationOptions, ")")
	defer func() { logger.Log("Exiting ChangeSystemRegistrationOptions()", err) }()
	if asServiceMgr, isServiceMgr := system.(*systemImpl); isServiceMgr && asServiceMgr != nil {
		asServiceMgr.defaultRegistrationOptions = registrationOptions
		return
	} else {
		err = ErrChangeSystemRegistrationOptions.NewF(ErrSystemParamSystem, system,
			ErrSystemParamOptions, registrationOptions, ErrSystemParamReason, ErrSystemReasonSystemTypeNotSupported)
		return
	}
}

func GetSystemRegistrationOptions(system System) (registrationOptions []RegistrationOption) {
	logger.Log("Entering GetSystemRegistrationOptions(", system, ")")
	defer func() { logger.Log("Exiting GetSystemRegistrationOptions()", registrationOptions) }()
	if asServiceMgr, isServiceMgr := system.(*systemImpl); isServiceMgr && asServiceMgr != nil {
		registrationOptions = asServiceMgr.defaultRegistrationOptions
		return
	}
	return
}

// RegistrationOptionAllowOnetimeRegistration returns old value instead of new in case two don't match.
// This option should be used to ensure only one service instance for a given name is registered.
func RegistrationOptionAllowOnetimeRegistration(errorHandleMode errext.HandleErrorMode, panicError any) RegistrationOption {
	return func(context context.Context, system System, serviceName string, newValue any, options []RegistrationOption, oldValue any, previousApplicableValue any, err *error) (applicableValue any) {
		applicableValue = newValue
		reason := ""
		if oldValue != nil && oldValue != newValue {
			//TODO: Comparable check
			applicableValue = oldValue
			reason = ErrSystemRegisterFailedReasonValueAlreadyPresent
		}
		errext.HandleOptionError("RegistrationOptionAllowOnetimeRegistration", errorHandleMode, err,
			WithRegistrationErrorGenerator(context, system, serviceName, newValue, options),
			reason, panicError)
		return
	}
}

var serviceTypeMapping = registry.NewRegister[string, map[string]struct{}]()

func WithServiceType[T any](serviceType ServiceType[T]) RegistrationOption {
	applicableServiceTypeName := serviceType.String()
	if asComparable, isComparable := serviceType.(Comparable[string]); isComparable {
		applicableServiceTypeName = asComparable.Key()
	}
	return func(context context.Context, system System, serviceName string, newValue any, options []RegistrationOption, oldValue any, previousApplicableValue any, err *error) (applicableValue any) {
		var applicableServiceTypeList = map[string]struct{}{}
		if listOfTypes := serviceTypeMapping.Get(serviceName); listOfTypes != nil {
			applicableServiceTypeList = listOfTypes
		} else {
			serviceTypeMapping.Register(serviceName, applicableServiceTypeList)
		}
		applicableServiceTypeList[applicableServiceTypeName] = struct{}{}
		return
	}
}

func RegistrationOptionDoNothing() RegistrationOption {
	return func(context context.Context, system System, serviceName string, newValue any, options []RegistrationOption, oldValue any, previousApplicableValue any, err *error) (applicableValue any) {
		applicableValue = previousApplicableValue
		return
	}
}

func WithRegistrationErrorGenerator[T any](context context.Context, system System, serviceIdentifier string, newService T,
	registrationOptions []RegistrationOption, addFields ...interface{}) errext.ErrorGenerator {
	return func(reason string, err error, additionalFields ...interface{}) error {
		totalFields := append([]interface{}{errext.NewField(ErrSystemParamContext, context),
			errext.NewField(ErrSystemParamSystem, system), errext.NewField(ErrSystemParamServiceName, serviceIdentifier),
			errext.NewField(ErrSystemParamServiceInstance, newService), errext.NewField(ErrSystemParamOptions, registrationOptions),
			errext.NewField(ErrSystemParamReason, reason),
		}, additionalFields...)
		totalFields = append(totalFields, addFields...)
		if err != nil {
			return ErrSystemRegisterServiceFailed.NewWithErrorF(err, totalFields...)
		} else {
			return ErrSystemRegisterServiceFailed.NewF(totalFields...)
		}
	}
}
