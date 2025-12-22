# Telemetry OTEL Package

An OpenTelemetry-based implementation of the telemetry interfaces using `go.opentelemetry.io/contrib/otelconf` for declarative configuration.

## Overview

The otel package provides a complete implementation of the `telemetry.Provider` interface using the OpenTelemetry Go SDK with otelconf declarative configuration. This is ideal for:

- **Production Deployments** - Export telemetry to backends like Jaeger, Zipkin, Prometheus, etc.
- **Declarative Configuration** - Uses otelconf.OpenTelemetryConfiguration schema
- **Config Package Integration** - Load configuration via config.Config.Unmarshal
- **Embedded Types** - Tracer and Meter embed their OpenTelemetry counterparts

## Installation

```bash
go get github.com/grinps/go-utils/telemetry/otel
```

## Quick Start

### Basic Tracing

```go
package main

import (
    "context"
    "log"

    "github.com/grinps/go-utils/telemetry/otel"
    "go.opentelemetry.io/otel/attribute"
)

func main() {
    ctx := context.Background()

    // Create provider with default configuration
    provider, err := otel.NewProvider(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer provider.Shutdown(ctx)

    // Create tracer
    tracer, _ := provider.Tracer("my-component")

    // Start a span
    ctx, span := tracer.Start(ctx, "operation")
    span.SetAttributes(attribute.String("user.id", "12345"))
    span.AddEvent("processing-started")
    span.End()
}
```

### Basic Metrics

```go
package main

import (
    "context"

    "github.com/grinps/go-utils/telemetry"
    "github.com/grinps/go-utils/telemetry/otel"
    "go.opentelemetry.io/otel/attribute"
)

func main() {
    ctx := context.Background()

    provider, _ := otel.NewProvider(ctx)
    defer provider.Shutdown(ctx)

    // Create meter
    meter, _ := provider.Meter("my-component")

    // Create a counter
    inst, _ := meter.NewInstrument("requests_total",
        telemetry.InstrumentTypeCounter,
        telemetry.CounterTypeMonotonic,
    )
    counter := inst.(telemetry.Counter[int64])

    // Record values with attributes
    counter.Add(ctx, 1, attribute.String("method", "GET"))
}
```

## Configuration

### Using otelconf.OpenTelemetryConfiguration

The provider uses the OpenTelemetry Configuration schema via `go.opentelemetry.io/contrib/otelconf`:

```go
import (
    "go.opentelemetry.io/contrib/otelconf"
)

// Create configuration directly
cfg := &otelconf.OpenTelemetryConfiguration{
    FileFormat: "0.3",
    // Configure tracer_provider, meter_provider, resource, etc.
}

provider, err := otel.NewProviderFromConfiguration(ctx, cfg)
```

### Loading from config.Config

Load configuration using the config package. The configuration is converted to YAML and parsed using `otelconf.ParseYAML`:

```go
import (
    "github.com/grinps/go-utils/config"
    "github.com/grinps/go-utils/config/ext"
)

// Create config with otelconf schema
cfg := ext.NewConfigWrapper(config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
    "opentelemetry": map[string]any{
        "file_format": "0.3",
        "resource": map[string]any{
            "schema_url":      "https://opentelemetry.io/schemas/1.39.0",
            "attributes_list": "service.name=my-service,service.namespace=my-namespace,service.version=1.0.0",
        },
        "tracer_provider": map[string]any{
            "processors": []any{
                map[string]any{
                    "batch": map[string]any{
                        "exporter": map[string]any{
                            "otlp_grpc": map[string]any{
                                "endpoint": "localhost:4317",
                                "insecure": true,
                            },
                        },
                    },
                },
            },
        },
    },
})))

provider, err := otel.NewProviderFromConfig(ctx, cfg)
```

### OTLP gRPC Configuration

For OTLP gRPC export, use the `otlp_grpc` exporter key:

```go
"tracer_provider": map[string]any{
    "processors": []any{
        map[string]any{
            "batch": map[string]any{
                "exporter": map[string]any{
                    "otlp_grpc": map[string]any{
                        "endpoint": "localhost:4317",
                        "insecure": true,  // For local development
                    },
                },
            },
        },
    },
},
```

**Resource configuration** uses `attributes_list` with comma-separated key=value pairs:

```go
"resource": map[string]any{
    "schema_url":      "https://opentelemetry.io/schemas/1.39.0",
    "attributes_list": "service.name=my-service,service.namespace=production,service.version=1.0.0",
},
```

### Provider Options

Simple configuration options for common use cases:

```go
provider, err := otel.NewProvider(ctx,
    otel.WithDisabled(false),
)
```

## Embedded Types

The package uses embedded types for Tracer, Meter, and Span:

```go
// Tracer embeds trace.Tracer with provider reference for TracerProvider() method
type Tracer struct {
    trace.Tracer
    provider *Provider
}

// Meter embeds metric.Meter (lightweight, no provider reference needed)
type Meter struct {
    metric.Meter
}

// Span embeds trace.Span with provider reference for TracerProvider() method
type Span struct {
    trace.Span
    provider *Provider
}
```

