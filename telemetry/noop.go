package telemetry

import (
	"context"
	"errors"
	"fmt"
)

// ErrorHandlingStrategy defines how errors are handled during telemetry operations.
type ErrorHandlingStrategy string

const (
	// ErrorHandlingStrategyIgnore silently ignores errors.
	ErrorHandlingStrategyIgnore ErrorHandlingStrategy = "ignore"
	// ErrorHandlingStrategyReturn returns errors to the caller.
	ErrorHandlingStrategyReturn ErrorHandlingStrategy = "return"
	// ErrorHandlingStrategyGenerateError generates an error for testing purposes.
	ErrorHandlingStrategyGenerateError ErrorHandlingStrategy = "generate-error"
)

var _ Provider = &NoopProvider{}

// NoopProvider is a no-op implementation of Provider that performs no actual telemetry.
// It is the default provider and can be used when telemetry is not needed.
type NoopProvider struct{}

func (np *NoopProvider) Tracer(name string, opts ...any) (Tracer, error) {
	for _, opt := range opts {
		switch optA := opt.(type) {
		case ErrorHandlingStrategy:
			if optA == ErrorHandlingStrategyGenerateError {
				return nil, ErrTracerCreation.New(ErrReasonNilTracer, ErrParamName, name)
			}
		}
	}
	return &noopTracer{}, nil
}

func (np *NoopProvider) Meter(name string, opts ...any) (Meter, error) {
	for _, opt := range opts {
		switch optA := opt.(type) {
		case ErrorHandlingStrategy:
			if optA == ErrorHandlingStrategyGenerateError {
				return nil, ErrMeterCreation.New(ErrReasonNilMeter, ErrParamName, name)
			}
		}
	}
	return &noopMeter{}, nil
}

func (np *NoopProvider) Shutdown(ctx context.Context) error {
	return nil
}

// noopTracer is a no-op implementation of Tracer.
type noopTracer struct{}

func (t *noopTracer) Start(ctx context.Context, name string, opts ...any) (context.Context, Span) {
	return ctx, &noopSpan{}
}

// noopSpan is a no-op implementation of Span.
type noopSpan struct{}

func (s *noopSpan) End(opts ...any)                        {}
func (s *noopSpan) IsRecording() bool                      { return false }
func (s *noopSpan) SetAttributes(attrs ...any)             {}
func (s *noopSpan) AddEvent(name string, opts ...any)      {}
func (s *noopSpan) RecordError(err error, opts ...any)     {}
func (s *noopSpan) SetStatus(code int, description string) {}
func (s *noopSpan) SetName(name string)                    {}
func (s *noopSpan) TracerProvider() Provider               { return nil }

// noopMeter is a no-op implementation of Meter.
type noopMeter struct{}

type noopInstrumentConfig struct {
	name                string
	instrumentType      InstrumentType
	counterType         CounterType
	aggregationStrategy AggregationStrategy
	precision           Precision
}

func (m *noopMeter) NewInstrument(name string, opts ...any) (Instrument, error) {
	var noopInstrumentConfig = noopInstrumentConfig{
		name:                name,
		instrumentType:      InstrumentTypeUnknown,
		counterType:         CounterTypeUnknown,
		aggregationStrategy: AggregationStrategyUnknown,
		precision:           PrecisionUnknown,
	}
	var instrument Instrument = nil
	var errs []error = []error{}
	var errorHandlingStrategy ErrorHandlingStrategy = ErrorHandlingStrategyIgnore
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		switch optA := opt.(type) {
		case string:
			noopInstrumentConfig.name = optA
		case InstrumentType:
			noopInstrumentConfig.instrumentType = optA
		case CounterType:
			noopInstrumentConfig.counterType = optA
		case AggregationStrategy:
			noopInstrumentConfig.aggregationStrategy = optA
		case Precision:
			noopInstrumentConfig.precision = optA
		case ErrorHandlingStrategy:
			errorHandlingStrategy = optA
		default:
			errs = append(errs, ErrInstrumentCreation.New(ErrReasonInvalidOption, ErrParamOption, opt))
		}
	}
	switch noopInstrumentConfig.instrumentType {
	case InstrumentTypeCounter:
		instrument = createNoopCounter(noopInstrumentConfig.counterType, noopInstrumentConfig.precision)
	case InstrumentTypeRecorder:
		instrument = createNoopRecorder(noopInstrumentConfig.aggregationStrategy, noopInstrumentConfig.precision)
	case InstrumentTypeObservableCounter:
		instrument = createNoopObservableCounter(noopInstrumentConfig.counterType, noopInstrumentConfig.precision)
	case InstrumentTypeObservableGauge:
		instrument = createNoopObservableGauge(noopInstrumentConfig.precision)
	default:
		errs = append(errs, ErrInstrumentCreation.New(ErrReasonInvalidInstrumentType, ErrParamInstrumentType, noopInstrumentConfig.instrumentType))
	}
	if len(errs) > 0 && errorHandlingStrategy == ErrorHandlingStrategyReturn {
		return instrument, fmt.Errorf("failed to create instrument: %w", errors.Join(errs...))
	}
	return instrument, nil
}

