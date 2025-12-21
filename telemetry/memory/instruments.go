package memory

import (
	"context"
	"time"

	"github.com/grinps/go-utils/telemetry"
)

// baseInstrument provides common functionality for all instruments.
type baseInstrument struct {
	name        string
	description string
	unit        string
}

func (b *baseInstrument) Name() string        { return b.name }
func (b *baseInstrument) Description() string { return b.description }
func (b *baseInstrument) Unit() string        { return b.unit }

// Compile-time checks for interface implementations.
var _ telemetry.Counter[int64] = &Counter[int64]{}
var _ telemetry.Counter[float64] = &Counter[float64]{}
var _ telemetry.Recorder[int64] = &Recorder[int64]{}
var _ telemetry.Recorder[float64] = &Recorder[float64]{}
var _ telemetry.ObservableCounter[int64] = &ObservableCounter[int64]{}
var _ telemetry.ObservableCounter[float64] = &ObservableCounter[float64]{}
var _ telemetry.ObservableGauge[int64] = &ObservableGauge[int64]{}
var _ telemetry.ObservableGauge[float64] = &ObservableGauge[float64]{}

// Counter implements telemetry.Counter[T telemetry.Number].
type Counter[T telemetry.Number] struct {
	baseInstrument
	meter     *Meter
	monotonic bool
	precision telemetry.Precision
}

func (c *Counter[T]) Instrument() {}

// Add adds a value to the counter.
func (c *Counter[T]) Add(ctx context.Context, value T, attrs ...any) {
	c.meter.recordMeasurement(&RecordedMeasurement{
		InstrumentName: c.name,
		InstrumentType: "Counter",
		Value:          value,
		Attributes:     parseAttributes(attrs),
		Timestamp:      time.Now(),
	})
}

// IsMonotonic returns true if the counter only increases.
func (c *Counter[T]) IsMonotonic() bool {
	return c.monotonic
}

// Precision returns the precision of the counter.
func (c *Counter[T]) Precision() telemetry.Precision {
	return c.precision
}

// Recorder implements telemetry.Recorder[T telemetry.Number].
type Recorder[T telemetry.Number] struct {
	baseInstrument
	meter               *Meter
	aggregationStrategy telemetry.AggregationStrategy
	precision           telemetry.Precision
}

func (r *Recorder[T]) Instrument() {}

// Record records a value.
func (r *Recorder[T]) Record(ctx context.Context, value T, attrs ...any) {
	r.meter.recordMeasurement(&RecordedMeasurement{
		InstrumentName: r.name,
		InstrumentType: "Recorder",
		Value:          value,
		Attributes:     parseAttributes(attrs),
		Timestamp:      time.Now(),
	})
}

// IsAggregating returns true if the recorder aggregates values.
func (r *Recorder[T]) IsAggregating() bool {
	return r.aggregationStrategy != telemetry.AggregationStrategyNone &&
		r.aggregationStrategy != telemetry.AggregationStrategyUnknown
}

// AggregationStrategy returns the aggregation strategy.
func (r *Recorder[T]) AggregationStrategy() telemetry.AggregationStrategy {
	return r.aggregationStrategy
}

// Precision returns the precision of the recorder.
func (r *Recorder[T]) Precision() telemetry.Precision {
	return r.precision
}

// parseAttributes extracts Attribute values from any slice.
// It handles both Attribute types and key-value pairs where the key is a string
// followed by any value. This allows callers to pass attributes without depending
// on the memory package's Attribute type:
//
//	counter.Add(ctx, 1, "user.id", "12345", "request.size", 1024)
//
// Or using the Attribute type:
//
//	counter.Add(ctx, 1, memory.String("user.id", "12345"))
func parseAttributes(attrs []any) []Attribute {
	result := make([]Attribute, 0, len(attrs))
	for i := 0; i < len(attrs); i++ {
		switch v := attrs[i].(type) {
		case Attribute:
			result = append(result, v)
		case string:
			// String followed by any value creates a key-value attribute
			if i+1 < len(attrs) {
				// Check next value is not an Attribute (to avoid consuming it)
				if _, isAttr := attrs[i+1].(Attribute); !isAttr {
					result = append(result, Attribute{Key: v, Value: attrs[i+1]})
					i++ // Skip the value we just consumed
				}
			}
		}
	}
	return result
}

// ObservableCounter implements telemetry.ObservableCounter[T telemetry.Number].
type ObservableCounter[T telemetry.Number] struct {
	baseInstrument
	meter      *Meter
	monotonic  bool
	precision  telemetry.Precision
	callback   telemetry.Callback[T]
	registered bool
}

func (c *ObservableCounter[T]) Instrument() {}

// Unregister removes the callback registration.
func (c *ObservableCounter[T]) Unregister() error {
	c.registered = false
	return nil
}

// IsMonotonic returns true if the counter only increases.
func (c *ObservableCounter[T]) IsMonotonic() bool {
	return c.monotonic
}

// Precision returns the precision of the counter.
func (c *ObservableCounter[T]) Precision() telemetry.Precision {
	return c.precision
}

// Collect triggers the callback and records observations.
func (c *ObservableCounter[T]) Collect(ctx context.Context) {
	if !c.registered || c.callback == nil {
		return
	}
	observer := &memoryObserver[T]{
		meter:          c.meter,
		instrumentName: c.name,
		instrumentType: "ObservableCounter",
	}
	c.callback(ctx, observer)
}

// ObservableGauge implements telemetry.ObservableGauge[T telemetry.Number].
type ObservableGauge[T telemetry.Number] struct {
	baseInstrument
	meter      *Meter
	precision  telemetry.Precision
	callback   telemetry.Callback[T]
	registered bool
}

func (g *ObservableGauge[T]) Instrument() {}

// Unregister removes the callback registration.
func (g *ObservableGauge[T]) Unregister() error {
	g.registered = false
	return nil
}

// Precision returns the precision of the gauge.
func (g *ObservableGauge[T]) Precision() telemetry.Precision {
	return g.precision
}

// Collect triggers the callback and records observations.
func (g *ObservableGauge[T]) Collect(ctx context.Context) {
	if !g.registered || g.callback == nil {
		return
	}
	observer := &memoryObserver[T]{
		meter:          g.meter,
		instrumentName: g.name,
		instrumentType: "ObservableGauge",
	}
	g.callback(ctx, observer)
}

// memoryObserver implements telemetry.Observer[T] for memory package.
type memoryObserver[T telemetry.Number] struct {
	meter          *Meter
	instrumentName string
	instrumentType string
}

// Observe records a value with the given attributes.
func (o *memoryObserver[T]) Observe(value T, attrs ...any) {
	o.meter.recordMeasurement(&RecordedMeasurement{
		InstrumentName: o.instrumentName,
		InstrumentType: o.instrumentType,
		Value:          value,
		Attributes:     parseAttributes(attrs),
		Timestamp:      time.Now(),
	})
}
