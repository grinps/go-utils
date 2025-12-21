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
//   - Instrument: Marker interface for all metric instruments (Instrument() method)
//   - Counter: Synchronous instrument for increments/decrements with Precision()
//   - Recorder: Synchronous instrument for point-in-time values with Precision()
//   - ObservableCounter: Async instrument for callback-based counter observations
//   - ObservableGauge: Async instrument for callback-based gauge observations
//   - Callback: Function type for async instrument observations
//   - Observer: Interface for reporting values in callbacks
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
//	// Retrieve provider from context (second param controls default fallback)
//	provider := telemetry.ContextTelemetry(ctx, true)  // falls back to Default()
//	provider := telemetry.ContextTelemetry(ctx, false) // returns nil if not found
//
// # Context-Based Tracer and Meter Access
//
// Tracers and meters can also be stored in and retrieved from context:
//
//	// Store and retrieve tracer
//	ctx = telemetry.ContextWithTracer(ctx, tracer)
//	tracer := telemetry.ContextTracer(ctx, true)       // falls back to noop
//	tracer, err := telemetry.ContextTracerE(ctx, true) // returns error on failure
//
//	// Store and retrieve meter
//	ctx = telemetry.ContextWithMeter(ctx, meter)
//	meter := telemetry.ContextMeter(ctx, true)         // falls back to noop
//	meter, err := telemetry.ContextMeterE(ctx, true)   // returns error on failure
//
// # Generic Instrument Creation
//
// Create type-safe instruments using the generic NewInstrument function:
//
//	counter, err := telemetry.NewInstrument[telemetry.Counter[int64]](ctx, "requests",
//	    telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)
//
// # Basic Usage
//
// Create a span from context tracer and use it to instrument your application:
//
//	// Start a span
//	ctx, span := telemetry.ContextTracer(ctx, true).Start(ctx, "operation-name")
//	defer span.End()
//
//	// Add attributes
//	span.SetAttributes("key", "value")
//
//	// Create an instrument
//	instrument, _ := telemetry.ContextMeter(ctx, true).NewInstrument("requests_total", telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)
//
// # Async Instruments
//
// Create callback-based instruments for values that are observed periodically:
//
//	// Observable counter with callback
//	callback := func(ctx context.Context, obs telemetry.Observer[int64]) {
//	    obs.Observe(getCurrentCount())
//	}
//	obsCounter, _ := meter.NewInstrument("active_connections",
//	    telemetry.InstrumentTypeObservableCounter,
//	    telemetry.CounterTypeMonotonic,
//	    callback)
//
//	// Observable gauge
//	gaugeCallback := func(ctx context.Context, obs telemetry.Observer[float64]) {
//	    obs.Observe(getCPUUsage())
//	}
//	obsGauge, _ := meter.NewInstrument("cpu_usage",
//	    telemetry.InstrumentTypeObservableGauge,
//	    gaugeCallback)
//
//	// Unregister when done
//	obsCounter.(telemetry.ObservableInstrument).Unregister()
//
// # Context Propagation
//
// The package uses Go's context.Context for span propagation:
//
//	func handleRequest(ctx context.Context) {
//	    ctx, span := telemetry.ContextTracer(ctx, true).Start(ctx, "handle-request")
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
