package memory

import (
	"context"
	"testing"

	"github.com/grinps/go-utils/telemetry"
)

func TestMeter_NewInstrument_Counter(t *testing.T) {
	p := NewProvider()
	meter, err := p.Meter("test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Create a counter instrument
	inst, err := meter.NewInstrument("test_counter",
		telemetry.InstrumentTypeCounter,
		telemetry.CounterTypeMonotonic,
	)
	if err != nil {
		t.Fatalf("Failed to create instrument: %v", err)
	}
	if inst == nil {
		t.Fatal("Expected non-nil instrument")
	}

	// Verify it implements Counter interface
	counter, ok := inst.(telemetry.Counter[int64])
	if !ok {
		t.Fatal("Expected instrument to implement Counter[int64]")
	}

	// Add a value
	counter.Add(context.Background(), 10)

	// Verify measurement recorded
	m := meter.(*Meter)
	measurements := m.RecordedMeasurements()
	if len(measurements) != 1 {
		t.Fatalf("Expected 1 measurement, got %d", len(measurements))
	}
	if measurements[0].Value != int64(10) {
		t.Errorf("Expected value 10, got %v", measurements[0].Value)
	}
}

func TestMeter_NewInstrument_Recorder(t *testing.T) {
	p := NewProvider()
	meter, err := p.Meter("test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Create a recorder instrument (gauge)
	inst, err := meter.NewInstrument("test_gauge",
		telemetry.InstrumentTypeRecorder,
		telemetry.AggregationStrategyNone,
	)
	if err != nil {
		t.Fatalf("Failed to create instrument: %v", err)
	}
	if inst == nil {
		t.Fatal("Expected non-nil instrument")
	}

	// Verify it implements Recorder interface
	recorder, ok := inst.(telemetry.Recorder[float64])
	if !ok {
		t.Fatal("Expected instrument to implement Recorder[float64]")
	}

	// Record a value
	recorder.Record(context.Background(), 3.14)

	// Verify measurement recorded
	m := meter.(*Meter)
	measurements := m.RecordedMeasurements()
	if len(measurements) != 1 {
		t.Fatalf("Expected 1 measurement, got %d", len(measurements))
	}
	if measurements[0].Value != 3.14 {
		t.Errorf("Expected value 3.14, got %v", measurements[0].Value)
	}
}

func TestMeter_NewInstrument_SameName(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")

	// Create first instrument
	inst1, err := meter.NewInstrument("test_counter", telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)
	if err != nil {
		t.Fatalf("Failed to create first instrument: %v", err)
	}

	// Create second instrument with same name
	inst2, err := meter.NewInstrument("test_counter", telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)
	if err != nil {
		t.Fatalf("Failed to create second instrument: %v", err)
	}

	// Should return the same instance
	if inst1 != inst2 {
		t.Error("Expected same instrument instance for same name")
	}
}

func TestMeter_NewInstrument_UnknownType(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")

	// Create instrument without type returns nil
	inst, err := meter.NewInstrument("test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if inst != nil {
		t.Error("Expected nil instrument for unknown type")
	}
}

func TestMeter_RecordedMeasurements(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")
	m := meter.(*Meter)

	// Initially no measurements
	measurements := m.RecordedMeasurements()
	if len(measurements) != 0 {
		t.Errorf("Expected 0 measurements, got %d", len(measurements))
	}
}

func TestMeter_RecordedMeasurementsByName(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")
	m := meter.(*Meter)

	// Create and use instruments
	counter1, _ := meter.NewInstrument("counter1", telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)
	counter2, _ := meter.NewInstrument("counter2", telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)

	counter1.(telemetry.Counter[int64]).Add(context.Background(), 1)
	counter1.(telemetry.Counter[int64]).Add(context.Background(), 2)
	counter2.(telemetry.Counter[int64]).Add(context.Background(), 10)

	// Filter by name
	measurements1 := m.RecordedMeasurementsByName("counter1")
	if len(measurements1) != 2 {
		t.Errorf("Expected 2 measurements for counter1, got %d", len(measurements1))
	}

	measurements2 := m.RecordedMeasurementsByName("counter2")
	if len(measurements2) != 1 {
		t.Errorf("Expected 1 measurement for counter2, got %d", len(measurements2))
	}
}

func TestMeter_Reset(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")
	m := meter.(*Meter)

	// Add some measurements manually
	m.recordMeasurement(&RecordedMeasurement{
		InstrumentName: "test",
		InstrumentType: "counter",
		Value:          int64(1),
	})

	// Verify measurement exists
	if len(m.RecordedMeasurements()) != 1 {
		t.Fatal("Expected 1 measurement before reset")
	}

	// Reset and verify cleared
	p.Reset()
	if len(m.RecordedMeasurements()) != 0 {
		t.Error("Expected 0 measurements after reset")
	}
}

func TestCounter_IsMonotonic(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")

	// Monotonic counter
	inst1, _ := meter.NewInstrument("monotonic", telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)
	counter1 := inst1.(*Counter[int64])
	if !counter1.IsMonotonic() {
		t.Error("Expected monotonic counter")
	}

	// Up-down counter
	inst2, _ := meter.NewInstrument("updown", telemetry.InstrumentTypeCounter, telemetry.CounterTypeUpDown)
	counter2 := inst2.(*Counter[int64])
	if counter2.IsMonotonic() {
		t.Error("Expected non-monotonic counter")
	}
}

func TestRecorder_AggregationStrategy(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")

	// Gauge (no aggregation)
	inst1, _ := meter.NewInstrument("gauge", telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyNone)
	recorder1 := inst1.(*Recorder[float64])
	if recorder1.IsAggregating() {
		t.Error("Expected gauge to not be aggregating")
	}
	if recorder1.AggregationStrategy() != telemetry.AggregationStrategyNone {
		t.Error("Expected AggregationStrategyNone")
	}

	// Histogram
	inst2, _ := meter.NewInstrument("histogram", telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyHistogram)
	recorder2 := inst2.(*Recorder[float64])
	if !recorder2.IsAggregating() {
		t.Error("Expected histogram to be aggregating")
	}
	if recorder2.AggregationStrategy() != telemetry.AggregationStrategyHistogram {
		t.Error("Expected AggregationStrategyHistogram")
	}
}

func TestParseAttributes_KeyValuePairs(t *testing.T) {
	// Test key-value pairs
	attrs := parseAttributes([]any{"key1", "value1", "key2", 42})
	if len(attrs) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(attrs))
	}
	if attrs[0].Key != "key1" || attrs[0].Value != "value1" {
		t.Errorf("First attribute mismatch: %v", attrs[0])
	}
	if attrs[1].Key != "key2" || attrs[1].Value != 42 {
		t.Errorf("Second attribute mismatch: %v", attrs[1])
	}

	// Test mixed Attribute and key-value pairs
	attrs = parseAttributes([]any{String("attr1", "val1"), "key2", 100})
	if len(attrs) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(attrs))
	}
	if attrs[0].Key != "attr1" {
		t.Error("First attribute should be 'attr1'")
	}
	if attrs[1].Key != "key2" || attrs[1].Value != 100 {
		t.Errorf("Second attribute mismatch: %v", attrs[1])
	}

	// Test string key followed by Attribute (should not consume Attribute)
	attrs = parseAttributes([]any{"orphan", String("real", "attr")})
	if len(attrs) != 1 {
		t.Fatalf("Expected 1 attribute, got %d", len(attrs))
	}
	if attrs[0].Key != "real" {
		t.Error("Should have 'real' attribute, not orphan key")
	}
}

