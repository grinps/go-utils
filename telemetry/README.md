# Telemetry Package

A vendor-agnostic API for application observability including distributed tracing and metrics collection.

## Overview

The telemetry package defines core interfaces for instrumenting Go applications with tracing and metrics capabilities. It follows OpenTelemetry semantics while providing a simplified, Go-idiomatic API.

## Installation

```bash
go get github.com/grinps/go-utils/telemetry
```

## Architecture

The package is organized around these core concepts:

| Component | Description |
|-----------|-------------|
| **Provider** | Entry point that creates Tracers and Meters |
| **Tracer** | Creates and manages Spans for distributed tracing |
| **Span** | Represents a unit of work with timing, attributes, and events |
| **Meter** | Creates metric instruments for measuring application behavior |
| **Instrument** | Base interface for all metric instruments |

## Quick Start

### Basic Tracing

```go
package main

import (
    "context"
    "github.com/grinps/go-utils/telemetry"
)

func main() {
    // Get the default provider (NoopProvider by default)
    provider := telemetry.Default()

    // Create a tracer
    tracer, err := provider.Tracer("my-service")
    if err != nil {
        panic(err)
    }

    // Start a span
    ctx, span := tracer.Start(context.Background(), "operation-name")
    defer span.End()

    // Add attributes
    span.SetAttributes("user.id", "12345")
    
    // Record an event
    span.AddEvent("processing-started")
    
    // Do work...
    doWork(ctx)
}

func doWork(ctx context.Context) {
    // Get tracer from context (falls back to default provider's tracer)
    tracer := telemetry.ContextTracer(ctx, true)
    
    // Child span automatically inherits parent context
    ctx, span := tracer.Start(ctx, "do-work")
    defer span.End()
    
    // Work happens here...
}
```

### Basic Metrics

```go
package main

import (
    "context"
    "github.com/grinps/go-utils/telemetry"
)

func main() {
    provider := telemetry.Default()

    // Create a meter
    meter, err := provider.Meter("my-service")
    if err != nil {
        panic(err)
    }

    // Create a counter instrument
    counter, err := meter.NewInstrument[telemetry.Counter[int64]]("requests_total",
        telemetry.InstrumentTypeCounter,
        telemetry.CounterTypeMonotonic,
    )
    if err != nil {
        panic(err)
    }

    // Use the counter
    c.Add(context.Background(), 1)
}
```

## Default Provider

The package includes a `NoopProvider` that performs no actual telemetry operations. This is the default provider and is useful for:

- Applications that don't need telemetry
- Testing without telemetry overhead
- Graceful degradation when telemetry backends are unavailable

```go
// Get the default provider (NoopProvider)
provider := telemetry.Default()

// Set a custom provider as default
telemetry.AsDefault(myProvider)
```

## Context Propagation

Providers, tracers, and meters can be stored in and retrieved from context for easy access throughout your application:

```go
// Store provider in context
ctx := telemetry.ContextWithTelemetry(ctx, provider)

// Retrieve provider from context (second param controls default fallback)
provider := telemetry.ContextTelemetry(ctx, true)  // falls back to Default()
provider := telemetry.ContextTelemetry(ctx, false) // returns nil if not found

// Store and retrieve tracer from context
ctx = telemetry.ContextWithTracer(ctx, tracer)
tracer := telemetry.ContextTracer(ctx, true)       // falls back to noop tracer
tracer, err := telemetry.ContextTracerE(ctx, true) // returns error if provider fails

// Store and retrieve meter from context
ctx = telemetry.ContextWithMeter(ctx, meter)
meter := telemetry.ContextMeter(ctx, true)         // falls back to noop meter
meter, err := telemetry.ContextMeterE(ctx, true)   // returns error if provider fails

// Create instrument using context's meter
counter, err := telemetry.NewInstrument[telemetry.Counter[int64]](ctx, "requests_total",
    telemetry.InstrumentTypeCounter,
    telemetry.CounterTypeMonotonic,
)
```

## Instrument Types

### Counters

Counters are for values that only increase (monotonic) or can increase and decrease (up-down):

```go
// Monotonic counter (only increases)
counter, _ := meter.NewInstrument[telemetry.Counter[int64]]("requests_total",
    telemetry.InstrumentTypeCounter,
    telemetry.CounterTypeMonotonic,
)

// Up-down counter (can increase or decrease)
updown, _ := meter.NewInstrument[telemetry.Counter[int64]]("active_connections",
    telemetry.InstrumentTypeCounter,
    telemetry.CounterTypeUpDown,
)
```

### Recorders

Recorders are for point-in-time values (gauges) or aggregated distributions (histograms):

```go
// Gauge (point-in-time value, no aggregation)
gauge, _ := meter.NewInstrument[telemetry.Recorder[int64]]("temperature",
    telemetry.InstrumentTypeRecorder,
    telemetry.AggregationStrategyNone,
)

// Histogram (aggregated distribution)
histogram, _ := meter.NewInstrument[telemetry.Recorder[int64]]("request_duration",
    telemetry.InstrumentTypeRecorder,
    telemetry.AggregationStrategyHistogram,
)
```

## Error Handling

The package provides structured error handling through error codes:

```go
tracer, err := provider.Tracer("my-service")
if err != nil {
    // Check for specific error types
    if errors.Is(err, telemetry.ErrTracerCreation) {
        // Handle tracer creation error
    }
}
```

### Error Handling Strategy

For testing purposes, you can control error behavior:

```go
// Generate errors for testing
tracer, err := provider.Tracer("test", telemetry.ErrorHandlingStrategyGenerateError)

// Return errors instead of ignoring them
inst, err := meter.NewInstrument[telemetry.Counter[int64]]("test", 
    telemetry.InstrumentTypeCounter,
    telemetry.ErrorHandlingStrategyReturn,
)
```

## Implementations

This package provides interfaces and a `NoopProvider`. Additional implementations are available:

| Package | Description |
|---------|-------------|
| `telemetry/memory` | In-memory implementation for testing |
| `telemetry/otel` | OpenTelemetry-based implementation for production |

## Thread Safety

All interfaces are designed to be safe for concurrent use. Implementations must ensure thread safety for all operations.

## API Reference

### Provider Interface

```go
type Provider interface {
    Tracer(name string, opts ...any) (Tracer, error)
    Meter(name string, opts ...any) (Meter, error)
    Shutdown(ctx context.Context) error
}
```

### Tracer Interface

```go
type Tracer interface {
    Start(ctx context.Context, name string, opts ...any) (context.Context, Span)
}
```

### Span Interface

```go
type Span interface {
    End(opts ...any)
    IsRecording() bool
    SetAttributes(attrs ...any)
    AddEvent(name string, opts ...any)
    RecordError(err error, opts ...any)
    SetStatus(code int, description string)
    SetName(name string)
    TracerProvider() Provider
}
```

### Meter Interface

```go
type Meter interface {
    NewInstrument(name string, opts ...any) (Instrument, error)
}
```

### Instrument Interface

```go
type Instrument interface {
    Name() string
    Description() string
    Unit() string
}
```

## License

See the LICENSE file in the repository root.
