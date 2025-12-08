package telemetry

import "context"

type Number interface {
	float64 | int64
}

// Meter is responsible for creating metric instruments.
// It is obtained from a Provider and should be named to identify
// the instrumentation library or application component.
//
// Implementations must be safe for concurrent use.
type Meter interface {
	// NewInstrument creates a new instrument based on various options
	// The name must be unique within this Meter.
	NewInstrument(name string, opts ...any) (Instrument, error)
}

// Instrument is the base interface for all metric instruments.
type Instrument interface {
	// Name returns the name of the instrument.
	Name() string
	// Description returns the description of the instrument.
	Description() string
	// Unit returns the unit of the instrument.
	Unit() string
}

// Synchronous instruments are used to record measurements as they happen.
type InstrumentType string

const (
	InstrumentTypeUnknown  InstrumentType = ""
	InstrumentTypeCounter  InstrumentType = "counter"
	InstrumentTypeRecorder InstrumentType = "recorder"
)

type CounterType string

const (
	CounterTypeUnknown   CounterType = ""
	CounterTypeMonotonic CounterType = "monotonic"
	CounterTypeUpDown    CounterType = "updown"
)

// Counter is a synchronous instrument for recording increments and decrements.
type Counter[T Number] interface {
	Instrument
	IsMonotonic() bool
	Add(ctx context.Context, value T, attrs ...any)
}

type AggregationStrategy string

func (aS AggregationStrategy) Name() string {
	return string(aS)
}

const (
	// AggregationStrategyUnknown is used when the aggregation strategy is unknown.
	AggregationStrategyUnknown AggregationStrategy = "unknown"
	// AggregationStrategyNone is used when the aggregation strategy is none i.e point-in-time value are reported.
	AggregationStrategyNone AggregationStrategy = "none"
	// AggregationStrategyHistogram is used when the aggregation strategy is histogram.
	AggregationStrategyHistogram AggregationStrategy = "histogram"
)

// Recorder is a synchronous instrument for recording point-in-time values.
type Recorder[T Number] interface {
	Instrument
	IsAggregating() bool
	AggregationStrategy() AggregationStrategy
	Record(ctx context.Context, value T, attrs ...any)
}

// Callback is a function called to observe values.
type Callback[T Number] func(ctx context.Context, observer Observer[T])

// Observer is used by Callback to report observations.
type Observer[T Number] interface {
	// Observe records a value with the given attributes.
	Observe(value T, opts ...any)
}

// ObservableInstrument is the base interface for asynchronous instruments.
type ObservableInstrument interface {
	Instrument
}
