package otel

import (
	"context"

	"github.com/grinps/go-utils/telemetry"
	"go.opentelemetry.io/otel/metric"
)

// Meter wraps an OpenTelemetry meter using embedding.
type Meter struct {
	metric.Meter
}

// NewInstrument creates a new instrument based on options.
// Supports telemetry.InstrumentType, telemetry.CounterType, telemetry.AggregationStrategy.
// OTEL options (metric.InstrumentOption) are passed through to the underlying meter.
func (m *Meter) NewInstrument(name string, opts ...any) (telemetry.Instrument, error) {
	var (
		instType            = telemetry.InstrumentTypeCounter
		counterType         = telemetry.CounterTypeMonotonic
		aggregationStrategy = telemetry.AggregationStrategyNone
		precision           = telemetry.PrecisionUnknown
		otelOpts            []metric.InstrumentOption
	)

	for _, opt := range opts {
		switch v := opt.(type) {
		case telemetry.InstrumentType:
			instType = v
		case telemetry.CounterType:
			counterType = v
		case telemetry.AggregationStrategy:
			aggregationStrategy = v
		case telemetry.Precision:
			precision = v
		case metric.InstrumentOption:
			otelOpts = append(otelOpts, v)
		}
	}

	switch instType {
	case telemetry.InstrumentTypeCounter:
		return m.createCounter(name, counterType, precision, otelOpts)
	case telemetry.InstrumentTypeRecorder:
		return m.createRecorder(name, aggregationStrategy, precision, otelOpts)
	case telemetry.InstrumentTypeObservableCounter:
		return m.createObservableCounter(name, counterType, precision, opts)
	case telemetry.InstrumentTypeObservableGauge:
		return m.createObservableGauge(name, precision, opts)
	default:
		return m.createCounter(name, counterType, precision, otelOpts)
	}
}

func (m *Meter) createCounter(name string, cType telemetry.CounterType, precision telemetry.Precision, opts []metric.InstrumentOption) (telemetry.Instrument, error) {
	switch precision {
	case telemetry.PrecisionFloat64:
		return m.createFloat64Counter(name, cType, opts)
	default: // Default to Int64
		return m.createInt64Counter(name, cType, opts)
	}
}

func (m *Meter) createInt64Counter(name string, cType telemetry.CounterType, opts []metric.InstrumentOption) (telemetry.Instrument, error) {
	switch cType {
	case telemetry.CounterTypeUpDown:
		var otelOpts []metric.Int64UpDownCounterOption
		for _, o := range opts {
			if v, ok := o.(metric.Int64UpDownCounterOption); ok {
				otelOpts = append(otelOpts, v)
			}
		}
		counter, err := m.Meter.Int64UpDownCounter(name, otelOpts...)
		if err != nil {
			return nil, err
		}
		return &Int64UpDownCounter{counter: counter}, nil
	default:
		var otelOpts []metric.Int64CounterOption
		for _, o := range opts {
			if v, ok := o.(metric.Int64CounterOption); ok {
				otelOpts = append(otelOpts, v)
			}
		}
		counter, err := m.Meter.Int64Counter(name, otelOpts...)
		if err != nil {
			return nil, err
		}
		return &Int64Counter{counter: counter}, nil
	}
}

func (m *Meter) createFloat64Counter(name string, cType telemetry.CounterType, opts []metric.InstrumentOption) (telemetry.Instrument, error) {
	switch cType {
	case telemetry.CounterTypeUpDown:
		var otelOpts []metric.Float64UpDownCounterOption
		for _, o := range opts {
			if v, ok := o.(metric.Float64UpDownCounterOption); ok {
				otelOpts = append(otelOpts, v)
			}
		}
		counter, err := m.Meter.Float64UpDownCounter(name, otelOpts...)
		if err != nil {
			return nil, err
		}
		return &Float64UpDownCounter{counter: counter}, nil
	default:
		var otelOpts []metric.Float64CounterOption
		for _, o := range opts {
			if v, ok := o.(metric.Float64CounterOption); ok {
				otelOpts = append(otelOpts, v)
			}
		}
		counter, err := m.Meter.Float64Counter(name, otelOpts...)
		if err != nil {
			return nil, err
		}
		return &Float64Counter{counter: counter}, nil
	}
}

