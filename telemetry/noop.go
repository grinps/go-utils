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
}

func (m *noopMeter) NewInstrument(name string, opts ...any) (Instrument, error) {
	var noopInstrumentConfig = noopInstrumentConfig{
		name:                name,
		instrumentType:      InstrumentTypeUnknown,
		counterType:         CounterTypeUnknown,
		aggregationStrategy: AggregationStrategyUnknown,
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
		case ErrorHandlingStrategy:
			errorHandlingStrategy = optA
		default:
			errs = append(errs, ErrInstrumentCreation.New(ErrReasonInvalidOption, ErrParamOption, opt))
		}
	}
	switch noopInstrumentConfig.instrumentType {
	case InstrumentTypeCounter:
		switch noopInstrumentConfig.counterType {
		case CounterTypeMonotonic:
			instrument = &noopInt64Counter{}
		default:
			errs = append(errs, ErrInstrumentCreation.New(ErrReasonInvalidCounterType, ErrParamCounterType, noopInstrumentConfig.counterType))
		}
	case InstrumentTypeRecorder:
		switch noopInstrumentConfig.aggregationStrategy {
		case AggregationStrategyNone:
			instrument = &noopFloat64Gauge{}
		default:
			errs = append(errs, ErrInstrumentCreation.New(ErrReasonInvalidAggregationStrategy, ErrParamAggregationStrategy, noopInstrumentConfig.aggregationStrategy))
		}
	default:
		errs = append(errs, ErrInstrumentCreation.New(ErrReasonInvalidInstrumentType, ErrParamInstrumentType, noopInstrumentConfig.instrumentType))
	}
	if len(errs) > 0 && errorHandlingStrategy == ErrorHandlingStrategyReturn {
		return instrument, fmt.Errorf("failed to create instrument: %w", errors.Join(errs...))
	}
	return instrument, nil
}

// No-op instruments for shutdown provider

type noopInstrument struct{}

func (n *noopInstrument) Name() string        { return "" }
func (n *noopInstrument) Description() string { return "" }
func (n *noopInstrument) Unit() string        { return "" }

var _ Counter[int64] = &noopInt64Counter{}
var _ Recorder[float64] = &noopFloat64Gauge{}

type noopInt64Counter struct{ noopInstrument }

func (n *noopInt64Counter) Add(ctx context.Context, value int64, attrs ...any) {}
func (n *noopInt64Counter) IsMonotonic() bool                                  { return true }

type noopFloat64Gauge struct{ noopInstrument }

func (n *noopFloat64Gauge) Record(ctx context.Context, value float64, attrs ...any) {}
func (n *noopFloat64Gauge) IsAggregating() bool                                     { return false }
func (n *noopFloat64Gauge) AggregationStrategy() AggregationStrategy                { return AggregationStrategyNone }
