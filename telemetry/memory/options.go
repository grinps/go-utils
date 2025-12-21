package memory

import (
	"context"
	"time"

	"github.com/grinps/go-utils/telemetry"
)

// TracerConfig holds configuration for creating a Tracer.
type TracerConfig struct {
	version    string
	schemaURL  string
	attributes []Attribute
}

// parseTracerOptions extracts tracer configuration from variadic any options.
// It supports key-value pairs using "version" and "schemaURL" as special keys,
// and creates attributes from other key-value pairs:
//
//	tracer, _ := provider.Tracer("my-service", "version", "1.0.0", "service.env", "prod")
//
// Or using TracerConfig directly:
//
//	tracer, _ := provider.Tracer("my-service", &memory.TracerConfig{...})
func parseTracerOptions(opts ...any) *TracerConfig {
	cfg := &TracerConfig{}
	for i := 0; i < len(opts); i++ {
		switch v := opts[i].(type) {
		case string:
			// Check if this is a key followed by a value
			if i+1 < len(opts) {
				next := opts[i+1]
				// Skip if next is a config type
				switch next.(type) {
				case TracerConfig, *TracerConfig, Attribute, []Attribute:
					continue
				}
				// Handle special keys
				switch v {
				case "version":
					if val, ok := next.(string); ok {
						cfg.version = val
						i++
						continue
					}
				case "schemaURL":
					if val, ok := next.(string); ok {
						cfg.schemaURL = val
						i++
						continue
					}
				default:
					// Create attribute from key-value pair
					cfg.attributes = append(cfg.attributes, Attribute{Key: v, Value: next})
					i++
					continue
				}
			}
			// Fallback: first unrecognized string is treated as version
			if cfg.version == "" {
				cfg.version = v
			}
		case TracerConfig:
			cfg = &v
		case *TracerConfig:
			if v != nil {
				cfg = v
			}
		case Attribute:
			cfg.attributes = append(cfg.attributes, v)
		case []Attribute:
			cfg.attributes = append(cfg.attributes, v...)
		}
	}
	return cfg
}

// MeterConfig holds configuration for creating a Meter.
type MeterConfig struct {
	version    string
	schemaURL  string
	attributes []Attribute
}

// parseMeterOptions extracts meter configuration from variadic any options.
// It supports key-value pairs using "version" and "schemaURL" as special keys,
// and creates attributes from other key-value pairs:
//
//	meter, _ := provider.Meter("my-service", "version", "1.0.0", "service.env", "prod")
//
// Or using MeterConfig directly:
//
//	meter, _ := provider.Meter("my-service", &memory.MeterConfig{...})
func parseMeterOptions(opts ...any) *MeterConfig {
	cfg := &MeterConfig{}
	for i := 0; i < len(opts); i++ {
		switch v := opts[i].(type) {
		case string:
			// Check if this is a key followed by a value
			if i+1 < len(opts) {
				next := opts[i+1]
				// Skip if next is a config type
				switch next.(type) {
				case MeterConfig, *MeterConfig, Attribute, []Attribute:
					continue
				}
				// Handle special keys
				switch v {
				case "version":
					if val, ok := next.(string); ok {
						cfg.version = val
						i++
						continue
					}
				case "schemaURL":
					if val, ok := next.(string); ok {
						cfg.schemaURL = val
						i++
						continue
					}
				default:
					// Create attribute from key-value pair
					cfg.attributes = append(cfg.attributes, Attribute{Key: v, Value: next})
					i++
					continue
				}
			}
			// Fallback: first unrecognized string is treated as version
			if cfg.version == "" {
				cfg.version = v
			}
		case MeterConfig:
			cfg = &v
		case *MeterConfig:
			if v != nil {
				cfg = v
			}
		case Attribute:
			cfg.attributes = append(cfg.attributes, v)
		case []Attribute:
			cfg.attributes = append(cfg.attributes, v...)
		}
	}
	return cfg
}

// SpanKind represents the type of span.
type SpanKind int

