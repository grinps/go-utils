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

// Counter implements telemetry.Counter[T telemetry.Number].
type Counter[T telemetry.Number] struct {
	baseInstrument
	meter     *Meter
	monotonic bool
}

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

// Recorder implements telemetry.Recorder[T telemetry.Number].
type Recorder[T telemetry.Number] struct {
	baseInstrument
	meter               *Meter
	aggregationStrategy telemetry.AggregationStrategy
}

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
