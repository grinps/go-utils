package telemetry

// Span represents a unit of work or operation.
// It provides methods to set attributes, record events, and manage the span lifecycle.
//
// Spans should be ended by calling End() when the operation completes.
// Once a span has ended, it should not be modified.
//
// Implementations must be safe for concurrent use.
type Span interface {
	// End completes the span. After End is called, the span should not be modified.
	// End should be called exactly once for each span.
	End(opts ...any)

	// IsRecording returns true if the span is recording events.
	// If false, SetAttributes, AddEvent, RecordError, and SetStatus have no effect.
	IsRecording() bool

	// SetAttributes sets attributes on the span.
	// If the span is not recording, this is a no-op.
	SetAttributes(attrs ...any)

	// AddEvent adds an event to the span.
	// If the span is not recording, this is a no-op.
	AddEvent(name string, opts ...any)

	// RecordError records an error as an exception event.
	// If the span is not recording, this is a no-op.
	// The error is recorded as an event with standard exception attributes.
	RecordError(err error, opts ...any)

	// SetStatus sets the status of the span.
	// If the span is not recording, this is a no-op.
	SetStatus(code int, description string)

	// SetName sets the name of the span.
	// This should only be used in situations where the span name
	// needs to be determined after the span has started.
	SetName(name string)

	// TracerProvider returns the TracerProvider that created this span's Tracer.
	// This can return nil if the provider is not available.
	TracerProvider() Provider
}
