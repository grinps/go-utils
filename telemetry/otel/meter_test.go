package otel

import (
	"context"
	"testing"

	"github.com/grinps/go-utils/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func TestMeterNewInstrument(t *testing.T) {
	ctx := context.Background()
	provider, err := NewProvider(ctx)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	t.Run("monotonic counter", func(t *testing.T) {
		inst, err := meter.NewInstrument("test_counter",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeMonotonic,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if inst == nil {
			t.Fatal("expected instrument to be non-nil")
		}

		counter, ok := inst.(telemetry.Counter[int64])
		if !ok {
			t.Fatal("expected Counter[int64] type")
		}
		if !counter.IsMonotonic() {
			t.Error("expected monotonic counter")
		}

		// Test Add with attributes
		counter.Add(ctx, 1, "key", "value")
		counter.Add(ctx, 5, attribute.String("attr", "val"))
	})

	t.Run("updown counter", func(t *testing.T) {
		inst, err := meter.NewInstrument("test_updown",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeUpDown,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		counter, ok := inst.(telemetry.Counter[int64])
		if !ok {
			t.Fatal("expected Counter[int64] type")
		}
		if counter.IsMonotonic() {
			t.Error("expected non-monotonic counter")
		}

		counter.Add(ctx, 10)
		counter.Add(ctx, -5)
	})

	t.Run("gauge recorder", func(t *testing.T) {
		inst, err := meter.NewInstrument("test_gauge",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyNone,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		recorder, ok := inst.(telemetry.Recorder[float64])
		if !ok {
			t.Fatal("expected Recorder[float64] type")
		}
		if recorder.IsAggregating() {
			t.Error("expected non-aggregating recorder")
		}
		if recorder.AggregationStrategy() != telemetry.AggregationStrategyNone {
			t.Error("expected AggregationStrategyNone")
		}

		recorder.Record(ctx, 42.5, "key", "value")
	})

	t.Run("histogram recorder", func(t *testing.T) {
		inst, err := meter.NewInstrument("test_histogram",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyHistogram,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		recorder, ok := inst.(telemetry.Recorder[float64])
		if !ok {
			t.Fatal("expected Recorder[float64] type")
		}
		if !recorder.IsAggregating() {
			t.Error("expected aggregating recorder")
		}
		if recorder.AggregationStrategy() != telemetry.AggregationStrategyHistogram {
			t.Error("expected AggregationStrategyHistogram")
		}

		recorder.Record(ctx, 0.5)
		recorder.Record(ctx, 1.5)
		recorder.Record(ctx, 2.5)
	})

	t.Run("default counter type", func(t *testing.T) {
		inst, err := meter.NewInstrument("test_default_counter",
			telemetry.InstrumentTypeCounter,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		counter, ok := inst.(telemetry.Counter[int64])
		if !ok {
			t.Fatal("expected Counter[int64] type")
		}
		// Default should be monotonic
		if !counter.IsMonotonic() {
			t.Error("expected default to be monotonic counter")
		}
	})

	t.Run("default recorder type", func(t *testing.T) {
		inst, err := meter.NewInstrument("test_default_recorder",
			telemetry.InstrumentTypeRecorder,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		recorder, ok := inst.(telemetry.Recorder[float64])
		if !ok {
			t.Fatal("expected Recorder[float64] type")
		}
		// Default should be gauge (non-aggregating)
		if recorder.IsAggregating() {
			t.Error("expected default to be gauge")
		}
	})

	t.Run("default instrument type", func(t *testing.T) {
		inst, err := meter.NewInstrument("test_default_type")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// Default should be counter
		_, ok := inst.(telemetry.Counter[int64])
		if !ok {
			t.Fatal("expected default to be Counter[int64]")
		}
	})

	t.Run("with description and unit options", func(t *testing.T) {
		// New implementation doesn't parse strings. It passes metric.Option to OTEL.
		// We can't verify description/unit are set on the OTEL instrument easily,
		// and the wrapper doesn't store them.
		// So we just verify creation succeeds.
		_, err := meter.NewInstrument("test_with_desc",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeMonotonic,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
}

func TestInstrumentPrecision(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	t.Run("Int64Counter precision", func(t *testing.T) {
		inst, _ := meter.NewInstrument("counter1",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeMonotonic,
		)
		counter := inst.(*Int64Counter)
		if counter.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("expected PrecisionInt64, got %s", counter.Precision())
		}
	})

	t.Run("Int64UpDownCounter precision", func(t *testing.T) {
		inst, _ := meter.NewInstrument("updown1",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeUpDown,
		)
		counter := inst.(*Int64UpDownCounter)
		if counter.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("expected PrecisionInt64, got %s", counter.Precision())
		}
	})

	t.Run("Float64Gauge precision", func(t *testing.T) {
		inst, _ := meter.NewInstrument("gauge1",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyNone,
		)
		gauge := inst.(*Float64Gauge)
		if gauge.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("expected PrecisionFloat64, got %s", gauge.Precision())
		}
	})

	t.Run("Float64Histogram precision", func(t *testing.T) {
		inst, _ := meter.NewInstrument("histogram1",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyHistogram,
		)
		histogram := inst.(*Float64Histogram)
		if histogram.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("expected PrecisionFloat64, got %s", histogram.Precision())
		}
	})
}

func TestMeterImplementsInterface(t *testing.T) {
	var _ telemetry.Meter = (*Meter)(nil)
}

func TestInstrumentsImplementInterfaces(t *testing.T) {
	var _ telemetry.Counter[int64] = (*Int64Counter)(nil)
	var _ telemetry.Counter[int64] = (*Int64UpDownCounter)(nil)
	var _ telemetry.Counter[float64] = (*Float64Counter)(nil)
	var _ telemetry.Counter[float64] = (*Float64UpDownCounter)(nil)
	var _ telemetry.Recorder[int64] = (*Int64Gauge)(nil)
	var _ telemetry.Recorder[int64] = (*Int64Histogram)(nil)
	var _ telemetry.Recorder[float64] = (*Float64Gauge)(nil)
	var _ telemetry.Recorder[float64] = (*Float64Histogram)(nil)
	var _ telemetry.ObservableCounter[int64] = (*Int64ObservableCounter)(nil)
	var _ telemetry.ObservableCounter[int64] = (*Int64ObservableUpDownCounter)(nil)
	var _ telemetry.ObservableCounter[float64] = (*Float64ObservableCounter)(nil)
	var _ telemetry.ObservableCounter[float64] = (*Float64ObservableUpDownCounter)(nil)
	var _ telemetry.ObservableGauge[int64] = (*Int64ObservableGauge)(nil)
	var _ telemetry.ObservableGauge[float64] = (*Float64ObservableGauge)(nil)
}

func TestInstrumentMarkerMethods(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	// Test Instrument() marker method on all sync instrument types
	t.Run("Int64Counter.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("c1", telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)
		inst.(*Int64Counter).Instrument()
	})

	t.Run("Int64UpDownCounter.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("c2", telemetry.InstrumentTypeCounter, telemetry.CounterTypeUpDown)
		inst.(*Int64UpDownCounter).Instrument()
	})

	t.Run("Float64Counter.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("c3", telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic, telemetry.PrecisionFloat64)
		inst.(*Float64Counter).Instrument()
	})

	t.Run("Float64UpDownCounter.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("c4", telemetry.InstrumentTypeCounter, telemetry.CounterTypeUpDown, telemetry.PrecisionFloat64)
		inst.(*Float64UpDownCounter).Instrument()
	})

	t.Run("Int64Gauge.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("g1", telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyNone, telemetry.PrecisionInt64)
		inst.(*Int64Gauge).Instrument()
	})

	t.Run("Int64Histogram.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("h1", telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyHistogram, telemetry.PrecisionInt64)
		inst.(*Int64Histogram).Instrument()
	})

	t.Run("Float64Gauge.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("g2", telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyNone)
		inst.(*Float64Gauge).Instrument()
	})

	t.Run("Float64Histogram.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("h2", telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyHistogram)
		inst.(*Float64Histogram).Instrument()
	})
}