func TestCounter_KeyValueAttributes(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")

	inst, _ := meter.NewInstrument("test_counter",
		telemetry.InstrumentTypeCounter,
		telemetry.CounterTypeMonotonic,
	)
	counter := inst.(telemetry.Counter[int64])

	// Add with key-value pairs
	counter.Add(context.Background(), 10, "user.id", "12345", "count", 5)

	m := meter.(*Meter)
	measurements := m.RecordedMeasurements()
	if len(measurements) != 1 {
		t.Fatalf("Expected 1 measurement, got %d", len(measurements))
	}

	attrs := measurements[0].Attributes
	if len(attrs) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(attrs))
	}
	if attrs[0].Key != "user.id" || attrs[0].Value != "12345" {
		t.Errorf("First attribute mismatch: %v", attrs[0])
	}
	if attrs[1].Key != "count" || attrs[1].Value != 5 {
		t.Errorf("Second attribute mismatch: %v", attrs[1])
	}
}

func TestInstrumentMarkerMethods(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")

	t.Run("Counter.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("c1", telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)
		counter := inst.(*Counter[int64])
		counter.Instrument() // Should not panic
	})

	t.Run("Recorder.Instrument", func(t *testing.T) {
		inst, _ := meter.NewInstrument("r1", telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyNone)
		recorder := inst.(*Recorder[float64])
		recorder.Instrument() // Should not panic
	})
}

