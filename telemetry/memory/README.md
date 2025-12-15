# Telemetry Memory Package

An in-memory implementation of the telemetry interfaces for testing and development.

## Overview

The memory package provides a complete implementation of the `telemetry.Provider` interface that stores all telemetry data in memory. This is ideal for:

- **Unit Testing** - Verify instrumentation without external dependencies
- **Integration Testing** - Assert on spans, attributes, and metrics
- **Development** - Quick local development without telemetry backends
- **Debugging** - Inspect recorded telemetry data

## Installation

```bash
go get github.com/grinps/go-utils/telemetry/memory
```

## Quick Start

### Basic Tracing

```go
package main

import (
    "context"
    "fmt"
    
    "github.com/grinps/go-utils/telemetry/memory"
)

func main() {
    // Create provider
    provider := memory.NewProvider()
    defer provider.Shutdown(context.Background())

    // Create tracer
    tracer, _ := provider.Tracer("my-service")

    // Start a span
    ctx, span := tracer.Start(context.Background(), "operation")
    span.SetAttributes(memory.String("user.id", "12345"))
    span.AddEvent("processing-started")
    span.End()

    // Access recorded spans
    spans := provider.RecordedSpans()
    fmt.Printf("Recorded %d spans\n", len(spans))
}
```

### Basic Metrics

```go
package main

import (
    "context"
    
    "github.com/grinps/go-utils/telemetry"
    "github.com/grinps/go-utils/telemetry/memory"
)

func main() {
    provider := memory.NewProvider()
    defer provider.Shutdown(context.Background())

    // Create meter
    meter, _ := provider.Meter("my-service")

    // Create a counter
    inst, _ := meter.NewInstrument("requests_total",
        telemetry.InstrumentTypeCounter,
        telemetry.CounterTypeMonotonic,
    )
    counter := inst.(telemetry.Counter[int64])

    // Record values
    counter.Add(context.Background(), 1)
    counter.Add(context.Background(), 5)

    // Access recorded measurements
    m := meter.(*memory.Meter)
    measurements := m.RecordedMeasurements()
    fmt.Printf("Recorded %d measurements\n", len(measurements))
}
```

## Testing Usage

### Asserting on Spans

```go
func TestMyFunction(t *testing.T) {
    provider := memory.NewProvider()
    tracer, _ := provider.Tracer("test")

    // Call function under test
    ctx, span := tracer.Start(context.Background(), "test-operation")
    span.SetAttributes(memory.String("key", "value"))
    span.End()

    // Assert on recorded spans
    spans := provider.RecordedSpans()
    if len(spans) != 1 {
        t.Fatalf("expected 1 span, got %d", len(spans))
    }

    span := spans[0]
    if span.Name != "test-operation" {
        t.Errorf("unexpected span name: %s", span.Name)
    }
    if !span.HasAttribute("key") {
        t.Error("expected attribute 'key'")
    }
    if span.GetAttribute("key") != "value" {
        t.Error("unexpected attribute value")
    }
}
```

### Asserting on Span Relationships

```go
func TestParentChildSpans(t *testing.T) {
    provider := memory.NewProvider()
    tracer, _ := provider.Tracer("test")

    // Create parent span
    ctx, parentSpan := tracer.Start(context.Background(), "parent")
    
    // Create child span
    _, childSpan := tracer.Start(ctx, "child")
    childSpan.End()
    parentSpan.End()

    spans := provider.RecordedSpans()
    
    // Find spans
    var parent, child *memory.RecordedSpan
    for _, s := range spans {
        if s.Name == "parent" {
            parent = s
        } else {
            child = s
        }
    }

    // Verify relationship
    if child.ParentSpanID != parent.SpanContext.SpanID() {
        t.Error("child should reference parent")
    }
    if child.SpanContext.TraceID() != parent.SpanContext.TraceID() {
        t.Error("spans should share trace ID")
    }
}
```

### Asserting on Metrics

```go
func TestMetrics(t *testing.T) {
    provider := memory.NewProvider()
    meter, _ := provider.Meter("test")

    // Create and use counter
    inst, _ := meter.NewInstrument("test_counter",
        telemetry.InstrumentTypeCounter,
        telemetry.CounterTypeMonotonic,
    )
    counter := inst.(telemetry.Counter[int64])
    counter.Add(context.Background(), 10)

    // Assert on measurements
    m := meter.(*memory.Meter)
    measurements := m.RecordedMeasurements()
    
    if len(measurements) != 1 {
        t.Fatalf("expected 1 measurement, got %d", len(measurements))
    }
    if measurements[0].Value != int64(10) {
        t.Errorf("unexpected value: %v", measurements[0].Value)
    }
}
```

