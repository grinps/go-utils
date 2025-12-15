// Package memory provides an in-memory implementation of the telemetry interfaces.
//
// This package is intended for testing and development purposes. It stores all
// telemetry data in memory, allowing tests to verify that instrumentation is
// working correctly without requiring a real telemetry backend.
//
// # Features
//
//   - Full implementation of telemetry.Provider, telemetry.Tracer, telemetry.Span, and telemetry.Meter interfaces
//   - Thread-safe operations
//   - Access to recorded spans and metrics for test assertions
//   - Support for span relationships and context propagation
//   - Counter and Recorder instrument implementations
//   - Key-value pair options for minimal memory package dependencies
//
// # Basic Usage
//
//	provider := memory.NewProvider()
//	defer provider.Shutdown(context.Background())
//
//	// Use the provider for tracing
//	tracer, _ := provider.Tracer("test-service")
//	ctx, span := tracer.Start(context.Background(), "test-operation")
//	span.SetAttributes(memory.String("key", "value"))
//	span.End()
//
//	// Access recorded spans for assertions
//	spans := provider.RecordedSpans()
//	if len(spans) != 1 {
//	    t.Errorf("expected 1 span, got %d", len(spans))
//	}
//
//	// Use the provider for metrics
//	meter, _ := provider.Meter("test-service")
//	counter, _ := meter.NewInstrument("requests_total",
//	    telemetry.InstrumentTypeCounter,
//	    telemetry.CounterTypeMonotonic,
//	)
//	counter.(telemetry.Counter[int64]).Add(ctx, 1)
//
//	// Access recorded metrics for assertions
//	m := meter.(*memory.Meter)
//	measurements := m.RecordedMeasurements()
//
// # Minimal Dependency Usage
//
// The package supports key-value pair options to reduce dependency on memory
// package types. This is useful when you want your instrumentation code to be
// independent of the specific telemetry implementation.
//
// For Tracer and Meter options, use "version" and "schemaURL" as special keys,
// and any other string keys create attributes:
//
//	tracer, _ := provider.Tracer("my-service", "version", "1.0.0", "service.env", "prod")
//	meter, _ := provider.Meter("my-service", "version", "1.0.0", "service.region", "us-east-1")
//
// For attributes on instruments and spans, use string key followed by any value:
//
//	counter.Add(ctx, 1, "user.id", "12345", "request.size", 1024)
//	span.AddEvent("my-event", "key1", "value1", "key2", 42)
//
// # Testing Assertions
//
// The package provides helper types and methods for test assertions:
//
//	// Check span attributes
//	span := provider.RecordedSpans()[0]
//	if !span.HasAttribute("key") {
//	    t.Error("expected attribute not found")
//	}
//	if span.GetAttribute("key") != "value" {
//	    t.Error("unexpected attribute value")
//	}
//
//	// Check span relationships
//	if span.ParentSpanID != expectedParentID {
//	    t.Error("incorrect parent relationship")
//	}
//
//	// Check span duration
//	if span.Duration() == 0 {
//	    t.Error("expected non-zero duration")
//	}
//
// # Thread Safety
//
// All operations on the in-memory provider and its components are thread-safe.
// Multiple goroutines can safely create spans, record metrics, and read
// recorded data concurrently.
package memory