func TestBaseInstrumentUnit(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")

	inst, _ := meter.NewInstrument("test_with_unit",
		telemetry.InstrumentTypeCounter,
		telemetry.CounterTypeMonotonic,
		"Test description",
		"requests",
	)
	counter := inst.(*Counter[int64])
	if counter.Unit() != "requests" {
		t.Errorf("Expected unit 'requests', got '%s'", counter.Unit())
	}
}

func TestRecorderPrecision(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")

	t.Run("Float64 precision (default)", func(t *testing.T) {
		inst, _ := meter.NewInstrument("gauge1", telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyNone)
		recorder := inst.(*Recorder[float64])
		if recorder.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("Expected PrecisionFloat64, got %s", recorder.Precision())
		}
	})

	t.Run("Int64 precision", func(t *testing.T) {
		inst, _ := meter.NewInstrument("gauge2", telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyNone, telemetry.PrecisionInt64)
		recorder := inst.(*Recorder[int64])
		if recorder.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("Expected PrecisionInt64, got %s", recorder.Precision())
		}
	})
}

func TestObservableCounter(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")
	ctx := context.Background()

	t.Run("Int64 observable counter with callback", func(t *testing.T) {
		callCount := 0
		callback := telemetry.Callback[int64](func(ctx context.Context, obs telemetry.Observer[int64]) {
			callCount++
			obs.Observe(42, "key", "value")
		})
		inst, err := meter.NewInstrument("obs_counter1",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeMonotonic,
			telemetry.PrecisionInt64,
			callback,
		)
		if err != nil {
			t.Fatalf("Failed to create observable counter: %v", err)
		}

		obsCounter := inst.(*ObservableCounter[int64])
		obsCounter.Instrument()

		if !obsCounter.IsMonotonic() {
			t.Error("Expected monotonic counter")
		}
		if obsCounter.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("Expected PrecisionInt64, got %s", obsCounter.Precision())
		}

		// Collect to trigger callback
		obsCounter.Collect(ctx)
		if callCount != 1 {
			t.Errorf("Expected callback to be called once, got %d", callCount)
		}

		// Unregister
		if err := obsCounter.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("Int64 observable updown counter", func(t *testing.T) {
		callback := telemetry.Callback[int64](func(ctx context.Context, obs telemetry.Observer[int64]) {
			obs.Observe(-5)
		})
		inst, _ := meter.NewInstrument("obs_updown1",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeUpDown,
			telemetry.PrecisionInt64,
			callback,
		)
		obsCounter := inst.(*ObservableCounter[int64])
		if obsCounter.IsMonotonic() {
			t.Error("Expected non-monotonic counter")
		}
		obsCounter.Collect(ctx)
		obsCounter.Unregister()
	})

	t.Run("Float64 observable counter", func(t *testing.T) {
		callback := telemetry.Callback[float64](func(ctx context.Context, obs telemetry.Observer[float64]) {
			obs.Observe(3.14)
		})
		inst, _ := meter.NewInstrument("obs_counter2",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeMonotonic,
			telemetry.PrecisionFloat64,
			callback,
		)
		obsCounter := inst.(*ObservableCounter[float64])
		if !obsCounter.IsMonotonic() {
			t.Error("Expected monotonic counter")
		}
		if obsCounter.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("Expected PrecisionFloat64, got %s", obsCounter.Precision())
		}
		obsCounter.Collect(ctx)
		obsCounter.Unregister()
	})

	t.Run("Float64 observable updown counter", func(t *testing.T) {
		callback := telemetry.Callback[float64](func(ctx context.Context, obs telemetry.Observer[float64]) {
			obs.Observe(-1.5)
		})
		inst, _ := meter.NewInstrument("obs_updown2",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeUpDown,
			telemetry.PrecisionFloat64,
			callback,
		)
		obsCounter := inst.(*ObservableCounter[float64])
		if obsCounter.IsMonotonic() {
			t.Error("Expected non-monotonic counter")
		}
		obsCounter.Collect(ctx)
		obsCounter.Unregister()
	})

	t.Run("Observable counter without callback", func(t *testing.T) {
		inst, _ := meter.NewInstrument("obs_counter_no_cb",
			telemetry.InstrumentTypeObservableCounter,
			telemetry.CounterTypeMonotonic,
			telemetry.PrecisionInt64,
		)
		obsCounter := inst.(*ObservableCounter[int64])
		obsCounter.Collect(ctx) // Should not panic
		obsCounter.Unregister()
	})
}