func (m *Meter) createRecorder(name string, strategy telemetry.AggregationStrategy, precision telemetry.Precision, opts []metric.InstrumentOption) (telemetry.Instrument, error) {
	switch precision {
	case telemetry.PrecisionInt64:
		return m.createInt64Recorder(name, strategy, opts)
	default: // Default to Float64
		return m.createFloat64Recorder(name, strategy, opts)
	}
}

func (m *Meter) createInt64Recorder(name string, strategy telemetry.AggregationStrategy, opts []metric.InstrumentOption) (telemetry.Instrument, error) {
	switch strategy {
	case telemetry.AggregationStrategyHistogram:
		var otelOpts []metric.Int64HistogramOption
		for _, o := range opts {
			if v, ok := o.(metric.Int64HistogramOption); ok {
				otelOpts = append(otelOpts, v)
			}
		}
		histogram, err := m.Meter.Int64Histogram(name, otelOpts...)
		if err != nil {
			return nil, err
		}
		return &Int64Histogram{histogram: histogram}, nil
	default:
		var otelOpts []metric.Int64GaugeOption
		for _, o := range opts {
			if v, ok := o.(metric.Int64GaugeOption); ok {
				otelOpts = append(otelOpts, v)
			}
		}
		gauge, err := m.Meter.Int64Gauge(name, otelOpts...)
		if err != nil {
			return nil, err
		}
		return &Int64Gauge{gauge: gauge}, nil
	}
}

func (m *Meter) createFloat64Recorder(name string, strategy telemetry.AggregationStrategy, opts []metric.InstrumentOption) (telemetry.Instrument, error) {
	switch strategy {
	case telemetry.AggregationStrategyHistogram:
		var otelOpts []metric.Float64HistogramOption
		for _, o := range opts {
			if v, ok := o.(metric.Float64HistogramOption); ok {
				otelOpts = append(otelOpts, v)
			}
		}
		histogram, err := m.Meter.Float64Histogram(name, otelOpts...)
		if err != nil {
			return nil, err
		}
		return &Float64Histogram{histogram: histogram}, nil
	default:
		var otelOpts []metric.Float64GaugeOption
		for _, o := range opts {
			if v, ok := o.(metric.Float64GaugeOption); ok {
				otelOpts = append(otelOpts, v)
			}
		}
		gauge, err := m.Meter.Float64Gauge(name, otelOpts...)
		if err != nil {
			return nil, err
		}
		return &Float64Gauge{gauge: gauge}, nil
	}
}

// Int64Counter is a monotonic counter instrument.
type Int64Counter struct {
	counter metric.Int64Counter
}

func (c *Int64Counter) Instrument()                    {}
func (c *Int64Counter) IsMonotonic() bool              { return true }
func (c *Int64Counter) Precision() telemetry.Precision { return telemetry.PrecisionInt64 }