// createNoopCounter creates a no-op counter based on counter type and precision.
func createNoopCounter(counterType CounterType, precision Precision) Instrument {
	switch precision {
	case PrecisionFloat64:
		switch counterType {
		case CounterTypeUpDown:
			return &noopFloat64UpDownCounter{}
		default:
			return &noopFloat64Counter{}
		}
	default: // Default to Int64
		switch counterType {
		case CounterTypeUpDown:
			return &noopInt64UpDownCounter{}
		default:
			return &noopInt64Counter{}
		}
	}
}

// createNoopRecorder creates a no-op recorder based on aggregation strategy and precision.
func createNoopRecorder(strategy AggregationStrategy, precision Precision) Instrument {
	switch precision {
	case PrecisionInt64:
		switch strategy {
		case AggregationStrategyHistogram:
			return &noopInt64Histogram{}
		default:
			return &noopInt64Gauge{}
		}
	default: // Default to Float64
		switch strategy {
		case AggregationStrategyHistogram:
			return &noopFloat64Histogram{}
		default:
			return &noopFloat64Gauge{}
		}
	}
}

// No-op counter instruments

var _ Counter[int64] = &noopInt64Counter{}
var _ Counter[int64] = &noopInt64UpDownCounter{}
var _ Counter[float64] = &noopFloat64Counter{}
var _ Counter[float64] = &noopFloat64UpDownCounter{}

type noopInt64Counter struct{}

func (n *noopInt64Counter) Instrument()                                        {}
func (n *noopInt64Counter) Add(ctx context.Context, value int64, attrs ...any) {}
func (n *noopInt64Counter) IsMonotonic() bool                                  { return true }
func (n *noopInt64Counter) Precision() Precision                               { return PrecisionInt64 }

type noopInt64UpDownCounter struct{}

func (n *noopInt64UpDownCounter) Instrument()                                        {}
func (n *noopInt64UpDownCounter) Add(ctx context.Context, value int64, attrs ...any) {}
func (n *noopInt64UpDownCounter) IsMonotonic() bool                                  { return false }
func (n *noopInt64UpDownCounter) Precision() Precision                               { return PrecisionInt64 }

type noopFloat64Counter struct{}

func (n *noopFloat64Counter) Instrument()                                          {}
func (n *noopFloat64Counter) Add(ctx context.Context, value float64, attrs ...any) {}
func (n *noopFloat64Counter) IsMonotonic() bool                                    { return true }
func (n *noopFloat64Counter) Precision() Precision                                 { return PrecisionFloat64 }

type noopFloat64UpDownCounter struct{}

func (n *noopFloat64UpDownCounter) Instrument()                                          {}
func (n *noopFloat64UpDownCounter) Add(ctx context.Context, value float64, attrs ...any) {}
func (n *noopFloat64UpDownCounter) IsMonotonic() bool                                    { return false }
func (n *noopFloat64UpDownCounter) Precision() Precision                                 { return PrecisionFloat64 }

// No-op recorder instruments

var _ Recorder[int64] = &noopInt64Gauge{}
var _ Recorder[int64] = &noopInt64Histogram{}
var _ Recorder[float64] = &noopFloat64Gauge{}
var _ Recorder[float64] = &noopFloat64Histogram{}

type noopInt64Gauge struct{}

func (n *noopInt64Gauge) Instrument()                                           {}
func (n *noopInt64Gauge) Record(ctx context.Context, value int64, attrs ...any) {}
func (n *noopInt64Gauge) IsAggregating() bool                                   { return false }
func (n *noopInt64Gauge) AggregationStrategy() AggregationStrategy              { return AggregationStrategyNone }
func (n *noopInt64Gauge) Precision() Precision                                  { return PrecisionInt64 }

type noopInt64Histogram struct{}

func (n *noopInt64Histogram) Instrument()                                           {}
func (n *noopInt64Histogram) Record(ctx context.Context, value int64, attrs ...any) {}
func (n *noopInt64Histogram) IsAggregating() bool                                   { return true }
func (n *noopInt64Histogram) AggregationStrategy() AggregationStrategy {
	return AggregationStrategyHistogram
}
func (n *noopInt64Histogram) Precision() Precision { return PrecisionInt64 }

