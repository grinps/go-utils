package otel

import (
	"context"

	"github.com/grinps/go-utils/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps an OpenTelemetry tracer using embedding.
type Tracer struct {
	trace.Tracer
	provider *Provider
}

// Start creates and starts a new span.
// Filters opts for trace.SpanStartOption and ignores non-matching parameters.
func (t *Tracer) Start(ctx context.Context, name string, opts ...any) (context.Context, telemetry.Span) {
	var otelOpts []trace.SpanStartOption
	for _, opt := range opts {
		if o, ok := opt.(trace.SpanStartOption); ok {
			otelOpts = append(otelOpts, o)
		}
	}

	ctx, otelSpan := t.Tracer.Start(ctx, name, otelOpts...)

	return ctx, &Span{
		Span:     otelSpan,
		provider: t.provider,
	}
}

// Span wraps an OpenTelemetry span using embedding.
type Span struct {
	trace.Span
	provider *Provider
}

// End completes the span.
// Filters opts for trace.SpanEndOption and ignores non-matching parameters.
func (s *Span) End(opts ...any) {
	var otelOpts []trace.SpanEndOption
	for _, opt := range opts {
		if o, ok := opt.(trace.SpanEndOption); ok {
			otelOpts = append(otelOpts, o)
		}
	}
	s.Span.End(otelOpts...)
}

// SetAttributes sets attributes on the span.
// Supports attribute.KeyValue and []attribute.KeyValue, ignores non-matching parameters.
func (s *Span) SetAttributes(attrs ...any) {
	var otelAttrs []attribute.KeyValue
	for _, a := range attrs {
		switch v := a.(type) {
		case attribute.KeyValue:
			otelAttrs = append(otelAttrs, v)
		case []attribute.KeyValue:
			otelAttrs = append(otelAttrs, v...)
		}
	}
	s.Span.SetAttributes(otelAttrs...)
}

// AddEvent adds an event to the span.
// Filters opts for trace.EventOption and ignores non-matching parameters.
func (s *Span) AddEvent(name string, opts ...any) {
	var otelOpts []trace.EventOption
	for _, opt := range opts {
		if o, ok := opt.(trace.EventOption); ok {
			otelOpts = append(otelOpts, o)
		}
	}

	s.Span.AddEvent(name, otelOpts...)
}

// RecordError records an error as an exception event.
// Filters opts for trace.EventOption and ignores non-matching parameters.
func (s *Span) RecordError(err error, opts ...any) {
	var otelOpts []trace.EventOption
	for _, opt := range opts {
		if o, ok := opt.(trace.EventOption); ok {
			otelOpts = append(otelOpts, o)
		}
	}

	s.Span.RecordError(err, otelOpts...)
}

// SetStatus sets the status of the span.
func (s *Span) SetStatus(code int, description string) {
	s.Span.SetStatus(toOtelStatusCode(StatusCode(code)), description)
}

// TracerProvider returns the Provider that created this span's Tracer.
func (s *Span) TracerProvider() telemetry.Provider {
	return s.provider
}

func toOtelStatusCode(code StatusCode) codes.Code {
	switch code {
	case StatusOK:
		return codes.Ok
	case StatusError:
		return codes.Error
	default:
		return codes.Unset
	}
}

// Ensure types implement interfaces.
var (
	_ telemetry.Tracer = (*Tracer)(nil)
	_ telemetry.Span   = (*Span)(nil)
)