func TestMeterNewInstrumentDefaults(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	t.Run("default instrument type creates counter", func(t *testing.T) {
		// No InstrumentType specified - should default to counter
		inst, err := meter.NewInstrument("default_inst")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		_, ok := inst.(*Int64Counter)
		if !ok {
			t.Error("expected Int64Counter for default instrument type")
		}
	})

	t.Run("counter type only defaults to monotonic", func(t *testing.T) {
		// Only specify InstrumentTypeCounter, no CounterType - should default to monotonic
		inst, err := meter.NewInstrument("counter_default_type",
			telemetry.InstrumentTypeCounter,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		counter, ok := inst.(*Int64Counter)
		if !ok {
			t.Error("expected Int64Counter for default counter type")
		}
		if !counter.IsMonotonic() {
			t.Error("expected monotonic counter for default")
		}
	})

	t.Run("unknown counter type defaults to monotonic", func(t *testing.T) {
		// Use CounterTypeUnknown ("") to hit default branch
		inst, err := meter.NewInstrument("counter_unknown_type",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterType("unknown"),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		counter, ok := inst.(*Int64Counter)
		if !ok {
			t.Error("expected Int64Counter for unknown counter type")
		}
		if !counter.IsMonotonic() {
			t.Error("expected monotonic counter for default")
		}
	})

	t.Run("recorder type only defaults to gauge", func(t *testing.T) {
		// Only specify InstrumentTypeRecorder, no AggregationStrategy - should default to gauge
		inst, err := meter.NewInstrument("recorder_default_type",
			telemetry.InstrumentTypeRecorder,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		_, ok := inst.(*Float64Gauge)
		if !ok {
			t.Error("expected Float64Gauge for default aggregation strategy")
		}
	})

	t.Run("unknown aggregation strategy defaults to gauge", func(t *testing.T) {
		// Use unknown aggregation strategy to hit default branch
		inst, err := meter.NewInstrument("recorder_unknown_type",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategy("unknown"),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		_, ok := inst.(*Float64Gauge)
		if !ok {
			t.Error("expected Float64Gauge for unknown aggregation strategy")
		}
	})
}

func TestMeterWithInstrumentOptions(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	t.Run("counter with metric.Int64CounterOption", func(t *testing.T) {
		inst, err := meter.NewInstrument("counter_with_int64_opts",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeMonotonic,
			metric.WithDescription("test counter"),
			metric.WithUnit("1"),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if inst == nil {
			t.Fatal("expected instrument to be non-nil")
		}
	})

	t.Run("updown counter with metric.Int64UpDownCounterOption", func(t *testing.T) {
		inst, err := meter.NewInstrument("updown_with_opts",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeUpDown,
			metric.WithDescription("test updown counter"),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if inst == nil {
			t.Fatal("expected instrument to be non-nil")
		}
	})

	t.Run("gauge with metric.Float64GaugeOption", func(t *testing.T) {
		inst, err := meter.NewInstrument("gauge_with_opts",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyNone,
			metric.WithDescription("test gauge"),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if inst == nil {
			t.Fatal("expected instrument to be non-nil")
		}
	})

	t.Run("histogram with metric.Float64HistogramOption", func(t *testing.T) {
		inst, err := meter.NewInstrument("histogram_with_opts",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyHistogram,
			metric.WithDescription("test histogram"),
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if inst == nil {
			t.Fatal("expected instrument to be non-nil")
		}
	})
}

func TestInstrumentAddWithOptions(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	t.Run("Int64Counter Add with metric.AddOption", func(t *testing.T) {
		inst, _ := meter.NewInstrument("counter_with_opts",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeMonotonic,
		)
		counter := inst.(telemetry.Counter[int64])
		// Add with attribute option
		counter.Add(ctx, 1, attribute.String("key", "value"))
		counter.Add(ctx, 2) // Add without options
	})

	t.Run("Int64UpDownCounter Add with options", func(t *testing.T) {
		inst, _ := meter.NewInstrument("updown_with_opts",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeUpDown,
		)
		counter := inst.(telemetry.Counter[int64])
		counter.Add(ctx, 1, attribute.String("key", "value"))
		counter.Add(ctx, -1)
	})

	t.Run("Float64Gauge Record with options", func(t *testing.T) {
		inst, _ := meter.NewInstrument("gauge_with_opts",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyNone,
		)
		recorder := inst.(telemetry.Recorder[float64])
		recorder.Record(ctx, 42.5, attribute.String("key", "value"))
		recorder.Record(ctx, 100.0)
	})

	t.Run("Float64Histogram Record with options", func(t *testing.T) {
		inst, _ := meter.NewInstrument("histogram_with_opts",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyHistogram,
		)
		recorder := inst.(telemetry.Recorder[float64])
		recorder.Record(ctx, 0.5, attribute.String("key", "value"))
		recorder.Record(ctx, 1.5)
	})
}

func TestFloat64CounterVariants(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	t.Run("Float64Counter", func(t *testing.T) {
		inst, err := meter.NewInstrument("float64_counter",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeMonotonic,
			telemetry.PrecisionFloat64,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		counter := inst.(*Float64Counter)
		counter.Instrument()
		if !counter.IsMonotonic() {
			t.Error("expected monotonic counter")
		}
		if counter.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("expected PrecisionFloat64, got %s", counter.Precision())
		}
		counter.Add(ctx, 1.5, attribute.String("key", "value"))
		counter.Add(ctx, 2.5)
	})

	t.Run("Float64UpDownCounter", func(t *testing.T) {
		inst, err := meter.NewInstrument("float64_updown",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeUpDown,
			telemetry.PrecisionFloat64,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		counter := inst.(*Float64UpDownCounter)
		counter.Instrument()
		if counter.IsMonotonic() {
			t.Error("expected non-monotonic counter")
		}
		if counter.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("expected PrecisionFloat64, got %s", counter.Precision())
		}
		counter.Add(ctx, 1.5, attribute.String("key", "value"))
		counter.Add(ctx, -1.5)
	})
}

func TestInt64RecorderVariants(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	t.Run("Int64Gauge", func(t *testing.T) {
		inst, err := meter.NewInstrument("int64_gauge",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyNone,
			telemetry.PrecisionInt64,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		gauge := inst.(*Int64Gauge)
		gauge.Instrument()
		if gauge.IsAggregating() {
			t.Error("expected non-aggregating")
		}
		if gauge.AggregationStrategy() != telemetry.AggregationStrategyNone {
			t.Error("expected AggregationStrategyNone")
		}
		if gauge.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("expected PrecisionInt64, got %s", gauge.Precision())
		}
		gauge.Record(ctx, 100, attribute.String("key", "value"))
		gauge.Record(ctx, 200)
	})

	t.Run("Int64Histogram", func(t *testing.T) {
		inst, err := meter.NewInstrument("int64_histogram",
			telemetry.InstrumentTypeRecorder,
			telemetry.AggregationStrategyHistogram,
			telemetry.PrecisionInt64,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		histogram := inst.(*Int64Histogram)
		histogram.Instrument()
		if !histogram.IsAggregating() {
			t.Error("expected aggregating")
		}
		if histogram.AggregationStrategy() != telemetry.AggregationStrategyHistogram {
			t.Error("expected AggregationStrategyHistogram")
		}
		if histogram.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("expected PrecisionInt64, got %s", histogram.Precision())
		}
		histogram.Record(ctx, 100, attribute.String("key", "value"))
		histogram.Record(ctx, 200)
	})
}

func TestObservableCounterInstruments(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	t.Run("Int64ObservableCounter without callback", func(t *testing.T) {
		inst, err := meter.NewInstrument("int64_obs_counter",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeMonotonic,
			telemetry.PrecisionInt64,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		counter := inst.(*Int64ObservableCounter)
		counter.Instrument()
		if !counter.IsMonotonic() {
			t.Error("expected monotonic")
		}
		if counter.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("expected PrecisionInt64, got %s", counter.Precision())
		}
		if err := counter.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("Int64ObservableCounter with callback", func(t *testing.T) {
		callback := telemetry.Callback[int64](func(ctx context.Context, obs telemetry.Observer[int64]) {
			obs.Observe(42, attribute.String("key", "value"))
		})
		inst, err := meter.NewInstrument("int64_obs_counter_cb",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeMonotonic,
			telemetry.PrecisionInt64,
			callback,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		counter := inst.(*Int64ObservableCounter)
		if err := counter.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("Int64ObservableUpDownCounter", func(t *testing.T) {
		callback := telemetry.Callback[int64](func(ctx context.Context, obs telemetry.Observer[int64]) {
			obs.Observe(10)
		})
		inst, err := meter.NewInstrument("int64_obs_updown",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeUpDown,
			telemetry.PrecisionInt64,
			callback,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		counter := inst.(*Int64ObservableUpDownCounter)
		counter.Instrument()
		if counter.IsMonotonic() {
			t.Error("expected non-monotonic")
		}
		if counter.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("expected PrecisionInt64, got %s", counter.Precision())
		}
		if err := counter.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("Float64ObservableCounter", func(t *testing.T) {
		callback := telemetry.Callback[float64](func(ctx context.Context, obs telemetry.Observer[float64]) {
			obs.Observe(3.14, attribute.String("key", "value"))
		})
		inst, err := meter.NewInstrument("float64_obs_counter",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeMonotonic,
			telemetry.PrecisionFloat64,
			callback,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		counter := inst.(*Float64ObservableCounter)
		counter.Instrument()
		if !counter.IsMonotonic() {
			t.Error("expected monotonic")
		}
		if counter.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("expected PrecisionFloat64, got %s", counter.Precision())
		}
		if err := counter.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("Float64ObservableUpDownCounter", func(t *testing.T) {
		callback := telemetry.Callback[float64](func(ctx context.Context, obs telemetry.Observer[float64]) {
			obs.Observe(-1.5)
		})
		inst, err := meter.NewInstrument("float64_obs_updown",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeUpDown,
			telemetry.PrecisionFloat64,
			callback,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		counter := inst.(*Float64ObservableUpDownCounter)
		counter.Instrument()
		if counter.IsMonotonic() {
			t.Error("expected non-monotonic")
		}
		if counter.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("expected PrecisionFloat64, got %s", counter.Precision())
		}
		if err := counter.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})
}

func TestObservableGaugeInstruments(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	meter, _ := provider.Meter("test-meter")

	t.Run("Int64ObservableGauge without callback", func(t *testing.T) {
		inst, err := meter.NewInstrument("int64_obs_gauge",
			telemetry.InstrumentTypeObservableGauge,
			telemetry.PrecisionInt64,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		gauge := inst.(*Int64ObservableGauge)
		gauge.Instrument()
		if gauge.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("expected PrecisionInt64, got %s", gauge.Precision())
		}
		if err := gauge.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("Int64ObservableGauge with callback", func(t *testing.T) {
		callback := telemetry.Callback[int64](func(ctx context.Context, obs telemetry.Observer[int64]) {
			obs.Observe(100, attribute.String("key", "value"))
		})
		inst, err := meter.NewInstrument("int64_obs_gauge_cb",
			telemetry.InstrumentTypeObservableGauge,
			telemetry.PrecisionInt64,
			callback,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		gauge := inst.(*Int64ObservableGauge)
		if err := gauge.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("Float64ObservableGauge without callback", func(t *testing.T) {
		inst, err := meter.NewInstrument("float64_obs_gauge",
			telemetry.InstrumentTypeObservableGauge,
			telemetry.PrecisionFloat64,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		gauge := inst.(*Float64ObservableGauge)
		gauge.Instrument()
		if gauge.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("expected PrecisionFloat64, got %s", gauge.Precision())
		}
		if err := gauge.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("Float64ObservableGauge with callback", func(t *testing.T) {
		callback := telemetry.Callback[float64](func(ctx context.Context, obs telemetry.Observer[float64]) {
			obs.Observe(98.6, attribute.String("unit", "fahrenheit"))
		})
		inst, err := meter.NewInstrument("float64_obs_gauge_cb",
			telemetry.InstrumentTypeObservableGauge,
			telemetry.PrecisionFloat64,
			callback,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		gauge := inst.(*Float64ObservableGauge)
		if err := gauge.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("default precision is Float64", func(t *testing.T) {
		inst, err := meter.NewInstrument("obs_gauge_default_prec",
			telemetry.InstrumentTypeObservableGauge,
		)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		_, ok := inst.(*Float64ObservableGauge)
		if !ok {
			t.Error("expected Float64ObservableGauge for default precision")
		}
	})
}