type noopFloat64Gauge struct{}

func (n *noopFloat64Gauge) Instrument()                                             {}
func (n *noopFloat64Gauge) Record(ctx context.Context, value float64, attrs ...any) {}
func (n *noopFloat64Gauge) IsAggregating() bool                                     { return false }
func (n *noopFloat64Gauge) AggregationStrategy() AggregationStrategy                { return AggregationStrategyNone }
func (n *noopFloat64Gauge) Precision() Precision                                    { return PrecisionFloat64 }

type noopFloat64Histogram struct{}

func (n *noopFloat64Histogram) Instrument()                                             {}
func (n *noopFloat64Histogram) Record(ctx context.Context, value float64, attrs ...any) {}
func (n *noopFloat64Histogram) IsAggregating() bool                                     { return true }
func (n *noopFloat64Histogram) AggregationStrategy() AggregationStrategy {
	return AggregationStrategyHistogram
}
func (n *noopFloat64Histogram) Precision() Precision { return PrecisionFloat64 }

// No-op observable counter factory

func createNoopObservableCounter(counterType CounterType, precision Precision) Instrument {
	switch precision {
	case PrecisionFloat64:
		switch counterType {
		case CounterTypeUpDown:
			return &noopObservableFloat64UpDownCounter{}
		default:
			return &noopObservableFloat64Counter{}
		}
	default: // Default to Int64
		switch counterType {
		case CounterTypeUpDown:
			return &noopObservableInt64UpDownCounter{}
		default:
			return &noopObservableInt64Counter{}
		}
	}
}

// No-op observable gauge factory

func createNoopObservableGauge(precision Precision) Instrument {
	switch precision {
	case PrecisionInt64:
		return &noopObservableInt64Gauge{}
	default: // Default to Float64
		return &noopObservableFloat64Gauge{}
	}
}

// No-op observable counter instruments

var _ ObservableCounter[int64] = &noopObservableInt64Counter{}
var _ ObservableCounter[int64] = &noopObservableInt64UpDownCounter{}
var _ ObservableCounter[float64] = &noopObservableFloat64Counter{}
var _ ObservableCounter[float64] = &noopObservableFloat64UpDownCounter{}

type noopObservableInt64Counter struct{}

func (n *noopObservableInt64Counter) Instrument()          {}
func (n *noopObservableInt64Counter) Unregister() error    { return nil }
func (n *noopObservableInt64Counter) IsMonotonic() bool    { return true }
func (n *noopObservableInt64Counter) Precision() Precision { return PrecisionInt64 }

type noopObservableInt64UpDownCounter struct{}

func (n *noopObservableInt64UpDownCounter) Instrument()          {}
func (n *noopObservableInt64UpDownCounter) Unregister() error    { return nil }
func (n *noopObservableInt64UpDownCounter) IsMonotonic() bool    { return false }
func (n *noopObservableInt64UpDownCounter) Precision() Precision { return PrecisionInt64 }

type noopObservableFloat64Counter struct{}

func (n *noopObservableFloat64Counter) Instrument()          {}
func (n *noopObservableFloat64Counter) Unregister() error    { return nil }
func (n *noopObservableFloat64Counter) IsMonotonic() bool    { return true }
func (n *noopObservableFloat64Counter) Precision() Precision { return PrecisionFloat64 }

type noopObservableFloat64UpDownCounter struct{}

func (n *noopObservableFloat64UpDownCounter) Instrument()          {}
func (n *noopObservableFloat64UpDownCounter) Unregister() error    { return nil }
func (n *noopObservableFloat64UpDownCounter) IsMonotonic() bool    { return false }
func (n *noopObservableFloat64UpDownCounter) Precision() Precision { return PrecisionFloat64 }

// No-op observable gauge instruments

var _ ObservableGauge[int64] = &noopObservableInt64Gauge{}
var _ ObservableGauge[float64] = &noopObservableFloat64Gauge{}

type noopObservableInt64Gauge struct{}

func (n *noopObservableInt64Gauge) Instrument()          {}
func (n *noopObservableInt64Gauge) Unregister() error    { return nil }
func (n *noopObservableInt64Gauge) Precision() Precision { return PrecisionInt64 }

type noopObservableFloat64Gauge struct{}

func (n *noopObservableFloat64Gauge) Instrument()          {}
func (n *noopObservableFloat64Gauge) Unregister() error    { return nil }
func (n *noopObservableFloat64Gauge) Precision() Precision { return PrecisionFloat64 }
