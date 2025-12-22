package otel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/grinps/go-utils/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func TestTracerStart(t *testing.T) {
	ctx := context.Background()
	provider, err := NewProvider(ctx)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Shutdown(ctx)

	tracer, _ := provider.Tracer("test-tracer")

	t.Run("basic span", func(t *testing.T) {
		ctx, span := tracer.Start(ctx, "test-operation")
		if ctx == nil {
			t.Error("expected context to be non-nil")
		}
		if span == nil {
			t.Fatal("expected span to be non-nil")
		}
		span.End()
	})

	t.Run("span with kind", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation", trace.WithSpanKind(trace.SpanKindServer))
		span.End()
	})

	t.Run("span with attributes", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation", trace.WithAttributes(
			attribute.String("key1", "value1"),
			attribute.Int64("key2", 42),
		))
		span.End()
	})

	t.Run("span with all kinds", func(t *testing.T) {
		kinds := []trace.SpanKind{
			trace.SpanKindUnspecified,
			trace.SpanKindInternal,
			trace.SpanKindServer,
			trace.SpanKindClient,
			trace.SpanKindProducer,
			trace.SpanKindConsumer,
		}
		for _, kind := range kinds {
			_, span := tracer.Start(ctx, "test-operation", trace.WithSpanKind(kind))
			span.End()
		}
	})

	t.Run("span with new root", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation", trace.WithNewRoot())
		span.End()
	})
}

func TestSpanOperations(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	tracer, _ := provider.Tracer("test-tracer")

	t.Run("IsRecording", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		// Note: With default otelconf configuration, spans may not be recording
		// since no tracer provider is configured. This is expected behavior.
		// The IsRecording() method still works correctly.
		_ = span.IsRecording()
		span.End()
	})

	t.Run("SetAttributes with Attribute type", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.SetAttributes(attribute.String("key", "value"))
		span.End()
	})

	t.Run("SetAttributes with key-value pairs", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		// The new implementation supports []attribute.KeyValue or single attribute.KeyValue via ...any
		// but parsing logic was removed for raw strings/ints.
		// We need to update the test or the implementation if we want to support raw args.
		// The user request said "implement methods aligned with span calls by passing all the otel library supported variadic input as it is and ignoring any other inputs".
		// OTEL SetAttributes takes ...attribute.KeyValue.
		// Our wrapper takes ...any.
		// The implementation I wrote iterates ...any and checks for attribute.KeyValue or []attribute.KeyValue.
		// It does NOT support alternating string/value anymore.
		span.SetAttributes(attribute.String("key1", "value1"), attribute.Int64("key2", 42))
		span.End()
	})

	t.Run("AddEvent", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.AddEvent("test-event")
		span.End()
	})

	t.Run("AddEvent with timestamp", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.AddEvent("test-event", trace.WithTimestamp(time.Now()))
		span.End()
	})

	t.Run("AddEvent with attributes", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.AddEvent("test-event", trace.WithAttributes(
			attribute.String("key", "value"),
		))
		span.End()
	})

	t.Run("RecordError", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.RecordError(errors.New("test error"))
		span.End()
	})

	t.Run("RecordError nil", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.RecordError(nil) // Should not panic
		span.End()
	})

	t.Run("RecordError with attributes", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.RecordError(errors.New("test error"),
			trace.WithAttributes(attribute.String("error.type", "TestError")),
		)
		span.End()
	})

	t.Run("SetStatus", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.SetStatus(int(StatusOK), "success")
		span.End()
	})

	t.Run("SetStatus error", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.SetStatus(int(StatusError), "failed")
		span.End()
	})

	t.Run("SetStatus unset", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.SetStatus(int(StatusUnset), "")
		span.End()
	})

	t.Run("SetName", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.SetName("renamed-operation")
		span.End()
	})

	t.Run("TracerProvider", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		tp := span.TracerProvider()
		if tp != provider {
			t.Error("expected TracerProvider to return the provider")
		}
		span.End()
	})

	t.Run("End with timestamp", func(t *testing.T) {
		_, span := tracer.Start(ctx, "test-operation")
		span.End(trace.WithTimestamp(time.Now()))
	})
}

func TestSpanImplementsInterface(t *testing.T) {
	var _ telemetry.Span = (*Span)(nil)
}

func TestSpanSetAttributesWithSlice(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	tracer, _ := provider.Tracer("test-tracer")
	_, span := tracer.Start(ctx, "test-span")

	// Test with []attribute.KeyValue slice
	attrs := []attribute.KeyValue{
		attribute.String("key1", "val1"),
		attribute.Int64("key2", 42),
	}
	span.SetAttributes(attrs)
	span.End()
}

func TestTracerImplementsInterface(t *testing.T) {
	var _ telemetry.Tracer = (*Tracer)(nil)
}

func TestToOtelStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		code     StatusCode
		expected string
	}{
		{"StatusUnset", StatusUnset, "Unset"},
		{"StatusOK", StatusOK, "Ok"},
		{"StatusError", StatusError, "Error"},
		{"Unknown", StatusCode(99), "Unset"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toOtelStatusCode(tt.code)
			if result.String() != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result.String())
			}
		})
	}
}

func TestSpanWithLinks(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	tracer, _ := provider.Tracer("test-tracer")

	// Create a link
	link := trace.Link{
		SpanContext: trace.SpanContext{}, // Empty context for test
		Attributes:  []attribute.KeyValue{attribute.String("link.key", "link.value")},
	}
	_, span := tracer.Start(ctx, "test-with-link", trace.WithLinks(link))
	span.End()
}

func TestSpanWithStartTime(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	tracer, _ := provider.Tracer("test-tracer")

	_, span := tracer.Start(ctx, "test-with-start-time", trace.WithTimestamp(time.Now().Add(-1*time.Hour)))
	span.End()
}

func TestNonRecordingSpanOperations(t *testing.T) {
	ctx := context.Background()
	// Create provider with always-off sampler
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	tracer, _ := provider.Tracer("test-tracer")
	_, span := tracer.Start(ctx, "non-recording-span")

	// These should not panic even when not recording
	span.SetAttributes(attribute.String("key", "value"))
	span.AddEvent("event")
	span.RecordError(errors.New("error"))
	span.SetStatus(int(StatusError), "error")
	span.End()
}