// Add increments the counter by the given value.
// Filters attrs for metric.AddOption and ignores non-matching parameters.
func (c *Int64Counter) Add(ctx context.Context, value int64, attrs ...any) {
	var otelOpts []metric.AddOption
	for _, a := range attrs {
		if v, ok := a.(metric.AddOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	c.counter.Add(ctx, value, otelOpts...)
}

// Int64UpDownCounter is a non-monotonic counter instrument.
type Int64UpDownCounter struct {
	counter metric.Int64UpDownCounter
}

func (c *Int64UpDownCounter) Instrument()                    {}
func (c *Int64UpDownCounter) IsMonotonic() bool              { return false }
func (c *Int64UpDownCounter) Precision() telemetry.Precision { return telemetry.PrecisionInt64 }

// Add increments or decrements the counter by the given value.
func (c *Int64UpDownCounter) Add(ctx context.Context, value int64, attrs ...any) {
	var otelOpts []metric.AddOption
	for _, a := range attrs {
		if v, ok := a.(metric.AddOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	c.counter.Add(ctx, value, otelOpts...)
}

// Float64Counter is a monotonic counter instrument for float64 values.
type Float64Counter struct {
	counter metric.Float64Counter
}

func (c *Float64Counter) Instrument()                    {}
func (c *Float64Counter) IsMonotonic() bool              { return true }
func (c *Float64Counter) Precision() telemetry.Precision { return telemetry.PrecisionFloat64 }

// Add increments the counter by the given value.
func (c *Float64Counter) Add(ctx context.Context, value float64, attrs ...any) {
	var otelOpts []metric.AddOption
	for _, a := range attrs {
		if v, ok := a.(metric.AddOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	c.counter.Add(ctx, value, otelOpts...)
}

// Float64UpDownCounter is a non-monotonic counter instrument for float64 values.
type Float64UpDownCounter struct {
	counter metric.Float64UpDownCounter
}

func (c *Float64UpDownCounter) Instrument()                    {}
func (c *Float64UpDownCounter) IsMonotonic() bool              { return false }
func (c *Float64UpDownCounter) Precision() telemetry.Precision { return telemetry.PrecisionFloat64 }

// Add increments or decrements the counter by the given value.
func (c *Float64UpDownCounter) Add(ctx context.Context, value float64, attrs ...any) {
	var otelOpts []metric.AddOption
	for _, a := range attrs {
		if v, ok := a.(metric.AddOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	c.counter.Add(ctx, value, otelOpts...)
}

// Int64Gauge is a gauge instrument for point-in-time int64 values.
type Int64Gauge struct {
	gauge metric.Int64Gauge
}

func (g *Int64Gauge) Instrument()         {}
func (g *Int64Gauge) IsAggregating() bool { return false }
func (g *Int64Gauge) AggregationStrategy() telemetry.AggregationStrategy {
	return telemetry.AggregationStrategyNone
}
func (g *Int64Gauge) Precision() telemetry.Precision { return telemetry.PrecisionInt64 }

// Record records a value.
func (g *Int64Gauge) Record(ctx context.Context, value int64, attrs ...any) {
	var otelOpts []metric.RecordOption
	for _, a := range attrs {
		if v, ok := a.(metric.RecordOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	g.gauge.Record(ctx, value, otelOpts...)
}

// Int64Histogram is a histogram instrument for int64 distributions.
type Int64Histogram struct {
	histogram metric.Int64Histogram
}

func (h *Int64Histogram) Instrument()         {}
func (h *Int64Histogram) IsAggregating() bool { return true }
func (h *Int64Histogram) AggregationStrategy() telemetry.AggregationStrategy {
	return telemetry.AggregationStrategyHistogram
}
func (h *Int64Histogram) Precision() telemetry.Precision { return telemetry.PrecisionInt64 }

// Record records a value.
func (h *Int64Histogram) Record(ctx context.Context, value int64, attrs ...any) {
	var otelOpts []metric.RecordOption
	for _, a := range attrs {
		if v, ok := a.(metric.RecordOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	h.histogram.Record(ctx, value, otelOpts...)
}

// Float64Gauge is a gauge instrument for point-in-time float64 values.
type Float64Gauge struct {
	gauge metric.Float64Gauge
}

func (g *Float64Gauge) Instrument()         {}
func (g *Float64Gauge) IsAggregating() bool { return false }
func (g *Float64Gauge) AggregationStrategy() telemetry.AggregationStrategy {
	return telemetry.AggregationStrategyNone
}
func (g *Float64Gauge) Precision() telemetry.Precision { return telemetry.PrecisionFloat64 }

// Record records a value.
func (g *Float64Gauge) Record(ctx context.Context, value float64, attrs ...any) {
	var otelOpts []metric.RecordOption
	for _, a := range attrs {
		if v, ok := a.(metric.RecordOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	g.gauge.Record(ctx, value, otelOpts...)
}

// Float64Histogram is a histogram instrument for float64 distributions.
type Float64Histogram struct {
	histogram metric.Float64Histogram
}

func (h *Float64Histogram) Instrument()         {}
func (h *Float64Histogram) IsAggregating() bool { return true }
func (h *Float64Histogram) AggregationStrategy() telemetry.AggregationStrategy {
	return telemetry.AggregationStrategyHistogram
}
func (h *Float64Histogram) Precision() telemetry.Precision { return telemetry.PrecisionFloat64 }

// Record records a value.
func (h *Float64Histogram) Record(ctx context.Context, value float64, attrs ...any) {
	var otelOpts []metric.RecordOption
	for _, a := range attrs {
		if v, ok := a.(metric.RecordOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	h.histogram.Record(ctx, value, otelOpts...)
}

// Observable counter creation

func (m *Meter) createObservableCounter(name string, cType telemetry.CounterType, precision telemetry.Precision, opts []any) (telemetry.Instrument, error) {
	switch precision {
	case telemetry.PrecisionFloat64:
		return m.createFloat64ObservableCounter(name, cType, opts)
	default: // Default to Int64
		return m.createInt64ObservableCounter(name, cType, opts)
	}
}

func (m *Meter) createInt64ObservableCounter(name string, cType telemetry.CounterType, opts []any) (telemetry.Instrument, error) {
	var callback telemetry.Callback[int64]
	for _, opt := range opts {
		if cb, ok := opt.(telemetry.Callback[int64]); ok {
			callback = cb
			break
		}
	}

	switch cType {
	case telemetry.CounterTypeUpDown:
		counter, err := m.Meter.Int64ObservableUpDownCounter(name)
		if err != nil {
			return nil, err
		}
		inst := &Int64ObservableUpDownCounter{counter: counter, callback: callback}
		if callback != nil {
			reg, err := m.Meter.RegisterCallback(inst.observe, counter)
			if err != nil {
				return nil, err
			}
			inst.registration = reg
		}
		return inst, nil
	default:
		counter, err := m.Meter.Int64ObservableCounter(name)
		if err != nil {
			return nil, err
		}
		inst := &Int64ObservableCounter{counter: counter, callback: callback}
		if callback != nil {
			reg, err := m.Meter.RegisterCallback(inst.observe, counter)
			if err != nil {
				return nil, err
			}
			inst.registration = reg
		}
		return inst, nil
	}
}

func (m *Meter) createFloat64ObservableCounter(name string, cType telemetry.CounterType, opts []any) (telemetry.Instrument, error) {
	var callback telemetry.Callback[float64]
	for _, opt := range opts {
		if cb, ok := opt.(telemetry.Callback[float64]); ok {
			callback = cb
			break
		}
	}

	switch cType {
	case telemetry.CounterTypeUpDown:
		counter, err := m.Meter.Float64ObservableUpDownCounter(name)
		if err != nil {
			return nil, err
		}
		inst := &Float64ObservableUpDownCounter{counter: counter, callback: callback}
		if callback != nil {
			reg, err := m.Meter.RegisterCallback(inst.observe, counter)
			if err != nil {
				return nil, err
			}
			inst.registration = reg
		}
		return inst, nil
	default:
		counter, err := m.Meter.Float64ObservableCounter(name)
		if err != nil {
			return nil, err
		}
		inst := &Float64ObservableCounter{counter: counter, callback: callback}
		if callback != nil {
			reg, err := m.Meter.RegisterCallback(inst.observe, counter)
			if err != nil {
				return nil, err
			}
			inst.registration = reg
		}
		return inst, nil
	}
}

// Observable gauge creation

func (m *Meter) createObservableGauge(name string, precision telemetry.Precision, opts []any) (telemetry.Instrument, error) {
	switch precision {
	case telemetry.PrecisionInt64:
		return m.createInt64ObservableGauge(name, opts)
	default: // Default to Float64
		return m.createFloat64ObservableGauge(name, opts)
	}
}

func (m *Meter) createInt64ObservableGauge(name string, opts []any) (telemetry.Instrument, error) {
	var callback telemetry.Callback[int64]
	for _, opt := range opts {
		if cb, ok := opt.(telemetry.Callback[int64]); ok {
			callback = cb
			break
		}
	}

	gauge, err := m.Meter.Int64ObservableGauge(name)
	if err != nil {
		return nil, err
	}
	inst := &Int64ObservableGauge{gauge: gauge, callback: callback}
	if callback != nil {
		reg, err := m.Meter.RegisterCallback(inst.observe, gauge)
		if err != nil {
			return nil, err
		}
		inst.registration = reg
	}
	return inst, nil
}

func (m *Meter) createFloat64ObservableGauge(name string, opts []any) (telemetry.Instrument, error) {
	var callback telemetry.Callback[float64]
	for _, opt := range opts {
		if cb, ok := opt.(telemetry.Callback[float64]); ok {
			callback = cb
			break
		}
	}

	gauge, err := m.Meter.Float64ObservableGauge(name)
	if err != nil {
		return nil, err
	}
	inst := &Float64ObservableGauge{gauge: gauge, callback: callback}
	if callback != nil {
		reg, err := m.Meter.RegisterCallback(inst.observe, gauge)
		if err != nil {
			return nil, err
		}
		inst.registration = reg
	}
	return inst, nil
}

// Observable counter instruments

type Int64ObservableCounter struct {
	counter      metric.Int64ObservableCounter
	callback     telemetry.Callback[int64]
	registration metric.Registration
}

func (c *Int64ObservableCounter) Instrument()                    {}
func (c *Int64ObservableCounter) IsMonotonic() bool              { return true }
func (c *Int64ObservableCounter) Precision() telemetry.Precision { return telemetry.PrecisionInt64 }
func (c *Int64ObservableCounter) Unregister() error {
	if c.registration != nil {
		return c.registration.Unregister()
	}
	return nil
}

func (c *Int64ObservableCounter) observe(ctx context.Context, o metric.Observer) error {
	if c.callback != nil {
		observer := &int64Observer{observer: o, instrument: c.counter}
		c.callback(ctx, observer)
	}
	return nil
}

type Int64ObservableUpDownCounter struct {
	counter      metric.Int64ObservableUpDownCounter
	callback     telemetry.Callback[int64]
	registration metric.Registration
}

func (c *Int64ObservableUpDownCounter) Instrument()       {}
func (c *Int64ObservableUpDownCounter) IsMonotonic() bool { return false }
func (c *Int64ObservableUpDownCounter) Precision() telemetry.Precision {
	return telemetry.PrecisionInt64
}
func (c *Int64ObservableUpDownCounter) Unregister() error {
	if c.registration != nil {
		return c.registration.Unregister()
	}
	return nil
}

func (c *Int64ObservableUpDownCounter) observe(ctx context.Context, o metric.Observer) error {
	if c.callback != nil {
		observer := &int64Observer{observer: o, instrument: c.counter}
		c.callback(ctx, observer)
	}
	return nil
}

type Float64ObservableCounter struct {
	counter      metric.Float64ObservableCounter
	callback     telemetry.Callback[float64]
	registration metric.Registration
}

func (c *Float64ObservableCounter) Instrument()                    {}
func (c *Float64ObservableCounter) IsMonotonic() bool              { return true }
func (c *Float64ObservableCounter) Precision() telemetry.Precision { return telemetry.PrecisionFloat64 }
func (c *Float64ObservableCounter) Unregister() error {
	if c.registration != nil {
		return c.registration.Unregister()
	}
	return nil
}

func (c *Float64ObservableCounter) observe(ctx context.Context, o metric.Observer) error {
	if c.callback != nil {
		observer := &float64Observer{observer: o, instrument: c.counter}
		c.callback(ctx, observer)
	}
	return nil
}

type Float64ObservableUpDownCounter struct {
	counter      metric.Float64ObservableUpDownCounter
	callback     telemetry.Callback[float64]
	registration metric.Registration
}

func (c *Float64ObservableUpDownCounter) Instrument()       {}
func (c *Float64ObservableUpDownCounter) IsMonotonic() bool { return false }
func (c *Float64ObservableUpDownCounter) Precision() telemetry.Precision {
	return telemetry.PrecisionFloat64
}
func (c *Float64ObservableUpDownCounter) Unregister() error {
	if c.registration != nil {
		return c.registration.Unregister()
	}
	return nil
}

func (c *Float64ObservableUpDownCounter) observe(ctx context.Context, o metric.Observer) error {
	if c.callback != nil {
		observer := &float64Observer{observer: o, instrument: c.counter}
		c.callback(ctx, observer)
	}
	return nil
}

// Observable gauge instruments

type Int64ObservableGauge struct {
	gauge        metric.Int64ObservableGauge
	callback     telemetry.Callback[int64]
	registration metric.Registration
}

func (g *Int64ObservableGauge) Instrument()                    {}
func (g *Int64ObservableGauge) Precision() telemetry.Precision { return telemetry.PrecisionInt64 }
func (g *Int64ObservableGauge) Unregister() error {
	if g.registration != nil {
		return g.registration.Unregister()
	}
	return nil
}

func (g *Int64ObservableGauge) observe(ctx context.Context, o metric.Observer) error {
	if g.callback != nil {
		observer := &int64Observer{observer: o, instrument: g.gauge}
		g.callback(ctx, observer)
	}
	return nil
}

type Float64ObservableGauge struct {
	gauge        metric.Float64ObservableGauge
	callback     telemetry.Callback[float64]
	registration metric.Registration
}

func (g *Float64ObservableGauge) Instrument()                    {}
func (g *Float64ObservableGauge) Precision() telemetry.Precision { return telemetry.PrecisionFloat64 }
func (g *Float64ObservableGauge) Unregister() error {
	if g.registration != nil {
		return g.registration.Unregister()
	}
	return nil
}

func (g *Float64ObservableGauge) observe(ctx context.Context, o metric.Observer) error {
	if g.callback != nil {
		observer := &float64Observer{observer: o, instrument: g.gauge}
		g.callback(ctx, observer)
	}
	return nil
}

// Observers for async instruments

type int64Observer struct {
	observer   metric.Observer
	instrument metric.Int64Observable
}

func (o *int64Observer) Observe(value int64, attrs ...any) {
	var otelOpts []metric.ObserveOption
	for _, a := range attrs {
		if v, ok := a.(metric.ObserveOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	o.observer.ObserveInt64(o.instrument, value, otelOpts...)
}

type float64Observer struct {
	observer   metric.Observer
	instrument metric.Float64Observable
}

func (o *float64Observer) Observe(value float64, attrs ...any) {
	var otelOpts []metric.ObserveOption
	for _, a := range attrs {
		if v, ok := a.(metric.ObserveOption); ok {
			otelOpts = append(otelOpts, v)
		}
	}
	o.observer.ObserveFloat64(o.instrument, value, otelOpts...)
}

// Ensure types implement interfaces.
var (
	_ telemetry.Meter                      = (*Meter)(nil)
	_ telemetry.Counter[int64]             = (*Int64Counter)(nil)
	_ telemetry.Counter[int64]             = (*Int64UpDownCounter)(nil)
	_ telemetry.Counter[float64]           = (*Float64Counter)(nil)
	_ telemetry.Counter[float64]           = (*Float64UpDownCounter)(nil)
	_ telemetry.Recorder[int64]            = (*Int64Gauge)(nil)
	_ telemetry.Recorder[int64]            = (*Int64Histogram)(nil)
	_ telemetry.Recorder[float64]          = (*Float64Gauge)(nil)
	_ telemetry.Recorder[float64]          = (*Float64Histogram)(nil)
	_ telemetry.ObservableCounter[int64]   = (*Int64ObservableCounter)(nil)
	_ telemetry.ObservableCounter[int64]   = (*Int64ObservableUpDownCounter)(nil)
	_ telemetry.ObservableCounter[float64] = (*Float64ObservableCounter)(nil)
	_ telemetry.ObservableCounter[float64] = (*Float64ObservableUpDownCounter)(nil)
	_ telemetry.ObservableGauge[int64]     = (*Int64ObservableGauge)(nil)
	_ telemetry.ObservableGauge[float64]   = (*Float64ObservableGauge)(nil)
)
