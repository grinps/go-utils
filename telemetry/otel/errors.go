package otel

import (
	"github.com/grinps/go-utils/errext"
)

// Error type for the otel package.
const ErrorTypeOtel = "github.com/grinps/go-utils/telemetry/otel"

// Error codes for the otel package.
// Using non-unique error codes with type classification.
var (
	// ErrCodeProviderCreation is returned when provider creation fails.
	ErrCodeProviderCreation = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 1, ErrorTypeOtel),
		errext.WithAttributes("component", "provider"),
	)

	// ErrCodeExporterCreation is returned when exporter creation fails.
	ErrCodeExporterCreation = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 2, ErrorTypeOtel),
		errext.WithAttributes("component", "exporter"),
	)

	// ErrCodeResourceCreation is returned when resource creation fails.
	ErrCodeResourceCreation = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 3, ErrorTypeOtel),
		errext.WithAttributes("component", "resource"),
	)

	// ErrCodeTracerProviderCreation is returned when tracer provider creation fails.
	ErrCodeTracerProviderCreation = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 4, ErrorTypeOtel),
		errext.WithAttributes("component", "tracer_provider"),
	)

	// ErrCodeMeterProviderCreation is returned when meter provider creation fails.
	ErrCodeMeterProviderCreation = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 5, ErrorTypeOtel),
		errext.WithAttributes("component", "meter_provider"),
	)

	// ErrCodeInstrumentCreation is returned when instrument creation fails.
	ErrCodeInstrumentCreation = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 6, ErrorTypeOtel),
		errext.WithAttributes("component", "instrument"),
	)

	// ErrCodeConfigInvalid is returned when configuration is invalid.
	ErrCodeConfigInvalid = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 7, ErrorTypeOtel),
		errext.WithAttributes("component", "config"),
	)

	// ErrCodeShutdown is returned when shutdown fails.
	ErrCodeShutdown = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 8, ErrorTypeOtel),
		errext.WithAttributes("component", "shutdown"),
	)

	// ErrCodeAlreadyShutdown is returned when provider is already shutdown.
	ErrCodeAlreadyShutdown = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 9, ErrorTypeOtel),
		errext.WithAttributes("component", "shutdown"),
	)

	// ErrCodeConfigLoadFailed is returned when configuration loading fails.
	ErrCodeConfigLoadFailed = errext.NewErrorCodeWithOptions(
		errext.WithErrorCodeAndType(false, 10, ErrorTypeOtel),
		errext.WithAttributes("component", "config"),
	)
)