const (
	SpanKindUnspecified SpanKind = iota
	SpanKindInternal
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

// SpanConfig holds configuration for creating a Span.
type SpanConfig struct {
	kind       SpanKind
	attributes []Attribute
	links      []Link
	newRoot    bool
	startTime  time.Time
}

// parseSpanOptions extracts span configuration from variadic any options.
func parseSpanOptions(opts ...any) *SpanConfig {
	cfg := &SpanConfig{
		kind: SpanKindInternal,
	}
	for _, opt := range opts {
		switch v := opt.(type) {
		case SpanKind:
			cfg.kind = v
		case SpanConfig:
			cfg = &v
		case *SpanConfig:
			if v != nil {
				cfg = v
			}
		case []Attribute:
			cfg.attributes = append(cfg.attributes, v...)
		case Attribute:
			cfg.attributes = append(cfg.attributes, v)
		case []Link:
			cfg.links = append(cfg.links, v...)
		case Link:
			cfg.links = append(cfg.links, v)
		}
	}
	return cfg
}

// SpanEndConfig holds configuration for ending a span.
type SpanEndConfig struct {
	endTime time.Time
}

// parseSpanEndOptions extracts span end configuration from variadic any options.
func parseSpanEndOptions(opts ...any) *SpanEndConfig {
	cfg := &SpanEndConfig{}
	for _, opt := range opts {
		switch v := opt.(type) {
		case time.Time:
			cfg.endTime = v
		case SpanEndConfig:
			cfg = &v
		case *SpanEndConfig:
			if v != nil {
				cfg = v
			}
		}
	}
	return cfg
}

// EventConfig holds configuration for an event.
type EventConfig struct {
	timestamp  time.Time
	attributes []Attribute
}

// parseEventOptions extracts event configuration from variadic any options.
// The first time.Time value is assigned to timestamp. It also supports key-value
// pairs where a string key is followed by any value to create attributes:
//
//	span.AddEvent("my-event", time.Now(), "user.id", "12345", "count", 10)
//
// Or using Attribute types:
//
//	span.AddEvent("my-event", memory.String("user.id", "12345"))
func parseEventOptions(opts ...any) *EventConfig {
	cfg := &EventConfig{}
	for i := 0; i < len(opts); i++ {
		switch v := opts[i].(type) {
		case time.Time:
			// First time.Time is timestamp
			if cfg.timestamp.IsZero() {
				cfg.timestamp = v
			}
		case EventConfig:
			cfg = &v
		case *EventConfig:
			if v != nil {
				cfg = v
			}
		case []Attribute:
			cfg.attributes = append(cfg.attributes, v...)
		case Attribute:
			cfg.attributes = append(cfg.attributes, v)
		case string:
			// String followed by any value creates a key-value attribute
			if i+1 < len(opts) {
				// Check next value is not an Attribute or EventConfig (to avoid consuming it)
				next := opts[i+1]
				switch next.(type) {
				case Attribute, EventConfig, *EventConfig, []Attribute:
					// Don't consume, let next iteration handle it
				default:
					cfg.attributes = append(cfg.attributes, Attribute{Key: v, Value: next})
					i++ // Skip the value we just consumed
				}
			}
		}
	}
	return cfg
}

// InstrumentConfig holds configuration for creating an instrument.
type InstrumentConfig struct {
	description         string
	unit                string
	buckets             []float64
	instrumentType      telemetry.InstrumentType
	counterType         telemetry.CounterType
	aggregationStrategy telemetry.AggregationStrategy
	precision           telemetry.Precision
}

// parseInstrumentOptions extracts instrument configuration from variadic any options.
func parseInstrumentOptions(opts ...any) *InstrumentConfig {
	cfg := &InstrumentConfig{}
	for _, opt := range opts {
		switch v := opt.(type) {
		case string:
			// First string is treated as description
			if cfg.description == "" {
				cfg.description = v
			} else if cfg.unit == "" {
				cfg.unit = v
			}
		case InstrumentConfig:
			cfg = &v
		case *InstrumentConfig:
			if v != nil {
				cfg = v
			}
		case []float64:
			cfg.buckets = v
		case telemetry.InstrumentType:
			cfg.instrumentType = v
		case telemetry.CounterType:
			cfg.counterType = v
		case telemetry.AggregationStrategy:
			cfg.aggregationStrategy = v
		case telemetry.Precision:
			cfg.precision = v
		}
	}
	return cfg
}

// Attribute represents a key-value pair.
type Attribute struct {
	Key   string
	Value any
}

// String creates a string attribute.
func String(key, value string) Attribute {
	return Attribute{Key: key, Value: value}
}

// Int64 creates an int64 attribute.
func Int64(key string, value int64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Float64 creates a float64 attribute.
func Float64(key string, value float64) Attribute {
	return Attribute{Key: key, Value: value}
}

// Bool creates a bool attribute.
func Bool(key string, value bool) Attribute {
	return Attribute{Key: key, Value: value}
}

// Link represents a link to another span.
type Link struct {
	SpanContext SpanContext
	Attributes  []Attribute
}

// SpanContext contains identifying trace information about a span.
type SpanContext struct {
	traceID    TraceID
	spanID     SpanID
	traceFlags TraceFlags
	remote     bool
}

// TraceID returns the trace ID.
func (sc SpanContext) TraceID() TraceID {
	return sc.traceID
}

// SpanID returns the span ID.
func (sc SpanContext) SpanID() SpanID {
	return sc.spanID
}

// TraceFlags returns the trace flags.
func (sc SpanContext) TraceFlags() TraceFlags {
	return sc.traceFlags
}

// IsRemote returns true if the span context is from a remote parent.
func (sc SpanContext) IsRemote() bool {
	return sc.remote
}

// IsValid returns true if the span context is valid.
func (sc SpanContext) IsValid() bool {
	return sc.traceID.IsValid() && sc.spanID.IsValid()
}

// NewSpanContext creates a new SpanContext.
func NewSpanContext(traceID TraceID, spanID SpanID, flags TraceFlags, remote bool) SpanContext {
	return SpanContext{
		traceID:    traceID,
		spanID:     spanID,
		traceFlags: flags,
		remote:     remote,
	}
}

// TraceID is a unique identifier for a trace.
type TraceID [16]byte

// IsValid returns true if the TraceID is valid (non-zero).
func (t TraceID) IsValid() bool {
	return t != TraceID{}
}

// SpanID is a unique identifier for a span within a trace.
type SpanID [8]byte

// IsValid returns true if the SpanID is valid (non-zero).
func (s SpanID) IsValid() bool {
	return s != SpanID{}
}

// TraceFlags contains trace flags.
type TraceFlags byte

const (
	TraceFlagsNone    TraceFlags = 0
	TraceFlagsSampled TraceFlags = 1 << 0
)

// IsSampled returns true if the sampled flag is set.
func (f TraceFlags) IsSampled() bool {
	return f&TraceFlagsSampled == TraceFlagsSampled
}

// StatusCode represents the status of a span.
type StatusCode int

const (
	StatusUnset StatusCode = iota
	StatusOK
	StatusError
)

// Status represents the status of a span.
type Status struct {
	Code        StatusCode
	Description string
}

// NewStatus creates a new Status.
func NewStatus(code StatusCode, description string) Status {
	return Status{Code: code, Description: description}
}

// Event represents an event that occurred during a span's lifetime.
type Event struct {
	Name       string
	Timestamp  time.Time
	Attributes []Attribute
}

// NewEvent creates a new Event.
func NewEvent(name string, opts ...any) Event {
	cfg := parseEventOptions(opts...)
	event := Event{
		Name:       name,
		Timestamp:  cfg.timestamp,
		Attributes: cfg.attributes,
	}
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}
	return event
}

// spanContextKey is the context key for storing spans.
type spanContextKey struct{}

// SpanFromContext returns the Span stored in the context, or nil if none.
func SpanFromContext(ctx context.Context) telemetry.Span {
	if ctx == nil {
		return nil
	}
	if span, ok := ctx.Value(spanContextKey{}).(telemetry.Span); ok {
		return span
	}
	return nil
}

// ContextWithSpan returns a new context with the span stored in it.
func ContextWithSpan(ctx context.Context, span telemetry.Span) context.Context {
	return context.WithValue(ctx, spanContextKey{}, span)
}

// Int64Callback is a callback for observing int64 values.
type Int64Callback func(ctx context.Context, observer Int64Observer)

// Int64Observer is used to observe int64 values.
type Int64Observer interface {
	Observe(value int64, attrs ...Attribute)
}

// Float64Callback is a callback for observing float64 values.
type Float64Callback func(ctx context.Context, observer Float64Observer)

// Float64Observer is used to observe float64 values.
type Float64Observer interface {
	Observe(value float64, attrs ...Attribute)
}