## API Reference

### Provider

```go
// Create a new in-memory provider
provider := memory.NewProvider()

// Get tracer (returns telemetry.Tracer, error)
tracer, err := provider.Tracer("name")

// Get meter (returns telemetry.Meter, error)
meter, err := provider.Meter("name")

// Shutdown provider
err := provider.Shutdown(ctx)

// Access recorded spans
spans := provider.RecordedSpans()
spans := provider.RecordedSpansByName("operation-name")

// Reset recorded data
provider.Reset()

// Check shutdown status
isShutdown := provider.IsShutdown()
```

### RecordedSpan

```go
type RecordedSpan struct {
    Name         string
    SpanContext  SpanContext
    ParentSpanID SpanID
    Kind         SpanKind
    StartTime    time.Time
    EndTime      time.Time
    Attributes   []Attribute
    Events       []Event
    Links        []Link
    Status       Status
    TracerName   string
}

// Helper methods
span.HasAttribute("key")      // bool
span.GetAttribute("key")      // any
span.HasEvent("event-name")   // bool
span.Duration()               // time.Duration
```

### Meter

```go
// Create instruments
inst, err := meter.NewInstrument("name",
    telemetry.InstrumentTypeCounter,
    telemetry.CounterTypeMonotonic,
)

// Access recorded measurements
m := meter.(*memory.Meter)
measurements := m.RecordedMeasurements()
measurements := m.RecordedMeasurementsByName("counter-name")
```

### Attribute Helpers

```go
memory.String("key", "value")    // String attribute
memory.Int64("key", 42)          // Int64 attribute
memory.Float64("key", 3.14)      // Float64 attribute
memory.Bool("key", true)         // Bool attribute
```

### Span Kinds

```go
memory.SpanKindUnspecified
memory.SpanKindInternal
memory.SpanKindServer
memory.SpanKindClient
memory.SpanKindProducer
memory.SpanKindConsumer
```

### Status Codes

```go
memory.StatusUnset
memory.StatusOK
memory.StatusError
```

## Minimal Dependency Usage

The package supports key-value pair options to reduce dependency on memory package types. This allows your instrumentation code to be more portable across different telemetry implementations.

### Tracer and Meter Options

Instead of using config structs, pass options as key-value pairs. Use `"version"` and `"schemaURL"` as special keys, and any other string keys create attributes:

```go
// Using key-value pairs (minimal dependency)
tracer, _ := provider.Tracer("my-service", 
    "version", "1.0.0", 
    "schemaURL", "http://example.com/schema",
    "service.env", "production",      // Creates attribute
    "service.region", "us-east-1",    // Creates attribute
)

meter, _ := provider.Meter("my-service", 
    "version", "1.0.0",
    "service.name", "my-app",         // Creates attribute
)
```

### Attributes as Key-Value Pairs

Instead of using `memory.String()`, `memory.Int64()`, etc., pass attributes as alternating key-value pairs:

```go
// Using key-value pairs (minimal dependency)
counter.Add(ctx, 1, "user.id", "12345", "request.size", 1024)
span.SetAttributes("key1", "value1", "key2", 42)
span.AddEvent("my-event", "detail", "something happened", "count", 5)

// Equivalent using Attribute helpers
counter.Add(ctx, 1, memory.String("user.id", "12345"), memory.Int64("request.size", 1024))
span.SetAttributes(memory.String("key1", "value1"), memory.Int64("key2", 42))
```

### Events with Timestamp

For events, the first `time.Time` value is used as the timestamp:

```go
import "time"

// Event with custom timestamp and key-value attributes
span.AddEvent("cache-hit", time.Now(), "cache.key", "user:123", "ttl", 300)
```

## Thread Safety

All operations on the Provider, Tracer, Meter, and Span are thread-safe. Multiple goroutines can safely:

- Create spans concurrently
- Record metrics concurrently
- Read recorded data while writes are happening

## Integration with Main Telemetry Package

```go
import (
    "github.com/grinps/go-utils/telemetry"
    "github.com/grinps/go-utils/telemetry/memory"
)

func setupTelemetry() {
    provider := memory.NewProvider()
    
    // Set as default provider
    telemetry.AsDefault(provider)
    
    // Use via context
    ctx := telemetry.ContextWithTelemetry(context.Background(), provider)
    tracer := telemetry.ContextTracer(ctx, true)
}
```

## License

See the LICENSE file in the repository root.
