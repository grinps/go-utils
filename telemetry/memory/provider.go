package memory

import (
	"context"
	"crypto/rand"
	"sync"
	"time"

	"github.com/grinps/go-utils/errext"
	"github.com/grinps/go-utils/telemetry"
)

// Compile-time check that Provider implements telemetry.Provider.
var _ telemetry.Provider = (*Provider)(nil)

// ErrMemoryProviderOperation is returned when operations are attempted on a shutdown provider.
var ErrMemoryProviderOperation = errext.NewErrorCodeOfType(51, telemetry.ErrTypePrefix)

// ErrReasonMemoryProviderShutdown is the reason for when operations are attempted on a shutdown provider.
const ErrReasonMemoryProviderShutdown = "provider has been shutdown"

// Provider is an in-memory implementation of telemetry.Provider.
// It stores all telemetry data in memory for testing and development.
type Provider struct {
	mu       sync.RWMutex
	tracers  map[string]*Tracer
	meters   map[string]*Meter
	spans    []*RecordedSpan
	shutdown bool
}

// ProviderOption is a functional option for configuring the Provider.
type ProviderOption func(*Provider)

// NewProvider creates a new in-memory Provider.
func NewProvider(opts ...ProviderOption) *Provider {
	p := &Provider{
		tracers: make(map[string]*Tracer),
		meters:  make(map[string]*Meter),
		spans:   make([]*RecordedSpan, 0),
	}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Tracer returns a Tracer for creating spans.
// If a Tracer with the same name already exists, it is returned.
// If the provider has been shutdown, it returns a tracer from the NoopProvider.
func (p *Provider) Tracer(name string, opts ...any) (telemetry.Tracer, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shutdown {
		return (&telemetry.NoopProvider{}).Tracer(name, opts...)
	}

	if t, ok := p.tracers[name]; ok {
		return t, nil
	}

	cfg := parseTracerOptions(opts...)

	t := &Tracer{
		provider:   p,
		name:       name,
		version:    cfg.version,
		schemaURL:  cfg.schemaURL,
		attributes: cfg.attributes,
	}
	p.tracers[name] = t
	return t, nil
}

// Meter returns a Meter for creating metric instruments.
// If a Meter with the same name already exists, it is returned.
// If the provider has been shutdown, it returns a meter from the NoopProvider.
func (p *Provider) Meter(name string, opts ...any) (telemetry.Meter, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shutdown {
		return (&telemetry.NoopProvider{}).Meter(name, opts...)
	}

	if m, ok := p.meters[name]; ok {
		return m, nil
	}

	cfg := parseMeterOptions(opts...)

	m := &Meter{
		provider:   p,
		name:       name,
		version:    cfg.version,
		schemaURL:  cfg.schemaURL,
		attributes: cfg.attributes,
	}
	p.meters[name] = m
	return m, nil
}

// Shutdown shuts down the provider.
// After Shutdown is called, all subsequent operations return no-op implementations.
func (p *Provider) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.shutdown {
		return ErrMemoryProviderOperation.New(ErrReasonMemoryProviderShutdown)
	}

	p.shutdown = true
	return nil
}

// RecordedSpans returns all recorded spans.
// This is useful for testing to verify span creation and attributes.
func (p *Provider) RecordedSpans() []*RecordedSpan {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*RecordedSpan, len(p.spans))
	copy(result, p.spans)
	return result
}

// RecordedSpansByName returns all recorded spans with the given name.
func (p *Provider) RecordedSpansByName(name string) []*RecordedSpan {
	p.mu.RLock()
	defer p.mu.RUnlock()

	var result []*RecordedSpan
	for _, s := range p.spans {
		if s.Name == name {
			result = append(result, s)
		}
	}
	return result
}

// Reset clears all recorded data.
// This is useful for resetting state between tests.
func (p *Provider) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.spans = make([]*RecordedSpan, 0)
	for _, m := range p.meters {
		m.reset()
	}
}

// IsShutdown returns true if the provider has been shutdown.
func (p *Provider) IsShutdown() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.shutdown
}

// recordSpan adds a span to the recorded spans.
func (p *Provider) recordSpan(span *RecordedSpan) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.spans = append(p.spans, span)
}

// RecordedSpan represents a span that was recorded by the in-memory provider.
type RecordedSpan struct {
	// Name is the span name.
	Name string
	// SpanContext is the span's context.
	SpanContext SpanContext
	// ParentSpanID is the parent span's ID, or zero if this is a root span.
	ParentSpanID SpanID
	// Kind is the span kind.
	Kind SpanKind
	// StartTime is when the span started.
	StartTime time.Time
	// EndTime is when the span ended.
	EndTime time.Time
	// Attributes are the span's attributes.
	Attributes []Attribute
	// Events are the span's events.
	Events []Event
	// Links are the span's links.
	Links []Link
	// Status is the span's status.
	Status Status
	// TracerName is the name of the tracer that created this span.
	TracerName string
}

// HasAttribute returns true if the span has an attribute with the given key.
func (s *RecordedSpan) HasAttribute(key string) bool {
	for _, attr := range s.Attributes {
		if attr.Key == key {
			return true
		}
	}
	return false
}

// GetAttribute returns the value of the attribute with the given key, or nil if not found.
func (s *RecordedSpan) GetAttribute(key string) any {
	for _, attr := range s.Attributes {
		if attr.Key == key {
			return attr.Value
		}
	}
	return nil
}

// HasEvent returns true if the span has an event with the given name.
func (s *RecordedSpan) HasEvent(name string) bool {
	for _, event := range s.Events {
		if event.Name == name {
			return true
		}
	}
	return false
}

// Duration returns the duration of the span.
// Returns zero if the span has not ended.
func (s *RecordedSpan) Duration() time.Duration {
	if s.EndTime.IsZero() {
		return 0
	}
	return s.EndTime.Sub(s.StartTime)
}

// generateTraceID generates a random trace ID.
func generateTraceID() TraceID {
	var id TraceID
	_, _ = rand.Read(id[:])
	return id
}

// generateSpanID generates a random span ID.
func generateSpanID() SpanID {
	var id SpanID
	_, _ = rand.Read(id[:])
	return id
}
