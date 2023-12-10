package system

import (
	"github.com/grinps/go-utils/errext"
)

const (
	ErrSystemParamContext         = "context"
	ErrSystemParamServiceName     = "serviceName"
	ErrSystemParamServiceType     = "serviceType"
	ErrSystemParamServiceInstance = "serviceInstance"
	ErrSystemParamReason          = "reason"
	ErrSystemParamSystem          = "system"
	ErrSystemParamOptions         = "Options"
)

const (
	ErrSystemReasonSystemIsNil            = "system is nil"
	ErrSystemReasonMissingServiceRegistry = "no service registry is available."
)

var ErrSystemNotInitialized = errext.NewErrorCodeWithOptions(errext.WithTemplate("System is not initialized due to", "[reason]"))

const (
	ErrSystemReasonSystemTypeNotSupported = "type of system not supported (only *systemImpl supported at this time)."
	ErrSystemReasonSystemNotInitialized   = "system is not initialized."
	ErrSystemReasonNotAvailable           = "reason is not available"
)

var ErrChangeSystemGetOptions = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to change Get Options for system ", "["+ErrSystemParamSystem+"]", "to",
	"["+ErrSystemParamOptions+"]", "due to error", "["+ErrSystemParamReason+"]"))
var ErrChangeSystemRegistrationOptions = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to change registration Options for system ", "["+ErrSystemParamSystem+"]", "to",
	"["+ErrSystemParamOptions+"]", "due to error", "["+ErrSystemParamReason+"]"))

const (
	ErrSystemReasonServiceTypeNil                  = "service type is nil"
	ErrSystemReasonServiceTypeTransformationFailed = "service transformation failed"
	ErrSystemReasonServiceTypeNotSet               = "no service type was set for the service during registration"
	ErrSystemReasonServiceTypeNotSupported         = "given service type is not supported by service (please provide service type during registration)"
)

var ErrSystemGetServiceFailed = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to get service", "["+ErrSystemParamServiceName+"]", "of type",
	"["+ErrSystemParamServiceType+"]", "due to error", "["+ErrSystemParamReason+"]"))

const (
	ErrSystemReasonServiceNil                        = "service is nil"
	ErrSystemRegisterFailedReasonValueAlreadyPresent = "registration of a service already registered is not allowed."
	ErrSystemRegisterFailedReasonValueTypeMismatch   = "given value is not of expected type."
)

var ErrSystemRegisterServiceFailed = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to register service", "["+ErrSystemParamServiceInstance+"]", "with name", "["+ErrSystemParamServiceName+"]", "due to", "[reason]"))
var ErrSystemUnregisterServiceFailed = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to unregister namedService", "[serviceName]", "due to", "[reason]"))

const (
	ErrSystemReasonServiceNameEmpty             = "service name is empty"
	ErrSystemReasonServiceInstanceNil           = "service instance is nil"
	ErrSystemReasonServiceTypeEmpty             = "service type name is empty"
	ErrSystemReasonServiceTypeAlreadyRegistered = "service type is already registered"
)

var ErrServiceCreationFailed = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to create service with name",
	"["+ErrSystemParamServiceName+"]", "using instance", "["+ErrSystemParamServiceInstance+"]", "due to", "["+ErrSystemParamReason+"]"))
var ErrServiceTypeCreationFailed = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to create service type with name",
	"["+ErrSystemParamServiceType+"]", "due to", "["+ErrSystemParamReason+"]"))