func TestObservableGauge(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")
	ctx := context.Background()

	t.Run("Int64 observable gauge with callback", func(t *testing.T) {
		callCount := 0
		callback := telemetry.Callback[int64](func(ctx context.Context, obs telemetry.Observer[int64]) {
			callCount++
			obs.Observe(100, "unit", "celsius")
		})
		inst, err := meter.NewInstrument("obs_gauge1",
			telemetry.InstrumentTypeObservableGauge,
			telemetry.PrecisionInt64,
			callback,
		)
		if err != nil {
			t.Fatalf("Failed to create observable gauge: %v", err)
		}

		obsGauge := inst.(*ObservableGauge[int64])
		obsGauge.Instrument()

		if obsGauge.Precision() != telemetry.PrecisionInt64 {
			t.Errorf("Expected PrecisionInt64, got %s", obsGauge.Precision())
		}

		// Collect to trigger callback
		obsGauge.Collect(ctx)
		if callCount != 1 {
			t.Errorf("Expected callback to be called once, got %d", callCount)
		}

		// Unregister
		if err := obsGauge.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("Float64 observable gauge", func(t *testing.T) {
		callback := telemetry.Callback[float64](func(ctx context.Context, obs telemetry.Observer[float64]) {
			obs.Observe(98.6)
		})
		inst, _ := meter.NewInstrument("obs_gauge2",
			telemetry.InstrumentTypeObservableGauge,
			telemetry.PrecisionFloat64,
			callback,
		)
		obsGauge := inst.(*ObservableGauge[float64])
		if obsGauge.Precision() != telemetry.PrecisionFloat64 {
			t.Errorf("Expected PrecisionFloat64, got %s", obsGauge.Precision())
		}
		obsGauge.Collect(ctx)
		obsGauge.Unregister()
	})

	t.Run("Observable gauge without callback", func(t *testing.T) {
		inst, _ := meter.NewInstrument("obs_gauge_no_cb",
			telemetry.InstrumentTypeObservableGauge,
			telemetry.PrecisionFloat64,
		)
		obsGauge := inst.(*ObservableGauge[float64])
		obsGauge.Collect(ctx) // Should not panic
		obsGauge.Unregister()
	})

	t.Run("Default precision is Float64", func(t *testing.T) {
		inst, _ := meter.NewInstrument("obs_gauge_default",
			telemetry.InstrumentTypeObservableGauge,
		)
		_, ok := inst.(*ObservableGauge[float64])
		if !ok {
			t.Error("Expected Float64 observable gauge by default")
		}
	})
}