## Instrument Types

Instruments implement the `telemetry.Instrument` marker interface and provide `Precision()`:

```go
// Counters (Int64Counter, Int64UpDownCounter)
counter.Instrument()  // marker method
counter.Precision()   // returns telemetry.PrecisionInt64
counter.IsMonotonic() // true for Counter, false for UpDownCounter

// Recorders (Float64Gauge, Float64Histogram)
recorder.Instrument()           // marker method
recorder.Precision()            // returns telemetry.PrecisionFloat64
recorder.IsAggregating()        // false for Gauge, true for Histogram
recorder.AggregationStrategy()  // None or Histogram
```

This allows passing OpenTelemetry options directly:

```go
// Tracer with instrumentation version
tracer, _ := provider.Tracer("my-service",
    trace.WithInstrumentationVersion("1.0.0"),
)

// Span with OpenTelemetry options
ctx, span := tracer.Start(ctx, "operation",
    trace.WithSpanKind(trace.SpanKindServer),
    trace.WithAttributes(attribute.String("key", "value")),
)
```

## Status Codes

The package provides status code constants:

```go
const (
    StatusUnset StatusCode = iota
    StatusOK
    StatusError
)

span.SetStatus(int(otel.StatusOK), "success")
span.SetStatus(int(otel.StatusError), "failed")
```

## Environment Variables

The OpenTelemetry SDK recognizes these environment variables:

- `OTEL_SERVICE_NAME`: Service name for resource
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OTLP endpoint URL
- `OTEL_EXPORTER_OTLP_PROTOCOL`: Protocol (grpc or http/protobuf)
- `OTEL_EXPORTER_OTLP_HEADERS`: Headers for OTLP requests

## Thread Safety

All types in this package are safe for concurrent use.

## Error Handling

The package uses `github.com/grinps/go-utils/errext` for structured error handling:

```go
// Error codes with type classification
var (
    ErrCodeProviderCreation  // Error creating provider
    ErrCodeConfigInvalid     // Invalid configuration
    ErrCodeShutdown          // Shutdown failure
    ErrCodeAlreadyShutdown   // Provider already shutdown
)

// Sentinel errors for simple checks
var (
    ErrProviderCreation   = ErrCodeProviderCreation.New("failed to create OTEL provider")
    ErrAlreadyShutdown    = ErrCodeAlreadyShutdown.New("provider already shutdown")
)

// Check error type
if errext.Is(err, otel.ErrCodeProviderCreation) {
    // Handle provider creation error
}
```

## Integration Testing

### Running the OpenTelemetry Collector

To run integration tests that export telemetry to an OTLP collector:

```bash
# Start the OpenTelemetry Collector with Docker
docker run \
  -p 127.0.0.1:4317:4317 \
  -p 127.0.0.1:4318:4318 \
  -p 127.0.0.1:55679:55679 \
  --name "otel-collector" \
  otel/opentelemetry-collector:0.141.0 \
  --config /etc/otelcol/config.yaml \
  --config 'yaml:service::pipelines::metrics::receivers: [otlp]'
```

**Ports:**
- `4317`: OTLP gRPC receiver
- `4318`: OTLP HTTP receiver
- `55679`: zpages extension (debugging)

### Running Integration Tests

```bash
# Run integration tests (requires collector running)
go test -tags=integration ./...

# Run with verbose output
go test -tags=integration -v ./...
```

### Integration Test Configuration

The integration tests use this configuration:

```go
var cfg = ext.NewConfigWrapper(config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
    "opentelemetry": map[string]any{
        "file_format": "0.3",
        "resource": map[string]any{
            "schema_url":      "https://opentelemetry.io/schemas/1.39.0",
            "attributes_list": "service.name=my-service,service.namespace=integration_test,service.version=1.0.0",
        },
        "tracer_provider": map[string]any{
            "processors": []any{
                map[string]any{
                    "batch": map[string]any{
                        "exporter": map[string]any{
                            "otlp_grpc": map[string]any{
                                "endpoint": "localhost:4317",
                                "insecure": true,
                            },
                        },
                    },
                },
            },
        },
    },
})))
```

## API Reference

### Provider

- `NewProvider(ctx, opts...) (*Provider, error)` - Create with options
- `NewProviderFromConfig(ctx, config.Config) (*Provider, error)` - Create from config.Config
- `Tracer(name, opts...) (telemetry.Tracer, error)` - Get a tracer
- `Meter(name, opts...) (telemetry.Meter, error)` - Get a meter
- `Shutdown(ctx) error` - Shutdown the provider

### Configuration Functions

- `LoadConfiguration(ctx, config.Config) (*otelconf.OpenTelemetryConfiguration, error)` - Load from config (uses YAML + otelconf.ParseYAML)
- `DefaultConfiguration() *otelconf.OpenTelemetryConfiguration` - Get default configuration

### Error Codes

- `ErrCodeProviderCreation` - Provider creation failed
- `ErrCodeConfigLoadFailed` - Configuration loading failed
- `ErrCodeConfigInvalid` - Invalid configuration
- `ErrCodeShutdown` - Shutdown failed
