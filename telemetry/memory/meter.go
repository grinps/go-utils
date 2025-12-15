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
		switch cfg.counterType {
		case telemetry.CounterTypeMonotonic:
			inst = &Counter[int64]{
				baseInstrument: baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
				meter:          m,
				monotonic:      true,
			}
		case telemetry.CounterTypeUpDown:
			inst = &Counter[int64]{
				baseInstrument: baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
				meter:          m,
				monotonic:      false,
			}
		default:
			// Default to monotonic counter
			inst = &Counter[int64]{
				baseInstrument: baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
				meter:          m,
				monotonic:      true,
			}
		}
	case telemetry.InstrumentTypeRecorder:
		switch cfg.aggregationStrategy {
		case telemetry.AggregationStrategyNone:
			inst = &Recorder[float64]{
				baseInstrument:      baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
				meter:               m,
				aggregationStrategy: telemetry.AggregationStrategyNone,
			}
		case telemetry.AggregationStrategyHistogram:
			inst = &Recorder[float64]{
				baseInstrument:      baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
				meter:               m,
				aggregationStrategy: telemetry.AggregationStrategyHistogram,
			}
		default:
			// Default to no aggregation (gauge)
			inst = &Recorder[float64]{
				baseInstrument:      baseInstrument{name: name, description: cfg.description, unit: cfg.unit},
				meter:               m,
				aggregationStrategy: telemetry.AggregationStrategyNone,
			}
		}
	default:
		// Return nil for unknown instrument type (matches NoopProvider behavior)
		return nil, nil
	}

	m.instruments[name] = inst
	return inst, nil
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
