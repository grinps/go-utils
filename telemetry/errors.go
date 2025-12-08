package telemetry

import (
	"github.com/grinps/go-utils/errext"
)

// Error parameter constants for structured error messages.
const (
	// ErrParamName is the parameter name for entity names (span, meter, instrument).
	ErrParamName = "name"
	// ErrParamReason is the parameter name for error reasons.
	ErrParamReason = "reason"
	// ErrParamType is the parameter name for types.
	ErrParamType = "type"
	// ErrParamOperation is the parameter name for operation names.
	ErrParamOperation = "operation"
	// ErrParamValue is the parameter name for values.
	ErrParamValue = "value"
	// ErrParamKey is the parameter name for attribute keys.
	ErrParamKey = "key"
	// ErrParamInstrumentType is the parameter name for instrument types.
	ErrParamInstrumentType = "instrumentType"
	// ErrParamCounterType is the parameter name for counter types.
	ErrParamCounterType = "counterType"
	// ErrParamAggregationStrategy is the parameter name for aggregation strategies.
	ErrParamAggregationStrategy = "aggregationStrategy"
	// ErrParamOption is the parameter name for options.
	ErrParamOption = "option"
)

// Common error reasons.
const (
	// ErrReasonNilProvider indicates a nil provider was used.
	ErrReasonNilProvider = "provider is nil"
	// ErrReasonNilTracer indicates a nil tracer was used.
	ErrReasonNilTracer = "tracer is nil"
	// ErrReasonNilMeter indicates a nil meter was used.
	ErrReasonNilMeter = "meter is nil"
	// ErrReasonNilSpan indicates a nil span was used.
	ErrReasonNilSpan = "span is nil"
	// ErrReasonNilContext indicates a nil context was used.
	ErrReasonNilContext = "context is nil"
	// ErrReasonEmptyName indicates an empty name was provided.
	ErrReasonEmptyName = "name cannot be empty"
	// ErrReasonInvalidValue indicates an invalid value was provided.
	ErrReasonInvalidValue = "invalid value"
	// ErrReasonAlreadyShutdown indicates the provider was already shutdown.
	ErrReasonAlreadyShutdown = "provider already shutdown"
	// ErrReasonNotInitialized indicates the component was not initialized.
	ErrReasonNotInitialized = "not initialized"
	// ErrReasonInstrumentExists indicates an instrument with the same name already exists.
	ErrReasonInstrumentExists = "instrument with same name already exists with different configuration"
	// ErrReasonInvalidInstrumentType indicates an invalid instrument type was provided.
	ErrReasonInvalidInstrumentType = "invalid instrument type"
	// ErrReasonInvalidCounterType indicates an invalid counter type was provided.
	ErrReasonInvalidCounterType = "invalid counter type"
	// ErrReasonInvalidAggregationStrategy indicates an invalid aggregation strategy was provided.
	ErrReasonInvalidAggregationStrategy = "invalid aggregation strategy"
	// ErrReasonInvalidOption indicates an invalid option was provided.
	ErrReasonInvalidOption = "invalid option"
)

// ErrTypePrefix is the error type prefix for all telemetry errors.
const ErrTypePrefix = "github.com/grinps/go-utils/telemetry"

// Error codes for the telemetry package.
var (
	// ErrProviderCreation is returned when provider creation fails.
	ErrProviderCreation = errext.NewErrorCodeOfType(1, ErrTypePrefix)

	// ErrProviderShutdown is returned when provider shutdown fails.
	ErrProviderShutdown = errext.NewErrorCodeOfType(2, ErrTypePrefix)

	// ErrTracerCreation is returned when tracer creation fails.
	ErrTracerCreation = errext.NewErrorCodeOfType(3, ErrTypePrefix)

	// ErrSpanCreation is returned when span creation fails.
	ErrSpanCreation = errext.NewErrorCodeOfType(4, ErrTypePrefix)

	// ErrSpanOperation is returned when a span operation fails.
	ErrSpanOperation = errext.NewErrorCodeOfType(5, ErrTypePrefix)

	// ErrMeterCreation is returned when meter creation fails.
	ErrMeterCreation = errext.NewErrorCodeOfType(6, ErrTypePrefix)

	// ErrInstrumentCreation is returned when instrument creation fails.
	ErrInstrumentCreation = errext.NewErrorCodeOfType(7, ErrTypePrefix)

	// ErrInstrumentOperation is returned when an instrument operation fails.
	ErrInstrumentOperation = errext.NewErrorCodeOfType(8, ErrTypePrefix)

	// ErrInvalidAttribute is returned when an invalid attribute is provided.
	ErrInvalidAttribute = errext.NewErrorCodeOfType(9, ErrTypePrefix)

	// ErrContextPropagation is returned when context propagation fails.
	ErrContextPropagation = errext.NewErrorCodeOfType(10, ErrTypePrefix)
)
