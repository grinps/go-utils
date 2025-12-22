// Package otel provides an OpenTelemetry-based implementation of the telemetry interfaces.
//
// This package wraps the OpenTelemetry Go SDK to provide a simplified API
// that conforms to the telemetry package interfaces. It uses go.opentelemetry.io/contrib/otelconf
// for declarative configuration based on the OpenTelemetry Configuration schema.
//
// # Features
//
//   - Full implementation of telemetry.Provider, telemetry.Tracer, telemetry.Span, and telemetry.Meter interfaces
//   - OpenTelemetry SDK integration via otelconf declarative configuration
//   - Configuration loading from github.com/grinps/go-utils/config using Unmarshal
//   - Embedded types (Tracer wraps trace.Tracer, Meter wraps metric.Meter)
//   - Counter (monotonic/up-down), Gauge, and Histogram instruments
//   - OTLP export support (gRPC and HTTP) via otelconf
//
// # Basic Usage
//
//	provider, err := otel.NewProvider(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer provider.Shutdown(ctx)
//
//	tracer, _ := provider.Tracer("my-component")
//	ctx, span := tracer.Start(ctx, "operation")
//	span.SetAttributes(attribute.String("user.id", "12345"))
//	defer span.End()
//
// # Metrics
//
//	meter, _ := provider.Meter("my-component")
//	inst, _ := meter.NewInstrument("requests_total",
//	    telemetry.InstrumentTypeCounter,
//	    telemetry.CounterTypeMonotonic,
//	)
//	counter := inst.(telemetry.Counter[int64])
//	counter.Add(ctx, 1, attribute.String("method", "GET"))
//
// # Configuration via otelconf
//
// The provider uses go.opentelemetry.io/contrib/otelconf.OpenTelemetryConfiguration.
// Configuration is loaded from config.Config, converted to YAML, and parsed using
// otelconf.ParseYAML to ensure proper handling of complex types like ResourceJSON.
//
//	// Using config package with otelconf schema
//	cfg := ext.NewConfigWrapper(config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
//	    "opentelemetry": map[string]any{
//	        "file_format": "0.3",
//	        "resource": map[string]any{
//	            "schema_url":      "https://opentelemetry.io/schemas/1.39.0",
//	            "attributes_list": "service.name=my-service,service.namespace=production,service.version=1.0.0",
//	        },
//	        "tracer_provider": map[string]any{
//	            "processors": []any{
//	                map[string]any{
//	                    "batch": map[string]any{
//	                        "exporter": map[string]any{
//	                            "otlp_grpc": map[string]any{
//	                                "endpoint": "localhost:4317",
//	                                "insecure": true,
//	                            },
//	                        },
//	                    },
//	                },
//	            },
//	        },
//	    },
//	})))
//	provider, _ := otel.NewProviderFromConfig(ctx, cfg)
//
// # OTLP gRPC Configuration
//
// For OTLP gRPC export, use the "otlp_grpc" exporter key:
//
//	"exporter": map[string]any{
//	    "otlp_grpc": map[string]any{
//	        "endpoint": "localhost:4317",
//	        "insecure": true,  // For local development
//	    },
//	}
//
// Resource attributes use "attributes_list" with comma-separated key=value pairs:
//
//	"resource": map[string]any{
//	    "attributes_list": "service.name=my-service,service.version=1.0.0",
//	}
//
// # Integration Testing
//
// Run integration tests with a local OpenTelemetry Collector:
//
//	docker run -p 127.0.0.1:4317:4317 -p 127.0.0.1:4318:4318 \
//	  --name otel-collector otel/opentelemetry-collector:0.141.0 \
//	  --config /etc/otelcol/config.yaml \
//	  --config 'yaml:service::pipelines::metrics::receivers: [otlp]'
//
//	go test -tags=integration ./...
//
// # Direct OTEL Dependency Usage
//
// Pass OTEL options directly to reduce wrapping overhead:
//
//	// Tracer with version
//	tracer, _ := provider.Tracer("my-service",
//	    trace.WithInstrumentationVersion("1.0.0"),
//	)
//
//	// Span attributes
//	span.SetAttributes(attribute.String("user.id", "12345"))
//
// # Embedded Types
//
// The Tracer, Span, and Meter types embed their OpenTelemetry counterparts:
//
//   - type Tracer struct { trace.Tracer; provider *Provider } - wraps trace.Tracer
//   - type Span struct { trace.Span; provider *Provider } - wraps trace.Span
//   - type Meter struct { metric.Meter } - wraps metric.Meter (lightweight, no provider)
//
// This allows passing OpenTelemetry options directly while maintaining
// the telemetry package interface compatibility.
//
// # Error Handling
//
// The package uses github.com/grinps/go-utils/errext for structured error handling.
// Error codes are defined with type classification for easy error matching:
//
//   - ErrCodeProviderCreation - Provider creation failed
//   - ErrCodeConfigInvalid - Invalid configuration
//   - ErrCodeShutdown - Shutdown failed
//   - ErrCodeAlreadyShutdown - Provider already shutdown
//
// # Environment Variables
//
// The following environment variables are recognized by OpenTelemetry SDK:
//
//   - OTEL_SERVICE_NAME: Service name for resource
//   - OTEL_EXPORTER_OTLP_ENDPOINT: OTLP endpoint URL
//   - OTEL_EXPORTER_OTLP_PROTOCOL: Protocol (grpc or http/protobuf)
//   - OTEL_EXPORTER_OTLP_HEADERS: Headers for OTLP requests
//
// # Thread Safety
//
// All types in this package are safe for concurrent use.
package otel
