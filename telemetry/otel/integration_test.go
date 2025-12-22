//go:build integration

package otel

import (
	"context"
	"testing"
	"time"

	"github.com/grinps/go-utils/config"
	"github.com/grinps/go-utils/config/ext"
	"github.com/grinps/go-utils/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// These tests require a running OpenTelemetry Collector.
// Run with: go test -tags=integration ./...
//
// Start the collector with:
//   docker run \
//   -p 127.0.0.1:4317:4317 \
//   -p 127.0.0.1:4318:4318 \
//   -p 127.0.0.1:55679:55679 \
//   --name "otel-collector" otel/opentelemetry-collector:0.141.0 ' \
//   --config' '/etc/otelcol/config.yaml' '--config' 'yaml:service::pipelines::metrics::receivers: [otlp]'

const (
	// Default OTLP gRPC endpoint
	testOTLPEndpoint = "localhost:4317"
)

// Create config with OTLP exporter settings
var cfg = ext.NewConfigWrapper(config.NewSimpleConfig(context.Background(), config.WithConfigurationMap(map[string]any{
	"opentelemetry": map[string]any{
		"file_format": "0.3",
		"resource": map[string]any{
			"attributes": []any{
				map[string]any{
					"name":  "service.name",
					"value": "TestIntegration_ErrorRecording",
				},
				map[string]any{
					"name":  "service.namespace",
					"value": "integration_test",
				},
				map[string]any{
					"name":  "service.version",
					"value": "1.0.0",
				},
			},
			"schema_url":      "https://opentelemetry.io/schemas/1.39.0",
			"attributes_list": "service.name=TestIntegration_ErrorRecording,service.namespace=integration_test,service.version=1.0.0",
		},
		"tracer_provider": map[string]any{
			"processors": []any{
				map[string]any{
					"batch": map[string]any{
						"exporter": map[string]any{
							"otlp_grpc": map[string]any{
								"endpoint": testOTLPEndpoint,
								"insecure": true,
							},
						},
					},
				},
			},
		},
	},
})))

func TestIntegration_TracingWithOTLPExport(t *testing.T) {
	ctx := context.Background()

	provider, err := NewProviderFromConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := provider.Shutdown(shutdownCtx); err != nil {
			t.Errorf("shutdown error: %v", err)
		}
	}()

	tracer, err := provider.Tracer("integration-tracer")
	if err != nil {
		t.Fatalf("failed to create tracer: %v", err)
	}

	// Create a parent span
	ctx, parentSpan := tracer.Start(ctx, "parent-operation")
	parentSpan.SetAttributes(attribute.String("test.type", "integration"))
	parentSpan.SetAttributes(attribute.Int64("test.run_id", time.Now().UnixNano()))

	// Create a child span
	_, childSpan := tracer.Start(ctx, "child-operation")
	childSpan.SetAttributes(attribute.Int("child.index", 1))
	childSpan.AddEvent("processing-started")

	// Simulate some work
	time.Sleep(10 * time.Millisecond)

	childSpan.AddEvent("processing-completed", attribute.Int("items_processed", 100))
	childSpan.SetStatus(int(codes.Ok), "success")
	childSpan.End()

	parentSpan.SetStatus(int(codes.Ok), "completed")
	parentSpan.End()

	t.Log("Traces exported to OTLP collector successfully")
}

func TestIntegration_MetricsWithOTLPExport(t *testing.T) {
	ctx := context.Background()

	provider, err := NewProviderFromConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := provider.Shutdown(shutdownCtx); err != nil {
			t.Errorf("shutdown error: %v", err)
		}
	}()

	meter, err := provider.Meter("integration-meter")
	if err != nil {
		t.Fatalf("failed to create meter: %v", err)
	}

	// Create a counter
	counterInst, err := meter.NewInstrument("integration_requests_total",
		telemetry.InstrumentTypeCounter,
		telemetry.CounterTypeMonotonic,
		"Total integration test requests",
		"1",
	)
	if err != nil {
		t.Fatalf("failed to create counter: %v", err)
	}
	counter := counterInst.(telemetry.Counter[int64])

	// Create a histogram
	histogramInst, err := meter.NewInstrument("integration_request_duration",
		telemetry.InstrumentTypeRecorder,
		telemetry.AggregationStrategyHistogram,
		"Request duration in milliseconds",
		"ms",
	)
	if err != nil {
		t.Fatalf("failed to create histogram: %v", err)
	}
	histogram := histogramInst.(telemetry.Recorder[float64])

	// Record some metrics
	for i := 0; i < 10; i++ {
		counter.Add(ctx, 1, attribute.String("method", "GET"), attribute.String("path", "/api/test"))
		histogram.Record(ctx, float64(10+i*5), attribute.String("method", "GET"))
		time.Sleep(5 * time.Millisecond)
	}

	t.Log("Metrics exported to OTLP collector successfully")
}

