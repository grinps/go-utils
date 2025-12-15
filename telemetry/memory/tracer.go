package memory

import (
	"context"
	"sync"
	"time"

	"github.com/grinps/go-utils/telemetry"
)

// Tracer is an in-memory implementation of telemetry.Tracer.
type Tracer struct {
	provider   *Provider
	name       string
	version    string
	schemaURL  string
	attributes []Attribute
}

// Start creates and starts a new span.
func (t *Tracer) Start(ctx context.Context, name string, opts ...any) (context.Context, telemetry.Span) {
	cfg := parseSpanOptions(opts...)

	// Generate IDs
	spanID := generateSpanID()
	var traceID TraceID
	var parentSpanID SpanID

	// Check for parent span in context
	parentSpan := SpanFromContext(ctx)
	if parentSpan != nil && !cfg.newRoot {
		if memSpan, ok := parentSpan.(*Span); ok {
			traceID = memSpan.spanContext.TraceID()
			parentSpanID = memSpan.spanContext.SpanID()
		}
	}
	if !traceID.IsValid() {
		traceID = generateTraceID()
	}

	spanCtx := NewSpanContext(traceID, spanID, TraceFlagsSampled, false)

	span := &Span{
		tracer:       t,
		name:         name,
		spanContext:  spanCtx,
		parentSpanID: parentSpanID,
		kind:         cfg.kind,
		startTime:    time.Now(),
		attributes:   make([]Attribute, 0, len(cfg.attributes)),
		events:       make([]Event, 0),
		links:        cfg.links,
		recording:    true,
		status:       Status{Code: StatusUnset},
	}

	// Add initial attributes
	span.attributes = append(span.attributes, cfg.attributes...)

	ctx = ContextWithSpan(ctx, span)
	return ctx, span
}

// Span is an in-memory implementation of telemetry.Span.
type Span struct {
	mu           sync.RWMutex
	tracer       *Tracer
	name         string
	spanContext  SpanContext
	parentSpanID SpanID
	kind         SpanKind
	startTime    time.Time
	endTime      time.Time
	attributes   []Attribute
	events       []Event
	links        []Link
	status       Status
	recording    bool
	ended        bool
}

// End completes the span.
func (s *Span) End(opts ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ended {
		return
	}

	cfg := parseSpanEndOptions(opts...)
	endTime := cfg.endTime
	if endTime.IsZero() {
		endTime = time.Now()
	}
	s.endTime = endTime
	s.ended = true
	s.recording = false

	// Record the span
	recorded := &RecordedSpan{
		Name:         s.name,
		SpanContext:  s.spanContext,
		ParentSpanID: s.parentSpanID,
		Kind:         s.kind,
		StartTime:    s.startTime,
		EndTime:      s.endTime,
		Attributes:   make([]Attribute, len(s.attributes)),
		Events:       make([]Event, len(s.events)),
		Links:        make([]Link, len(s.links)),
		Status:       s.status,
		TracerName:   s.tracer.name,
	}
	copy(recorded.Attributes, s.attributes)
	copy(recorded.Events, s.events)
	copy(recorded.Links, s.links)

	s.tracer.provider.recordSpan(recorded)
}

// IsRecording returns true if the span is recording.
func (s *Span) IsRecording() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.recording
}

// SetAttributes sets attributes on the span.
func (s *Span) SetAttributes(attrs ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.recording {
		return
	}

	for _, attr := range attrs {
		if a, ok := attr.(Attribute); ok {
			s.attributes = append(s.attributes, a)
		}
	}
}

// AddEvent adds an event to the span.
func (s *Span) AddEvent(name string, opts ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.recording {
		return
	}

	event := NewEvent(name, opts...)
	s.events = append(s.events, event)
}

// RecordError records an error as an exception event.
func (s *Span) RecordError(err error, opts ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.recording || err == nil {
		return
	}

	// Add exception attributes
	attrs := []Attribute{
		String("exception.type", typeStr(err)),
		String("exception.message", err.Error()),
	}

	cfg := parseEventOptions(opts...)
	attrs = append(attrs, cfg.attributes...)

	event := Event{
		Name:       "exception",
		Timestamp:  cfg.timestamp,
		Attributes: attrs,
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	s.events = append(s.events, event)
}

// SetStatus sets the status of the span.
func (s *Span) SetStatus(code int, description string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.recording {
		return
	}

	statusCode := StatusCode(code)
	// Only update if new status is higher priority
	// Error > OK > Unset
	if statusCode > s.status.Code {
		s.status = NewStatus(statusCode, description)
	}
}

// SetName sets the name of the span.
func (s *Span) SetName(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.recording {
		return
	}

	s.name = name
}

// TracerProvider returns the TracerProvider that created this span's Tracer.
func (s *Span) TracerProvider() telemetry.Provider {
	return s.tracer.provider
}

// typeStr returns the type name of an error.
func typeStr(err error) string {
	if err == nil {
		return ""
	}
	return "error"
}
