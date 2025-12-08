package telemetry

import (
	"context"
)

// Provider is the entry point for the telemetry API.
// It provides access to Tracers and Meters for instrumenting applications.
//
// Implementations must be safe for concurrent use.
type Provider interface {
	// Tracer returns a Tracer for creating spans.
	// The name should identify the instrumentation library or application component.
	// Calling Tracer multiple times with the same name returns the same Tracer instance.
	//
	// The name should follow reverse-DNS style naming (e.g., "com.example.myapp").
	Tracer(name string, opts ...any) (Tracer, error)

	// Meter returns a Meter for creating metric instruments.
	// The name should identify the instrumentation library or application component.
	// Calling Meter multiple times with the same name returns the same Meter instance.
	//
	// The name should follow reverse-DNS style naming (e.g., "com.example.myapp").
	Meter(name string, opts ...any) (Meter, error)

	// Shutdown shuts down the provider, flushing any remaining telemetry data.
	// After Shutdown is called, the provider should not be used.
	// Shutdown should be called before application exit to ensure all data is exported.
	//
	// The context can be used to set a timeout for the shutdown operation.
	Shutdown(ctx context.Context) error
}