func TestIntegration_FullTelemetryFlow(t *testing.T) {
	ctx := context.Background()

	provider, err := NewProviderFromConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := provider.Shutdown(shutdownCtx); err != nil {
			t.Errorf("shutdown error: %v", err)
		}
	}()

	// Setup tracer and meter
	tracer, _ := provider.Tracer("full-test-tracer", "version", "1.0.0")
	meter, _ := provider.Meter("full-test-meter", "version", "1.0.0")

	// Create instruments
	requestCounter, _ := meter.NewInstrument("http_requests_total",
		telemetry.InstrumentTypeCounter,
		telemetry.CounterTypeMonotonic,
	)
	counter := requestCounter.(telemetry.Counter[int64])

	activeRequests, _ := meter.NewInstrument("http_requests_active",
		telemetry.InstrumentTypeCounter,
		telemetry.CounterTypeUpDown,
	)
	activeCounter := activeRequests.(telemetry.Counter[int64])

	requestDuration, _ := meter.NewInstrument("http_request_duration_ms",
		telemetry.InstrumentTypeRecorder,
		telemetry.AggregationStrategyHistogram,
	)
	durationHistogram := requestDuration.(telemetry.Recorder[float64])

	// Simulate HTTP request handling
	simulateRequest := func(method, path string, duration time.Duration) {
		ctx, span := tracer.Start(ctx, "http.request", trace.WithSpanKind(trace.SpanKindServer))
		span.SetAttributes(
			attribute.String("http.method", method),
			attribute.String("http.path", path),
			attribute.String("http.host", "localhost"),
		)

		activeCounter.Add(ctx, 1, attribute.String("method", method))
		defer activeCounter.Add(ctx, -1, attribute.String("method", method))

		span.AddEvent("request-received")

		// Simulate processing
		time.Sleep(duration)

		span.AddEvent("request-processed")
		span.SetStatus(int(codes.Ok), "200 OK")
		span.End()

		counter.Add(ctx, 1, attribute.String("method", method), attribute.String("path", path), attribute.String("status", "200"))
		durationHistogram.Record(ctx, float64(duration.Milliseconds()), attribute.String("method", method))
	}

	// Simulate multiple requests
	simulateRequest("GET", "/api/users", 15*time.Millisecond)
	simulateRequest("POST", "/api/users", 25*time.Millisecond)
	simulateRequest("GET", "/api/products", 10*time.Millisecond)
	simulateRequest("PUT", "/api/users/123", 20*time.Millisecond)
	simulateRequest("DELETE", "/api/users/456", 5*time.Millisecond)

	t.Log("Full telemetry flow completed - traces and metrics exported to OTLP collector")
}

func TestIntegration_ErrorRecording(t *testing.T) {
	ctx := context.Background()

	provider, err := NewProviderFromConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("failed to create provider: %+v", err)
	}
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		provider.Shutdown(shutdownCtx)
	}()

	tracer, _ := provider.Tracer("error-tracer")

	// Create span with error
	_, span := tracer.Start(ctx, "failing-operation")
	span.SetAttributes(attribute.String("operation.type", "database-query"))

	// Simulate an error
	simulatedErr := &testError{message: "connection timeout", code: "DB_TIMEOUT"}
	span.RecordError(simulatedErr, trace.WithAttributes(attribute.String("error.code", simulatedErr.code)))
	span.SetStatus(int(codes.Error), "database connection failed")
	span.End()

	t.Log("Error span exported to OTLP collector successfully")
}

type testError struct {
	message string
	code    string
}

func (e *testError) Error() string {
	return e.message
}
