package memory

import (
	"sync"
	"time"

	"github.com/grinps/go-utils/telemetry"
)

// Compile-time check that Meter implements telemetry.Meter.
var _ telemetry.Meter = (*Meter)(nil)

// Meter is an in-memory implementation of telemetry.Meter.
type Meter struct {
	mu           sync.RWMutex
	provider     *Provider
	name         string
	version      string
	schemaURL    string
	attributes   []Attribute
	instruments  map[string]telemetry.Instrument
	measurements []*RecordedMeasurement
}

// reset clears all recorded measurements.
func (m *Meter) reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.measurements = nil
}

// RecordedMeasurements returns all recorded measurements for this meter.
func (m *Meter) RecordedMeasurements() []*RecordedMeasurement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]*RecordedMeasurement, len(m.measurements))
	copy(result, m.measurements)
	return result
}

// RecordedMeasurementsByName returns all recorded measurements with the given instrument name.
func (m *Meter) RecordedMeasurementsByName(name string) []*RecordedMeasurement {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var result []*RecordedMeasurement
	for _, measurement := range m.measurements {
		if measurement.InstrumentName == name {
			result = append(result, measurement)
		}
	}
	return result
}

func (m *Meter) recordMeasurement(measurement *RecordedMeasurement) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.measurements = append(m.measurements, measurement)
}

// NewInstrument creates a new instrument based on options.
// The name must be unique within this Meter.
// Options should include telemetry.InstrumentType and optionally telemetry.CounterType
// or telemetry.AggregationStrategy depending on the instrument type.
func (m *Meter) NewInstrument(name string, opts ...any) (telemetry.Instrument, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.instruments == nil {
		m.instruments = make(map[string]telemetry.Instrument)
	}

	if existing, ok := m.instruments[name]; ok {
		return existing, nil
	}

	cfg := parseInstrumentOptions(opts...)

	var inst telemetry.Instrument
	switch cfg.instrumentType {
	case telemetry.InstrumentTypeCounter:
		inst = m.createCounter(name, cfg)
	case telemetry.InstrumentTypeRecorder:
		inst = m.createRecorder(name, cfg)
	case telemetry.InstrumentTypeObservableCounter:
		inst = m.createObservableCounter(name, cfg, opts)
	case telemetry.InstrumentTypeObservableGauge:
		inst = m.createObservableGauge(name, cfg, opts)
	default:
		// Return nil for unknown instrument type (matches NoopProvider behavior)
		return nil, nil
	}

	m.instruments[name] = inst
	return inst, nil
}

// createCounter creates a counter instrument based on configuration.
func (m *Meter) createCounter(name string, cfg *InstrumentConfig) telemetry.Instrument {
	monotonic := cfg.counterType != telemetry.CounterTypeUpDown
	switch cfg.precision {
	case telemetry.PrecisionFloat64:
		return &Counter[float64]{
			baseInstrument: baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
			meter:          m,
			monotonic:      monotonic,
			precision:      telemetry.PrecisionFloat64,
		}
	default: // Default to Int64
		return &Counter[int64]{
			baseInstrument: baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
			meter:          m,
			monotonic:      monotonic,
			precision:      telemetry.PrecisionInt64,
		}
	}
}

// createRecorder creates a recorder instrument based on configuration.
func (m *Meter) createRecorder(name string, cfg *InstrumentConfig) telemetry.Instrument {
	strategy := cfg.aggregationStrategy
	if strategy == telemetry.AggregationStrategyUnknown {
		strategy = telemetry.AggregationStrategyNone
	}
	switch cfg.precision {
	case telemetry.PrecisionInt64:
		return &Recorder[int64]{
			baseInstrument:      baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
			meter:               m,
			aggregationStrategy: strategy,
			precision:           telemetry.PrecisionInt64,
		}
	default: // Default to Float64
		return &Recorder[float64]{
			baseInstrument:      baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
			meter:               m,
			aggregationStrategy: strategy,
			precision:           telemetry.PrecisionFloat64,
		}
	}
}

// createObservableCounter creates an observable counter instrument based on configuration.
func (m *Meter) createObservableCounter(name string, cfg *InstrumentConfig, opts []any) telemetry.Instrument {
	monotonic := cfg.counterType != telemetry.CounterTypeUpDown
	switch cfg.precision {
	case telemetry.PrecisionFloat64:
		var callback telemetry.Callback[float64]
		for _, opt := range opts {
			if cb, ok := opt.(telemetry.Callback[float64]); ok {
				callback = cb
				break
			}
		}
		return &ObservableCounter[float64]{
			baseInstrument: baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
			meter:          m,
			monotonic:      monotonic,
			precision:      telemetry.PrecisionFloat64,
			callback:       callback,
			registered:     callback != nil,
		}
	default: // Default to Int64
		var callback telemetry.Callback[int64]
		for _, opt := range opts {
			if cb, ok := opt.(telemetry.Callback[int64]); ok {
				callback = cb
				break
			}
		}
		return &ObservableCounter[int64]{
			baseInstrument: baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
			meter:          m,
			monotonic:      monotonic,
			precision:      telemetry.PrecisionInt64,
			callback:       callback,
			registered:     callback != nil,
		}
	}
}

// createObservableGauge creates an observable gauge instrument based on configuration.
func (m *Meter) createObservableGauge(name string, cfg *InstrumentConfig, opts []any) telemetry.Instrument {
	switch cfg.precision {
	case telemetry.PrecisionInt64:
		var callback telemetry.Callback[int64]
		for _, opt := range opts {
			if cb, ok := opt.(telemetry.Callback[int64]); ok {
				callback = cb
				break
			}
		}
		return &ObservableGauge[int64]{
			baseInstrument: baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
			meter:          m,
			precision:      telemetry.PrecisionInt64,
			callback:       callback,
			registered:     callback != nil,
		}
	default: // Default to Float64
		var callback telemetry.Callback[float64]
		for _, opt := range opts {
			if cb, ok := opt.(telemetry.Callback[float64]); ok {
				callback = cb
				break
			}
		}
		return &ObservableGauge[float64]{
			baseInstrument: baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
			meter:          m,
			precision:      telemetry.PrecisionFloat64,
			callback:       callback,
			registered:     callback != nil,
		}
	}
}

// RecordedMeasurement represents a measurement that was recorded.
type RecordedMeasurement struct {
	// InstrumentName is the name of the instrument.
	InstrumentName string
	// InstrumentType is the type of the instrument.
	InstrumentType string
	// Value is the recorded value.
	Value any
	// Attributes are the measurement attributes.
	Attributes []Attribute
	// Timestamp is when the measurement was recorded.
	Timestamp time.Time
}
