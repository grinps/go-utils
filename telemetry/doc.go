// Package telemetry provides a vendor-agnostic API for application observability
// including distributed tracing and metrics collection.
//
// # Overview
//
// The telemetry package defines core interfaces for instrumenting Go applications
// with tracing and metrics capabilities. It follows OpenTelemetry semantics while
// providing a simplified, Go-idiomatic API.
//
// # Architecture
//
// The package is organized around these core concepts:
//
//   - Provider: Entry point that creates Tracers and Meters
//   - Tracer: Creates and manages Spans for distributed tracing
//   - Span: Represents a unit of work with timing, attributes, and events
//   - Meter: Creates metric instruments for measuring application behavior
//   - Instrument: Base interface for all metric instruments
//
// # Default Provider
//
// The package supports a global default provider that can be set and retrieved.
// By default, a NoopProvider is used until a custom provider is set:
//
//	// Set a custom provider as default
//	telemetry.AsDefault(myProvider)
//
//	// Get the default provider (returns NoopProvider if not set)
//	provider := telemetry.Default()
//
// # Context-Based Provider Access
//
// Providers can be stored in and retrieved from context:
//
//	// Store provider in context
//	ctx := telemetry.ContextWithTelemetry(ctx, provider)
//
//	// Retrieve provider from context (falls back to Default)
//	provider := telemetry.ContextTelemetry(ctx)
//
// # Convenience Functions
//
// Helper functions for quick access to tracers and meters:
//
//	// Get a tracer using context's provider
//	tracer, _ := telemetry.NewTracer(ctx, "my-service")
//
//	// Get a meter using context's provider
//	meter, _ := telemetry.NewMeter(ctx, "my-service")
//
// # Basic Usage
//
// Create a provider and use it to instrument your application:
//
//	// Get a tracer from the provider
//	tracer, err := provider.Tracer("my-service")
//	if err != nil {
//	    // handle error
//	}
//
//	// Start a span
//	ctx, span := tracer.Start(ctx, "operation-name")
//	defer span.End()
//
//	// Add attributes
//	span.SetAttributes("key", "value")
//
//	// Get a meter for metrics
//	meter, err := provider.Meter("my-service")
//	if err != nil {
//	    // handle error
//	}
//
//	// Create an instrument
//	instrument, _ := meter.NewInstrument("requests_total")
//
// # Context Propagation
//
// The package uses Go's context.Context for span propagation:
//
//	func handleRequest(ctx context.Context) {
//	    ctx, span := tracer.Start(ctx, "handle-request")
//	    defer span.End()
//
//	    // Child spans automatically inherit parent context
//	    processData(ctx)
//	}
//
// # Implementations
//
// This package provides interfaces and a NoopProvider. Additional implementations are available in:
//
//   - telemetry/memory: In-memory implementation for testing
//   - telemetry/otel: OpenTelemetry-based implementation for production
//
// # Thread Safety
//
// All interfaces are designed to be safe for concurrent use. Implementations
// must ensure thread safety for all operations.
package telemetry
