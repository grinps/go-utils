package telemetry

import (
	"context"
	"testing"
)

func TestNoopProvider_Tracer(t *testing.T) {
	np := &NoopProvider{}

	t.Run("returns tracer successfully", func(t *testing.T) {
		tracer, err := np.Tracer("test-tracer")
		if err != nil {
			t.Fatalf("Tracer() error = %v", err)
		}
		if tracer == nil {
			t.Fatal("Tracer() returned nil")
		}
	})

	t.Run("with ErrorHandlingStrategyIgnore", func(t *testing.T) {
		tracer, err := np.Tracer("test", ErrorHandlingStrategyIgnore)
		if err != nil {
			t.Fatalf("Tracer() error = %v", err)
		}
		if tracer == nil {
			t.Fatal("Tracer() returned nil")
		}
	})

	t.Run("with ErrorHandlingStrategyReturn", func(t *testing.T) {
		tracer, err := np.Tracer("test", ErrorHandlingStrategyReturn)
		if err != nil {
			t.Fatalf("Tracer() error = %v", err)
		}
		if tracer == nil {
			t.Fatal("Tracer() returned nil")
		}
	})

	t.Run("with ErrorHandlingStrategyGenerateError", func(t *testing.T) {
		tracer, err := np.Tracer("test", ErrorHandlingStrategyGenerateError)
		if err == nil {
			t.Fatal("Expected error with ErrorHandlingStrategyGenerateError")
		}
		if tracer != nil {
			t.Error("Expected nil tracer on error")
		}
	})

	t.Run("with unknown option type", func(t *testing.T) {
		tracer, err := np.Tracer("test", 123, "unknown")
		if err != nil {
			t.Fatalf("Tracer() error = %v", err)
		}
		if tracer == nil {
			t.Fatal("Tracer() returned nil")
		}
	})
}

func TestNoopProvider_Meter(t *testing.T) {
	np := &NoopProvider{}

	t.Run("returns meter successfully", func(t *testing.T) {
		meter, err := np.Meter("test-meter")
		if err != nil {
			t.Fatalf("Meter() error = %v", err)
		}
		if meter == nil {
			t.Fatal("Meter() returned nil")
		}
	})

	t.Run("with ErrorHandlingStrategyIgnore", func(t *testing.T) {
		meter, err := np.Meter("test", ErrorHandlingStrategyIgnore)
		if err != nil {
			t.Fatalf("Meter() error = %v", err)
		}
		if meter == nil {
			t.Fatal("Meter() returned nil")
		}
	})

	t.Run("with ErrorHandlingStrategyReturn", func(t *testing.T) {
		meter, err := np.Meter("test", ErrorHandlingStrategyReturn)
		if err != nil {
			t.Fatalf("Meter() error = %v", err)
		}
		if meter == nil {
			t.Fatal("Meter() returned nil")
		}
	})

	t.Run("with ErrorHandlingStrategyGenerateError", func(t *testing.T) {
		meter, err := np.Meter("test", ErrorHandlingStrategyGenerateError)
		if err == nil {
			t.Fatal("Expected error with ErrorHandlingStrategyGenerateError")
		}
		if meter != nil {
			t.Error("Expected nil meter on error")
		}
	})
}

func TestNoopProvider_Shutdown(t *testing.T) {
	np := &NoopProvider{}
	err := np.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}
}

func TestNoopTracer_Start(t *testing.T) {
	np := &NoopProvider{}
	tracer, _ := np.Tracer("test")

	ctx := context.Background()
	newCtx, span := tracer.Start(ctx, "test-span")

	if newCtx == nil {
		t.Fatal("Start() returned nil context")
	}
	if span == nil {
		t.Fatal("Start() returned nil span")
	}

	// Verify it returns the same context (noop behavior)
	if newCtx != ctx {
		t.Error("NoopTracer.Start() should return the same context")
	}
}

func TestNoopSpan(t *testing.T) {
	np := &NoopProvider{}
	tracer, _ := np.Tracer("test")
	_, span := tracer.Start(context.Background(), "test-span")

	t.Run("End does nothing", func(t *testing.T) {
		span.End() // Should not panic
		span.End("option1", "option2")
	})

	t.Run("IsRecording returns false", func(t *testing.T) {
		if span.IsRecording() {
			t.Error("NoopSpan.IsRecording() should return false")
		}
	})

	t.Run("SetAttributes does nothing", func(t *testing.T) {
		span.SetAttributes("key", "value") // Should not panic
	})

	t.Run("AddEvent does nothing", func(t *testing.T) {
		span.AddEvent("event-name")              // Should not panic
		span.AddEvent("event-name", "opt1", 123) // Should not panic
	})

	t.Run("RecordError does nothing", func(t *testing.T) {
		span.RecordError(nil)                                    // Should not panic
		span.RecordError(context.DeadlineExceeded)               // Should not panic
		span.RecordError(context.Canceled, "option1", "option2") // Should not panic
	})

	t.Run("SetStatus does nothing", func(t *testing.T) {
		span.SetStatus(0, "")               // Should not panic
		span.SetStatus(1, "error occurred") // Should not panic
	})

	t.Run("SetName does nothing", func(t *testing.T) {
		span.SetName("new-name") // Should not panic
	})

	t.Run("TracerProvider returns nil", func(t *testing.T) {
		provider := span.TracerProvider()
		if provider != nil {
			t.Error("NoopSpan.TracerProvider() should return nil")
		}
	})
}

func TestNoopMeter_NewInstrument(t *testing.T) {
	np := &NoopProvider{}
	meter, _ := np.Meter("test")

	t.Run("with no options returns nil instrument", func(t *testing.T) {
		inst, err := meter.NewInstrument("test-instrument")
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		// With no valid instrument type, returns nil instrument
		if inst != nil {
			t.Error("Expected nil instrument with no options")
		}
	})

	t.Run("with nil option skipped", func(t *testing.T) {
		inst, err := meter.NewInstrument("test", nil)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		if inst != nil {
			t.Error("Expected nil instrument")
		}
	})

	t.Run("with string option sets name", func(t *testing.T) {
		inst, err := meter.NewInstrument("original", "new-name")
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		// Still nil because no valid instrument type
		if inst != nil {
			t.Error("Expected nil instrument without valid type")
		}
	})

	t.Run("counter with monotonic type", func(t *testing.T) {
		inst, err := meter.NewInstrument("counter", InstrumentTypeCounter, CounterTypeMonotonic)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		if inst == nil {
			t.Fatal("Expected non-nil instrument")
		}

		counter, ok := inst.(Counter[int64])
		if !ok {
			t.Fatalf("Expected Counter[int64], got %T", inst)
		}

		if !counter.IsMonotonic() {
			t.Error("noopInt64Counter.IsMonotonic() should return true")
		}
		counter.Add(context.Background(), 10) // Should not panic
	})

	t.Run("counter with unknown counter type defaults to monotonic", func(t *testing.T) {
		inst, err := meter.NewInstrument("counter_default", InstrumentTypeCounter, CounterTypeUnknown)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		if inst == nil {
			t.Fatal("Expected non-nil instrument")
		}
		counter, ok := inst.(Counter[int64])
		if !ok {
			t.Fatalf("Expected Counter[int64], got %T", inst)
		}
		if !counter.IsMonotonic() {
			t.Error("Default counter should be monotonic")
		}
	})

	t.Run("recorder with none aggregation", func(t *testing.T) {
		inst, err := meter.NewInstrument("gauge", InstrumentTypeRecorder, AggregationStrategyNone)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		if inst == nil {
			t.Fatal("Expected non-nil instrument")
		}

		recorder, ok := inst.(Recorder[float64])
		if !ok {
			t.Fatalf("Expected Recorder[float64], got %T", inst)
		}

		if recorder.IsAggregating() {
			t.Error("noopFloat64Gauge.IsAggregating() should return false")
		}
		if recorder.AggregationStrategy() != AggregationStrategyNone {
			t.Error("noopFloat64Gauge.AggregationStrategy() should return AggregationStrategyNone")
		}
		recorder.Record(context.Background(), 1.5) // Should not panic
	})

	t.Run("recorder with unknown aggregation defaults to gauge", func(t *testing.T) {
		inst, err := meter.NewInstrument("gauge_default", InstrumentTypeRecorder, AggregationStrategyUnknown)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		if inst == nil {
			t.Fatal("Expected non-nil instrument")
		}
		recorder, ok := inst.(Recorder[float64])
		if !ok {
			t.Fatalf("Expected Recorder[float64], got %T", inst)
		}
		if recorder.IsAggregating() {
			t.Error("Default recorder should not be aggregating (gauge)")
		}
	})

	t.Run("unknown instrument type", func(t *testing.T) {
		inst, err := meter.NewInstrument("unknown", InstrumentTypeUnknown, ErrorHandlingStrategyReturn)
		if err == nil {
			t.Fatal("Expected error with unknown instrument type")
		}
		if inst != nil {
			t.Error("Expected nil instrument on error")
		}
	})

	t.Run("invalid option type with ErrorHandlingStrategyReturn", func(t *testing.T) {
		inst, err := meter.NewInstrument("test", struct{}{}, ErrorHandlingStrategyReturn)
		if err == nil {
			t.Fatal("Expected error with invalid option type")
		}
		if inst != nil {
			t.Error("Expected nil instrument on error")
		}
	})

	t.Run("invalid option type with ErrorHandlingStrategyIgnore", func(t *testing.T) {
		inst, err := meter.NewInstrument("test", struct{}{}, ErrorHandlingStrategyIgnore)
		if err != nil {
			t.Errorf("Should not return error with ignore strategy: %v", err)
		}
		// Still nil because no valid instrument type
		if inst != nil {
			t.Error("Expected nil instrument without valid type")
		}
	})
}

func TestNoopInt64Counter(t *testing.T) {
	counter := &noopInt64Counter{}

	t.Run("Instrument marker method", func(t *testing.T) {
		counter.Instrument() // Should not panic
	})

	t.Run("Add does nothing", func(t *testing.T) {
		counter.Add(context.Background(), 1)
		counter.Add(context.Background(), 100, "attr1", "attr2")
	})

	t.Run("IsMonotonic returns true", func(t *testing.T) {
		if !counter.IsMonotonic() {
			t.Error("IsMonotonic() should return true")
		}
	})

	t.Run("Precision returns PrecisionInt64", func(t *testing.T) {
		if counter.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", counter.Precision(), PrecisionInt64)
		}
	})
}

func TestNoopFloat64Gauge(t *testing.T) {
	gauge := &noopFloat64Gauge{}

	t.Run("Instrument marker method", func(t *testing.T) {
		gauge.Instrument() // Should not panic
	})

	t.Run("Record does nothing", func(t *testing.T) {
		gauge.Record(context.Background(), 1.5)
		gauge.Record(context.Background(), 100.0, "attr1", "attr2")
	})

	t.Run("IsAggregating returns false", func(t *testing.T) {
		if gauge.IsAggregating() {
			t.Error("IsAggregating() should return false")
		}
	})

	t.Run("AggregationStrategy returns None", func(t *testing.T) {
		if gauge.AggregationStrategy() != AggregationStrategyNone {
			t.Errorf("AggregationStrategy() = %v, want %v", gauge.AggregationStrategy(), AggregationStrategyNone)
		}
	})

	t.Run("Precision returns PrecisionFloat64", func(t *testing.T) {
		if gauge.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", gauge.Precision(), PrecisionFloat64)
		}
	})
}

func TestNoopObservableCounter(t *testing.T) {
	provider := &NoopProvider{}
	meter, _ := provider.Meter("test")

	t.Run("observable counter int64", func(t *testing.T) {
		inst, err := meter.NewInstrument("obs_counter",
			InstrumentTypeObservableCounter,
			CounterTypeMonotonic,
			PrecisionInt64,
		)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		counter, ok := inst.(ObservableCounter[int64])
		if !ok {
			t.Fatalf("Expected ObservableCounter[int64], got %T", inst)
		}
		if !counter.IsMonotonic() {
			t.Error("IsMonotonic() should return true")
		}
		if counter.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", counter.Precision(), PrecisionInt64)
		}
		if err := counter.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("observable updown counter float64", func(t *testing.T) {
		inst, err := meter.NewInstrument("obs_updown",
			InstrumentTypeObservableCounter,
			CounterTypeUpDown,
			PrecisionFloat64,
		)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		counter, ok := inst.(ObservableCounter[float64])
		if !ok {
			t.Fatalf("Expected ObservableCounter[float64], got %T", inst)
		}
		if counter.IsMonotonic() {
			t.Error("IsMonotonic() should return false for updown counter")
		}
		if counter.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", counter.Precision(), PrecisionFloat64)
		}
	})
}

func TestNoopObservableGauge(t *testing.T) {
	provider := &NoopProvider{}
	meter, _ := provider.Meter("test")

	t.Run("observable gauge float64", func(t *testing.T) {
		inst, err := meter.NewInstrument("obs_gauge",
			InstrumentTypeObservableGauge,
			PrecisionFloat64,
		)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		gauge, ok := inst.(ObservableGauge[float64])
		if !ok {
			t.Fatalf("Expected ObservableGauge[float64], got %T", inst)
		}
		if gauge.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", gauge.Precision(), PrecisionFloat64)
		}
		if err := gauge.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
	})

	t.Run("observable gauge int64", func(t *testing.T) {
		inst, err := meter.NewInstrument("obs_gauge_int",
			InstrumentTypeObservableGauge,
			PrecisionInt64,
		)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		gauge, ok := inst.(ObservableGauge[int64])
		if !ok {
			t.Fatalf("Expected ObservableGauge[int64], got %T", inst)
		}
		if gauge.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", gauge.Precision(), PrecisionInt64)
		}
	})
}

func TestNoopObservableInstrumentMethods(t *testing.T) {
	// Test all noop observable counter variants
	t.Run("noopObservableInt64Counter", func(t *testing.T) {
		c := &noopObservableInt64Counter{}
		c.Instrument()
		if err := c.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
		if !c.IsMonotonic() {
			t.Error("IsMonotonic() should return true")
		}
		if c.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", c.Precision(), PrecisionInt64)
		}
	})

	t.Run("noopObservableInt64UpDownCounter", func(t *testing.T) {
		c := &noopObservableInt64UpDownCounter{}
		c.Instrument()
		if err := c.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
		if c.IsMonotonic() {
			t.Error("IsMonotonic() should return false")
		}
		if c.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", c.Precision(), PrecisionInt64)
		}
	})

	t.Run("noopObservableFloat64Counter", func(t *testing.T) {
		c := &noopObservableFloat64Counter{}
		c.Instrument()
		if err := c.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
		if !c.IsMonotonic() {
			t.Error("IsMonotonic() should return true")
		}
		if c.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", c.Precision(), PrecisionFloat64)
		}
	})

	t.Run("noopObservableFloat64UpDownCounter", func(t *testing.T) {
		c := &noopObservableFloat64UpDownCounter{}
		c.Instrument()
		if err := c.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
		if c.IsMonotonic() {
			t.Error("IsMonotonic() should return false")
		}
		if c.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", c.Precision(), PrecisionFloat64)
		}
	})

	// Test all noop observable gauge variants
	t.Run("noopObservableInt64Gauge", func(t *testing.T) {
		g := &noopObservableInt64Gauge{}
		g.Instrument()
		if err := g.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
		if g.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", g.Precision(), PrecisionInt64)
		}
	})

	t.Run("noopObservableFloat64Gauge", func(t *testing.T) {
		g := &noopObservableFloat64Gauge{}
		g.Instrument()
		if err := g.Unregister(); err != nil {
			t.Errorf("Unregister() error = %v", err)
		}
		if g.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", g.Precision(), PrecisionFloat64)
		}
	})
}

func TestNoopSpanMethods(t *testing.T) {
	span := &noopSpan{}

	// Test all span methods for coverage
	span.End("opt1", 123)
	span.SetAttributes("key", "value", "key2", 123)
	span.AddEvent("event", "opt1")
	span.RecordError(nil, "opt1")
	span.SetStatus(1, "error")
	span.SetName("new-name")

	if span.IsRecording() {
		t.Error("IsRecording() should return false")
	}
	if span.TracerProvider() != nil {
		t.Error("TracerProvider() should return nil")
	}
}

func TestNoopCounterVariants(t *testing.T) {
	ctx := context.Background()

	t.Run("noopInt64Counter", func(t *testing.T) {
		c := &noopInt64Counter{}
		c.Instrument()
		c.Add(ctx, 1)
		c.Add(ctx, 100, "attr1", "attr2")
		if !c.IsMonotonic() {
			t.Error("IsMonotonic() should return true")
		}
		if c.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", c.Precision(), PrecisionInt64)
		}
	})

	t.Run("noopInt64UpDownCounter", func(t *testing.T) {
		c := &noopInt64UpDownCounter{}
		c.Instrument()
		c.Add(ctx, 1)
		c.Add(ctx, -1, "attr")
		if c.IsMonotonic() {
			t.Error("IsMonotonic() should return false")
		}
		if c.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", c.Precision(), PrecisionInt64)
		}
	})

	t.Run("noopFloat64Counter", func(t *testing.T) {
		c := &noopFloat64Counter{}
		c.Instrument()
		c.Add(ctx, 1.5)
		c.Add(ctx, 2.5, "attr")
		if !c.IsMonotonic() {
			t.Error("IsMonotonic() should return true")
		}
		if c.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", c.Precision(), PrecisionFloat64)
		}
	})

	t.Run("noopFloat64UpDownCounter", func(t *testing.T) {
		c := &noopFloat64UpDownCounter{}
		c.Instrument()
		c.Add(ctx, 1.5)
		c.Add(ctx, -1.5, "attr")
		if c.IsMonotonic() {
			t.Error("IsMonotonic() should return false")
		}
		if c.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", c.Precision(), PrecisionFloat64)
		}
	})
}

func TestNoopRecorderVariants(t *testing.T) {
	ctx := context.Background()

	t.Run("noopInt64Gauge", func(t *testing.T) {
		g := &noopInt64Gauge{}
		g.Instrument()
		g.Record(ctx, 100)
		g.Record(ctx, 200, "attr")
		if g.IsAggregating() {
			t.Error("IsAggregating() should return false")
		}
		if g.AggregationStrategy() != AggregationStrategyNone {
			t.Errorf("AggregationStrategy() = %v, want %v", g.AggregationStrategy(), AggregationStrategyNone)
		}
		if g.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", g.Precision(), PrecisionInt64)
		}
	})

	t.Run("noopInt64Histogram", func(t *testing.T) {
		h := &noopInt64Histogram{}
		h.Instrument()
		h.Record(ctx, 100)
		h.Record(ctx, 200, "attr")
		if !h.IsAggregating() {
			t.Error("IsAggregating() should return true")
		}
		if h.AggregationStrategy() != AggregationStrategyHistogram {
			t.Errorf("AggregationStrategy() = %v, want %v", h.AggregationStrategy(), AggregationStrategyHistogram)
		}
		if h.Precision() != PrecisionInt64 {
			t.Errorf("Precision() = %v, want %v", h.Precision(), PrecisionInt64)
		}
	})

	t.Run("noopFloat64Gauge", func(t *testing.T) {
		g := &noopFloat64Gauge{}
		g.Instrument()
		g.Record(ctx, 1.5)
		g.Record(ctx, 2.5, "attr")
		if g.IsAggregating() {
			t.Error("IsAggregating() should return false")
		}
		if g.AggregationStrategy() != AggregationStrategyNone {
			t.Errorf("AggregationStrategy() = %v, want %v", g.AggregationStrategy(), AggregationStrategyNone)
		}
		if g.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", g.Precision(), PrecisionFloat64)
		}
	})

	t.Run("noopFloat64Histogram", func(t *testing.T) {
		h := &noopFloat64Histogram{}
		h.Instrument()
		h.Record(ctx, 1.5)
		h.Record(ctx, 2.5, "attr")
		if !h.IsAggregating() {
			t.Error("IsAggregating() should return true")
		}
		if h.AggregationStrategy() != AggregationStrategyHistogram {
			t.Errorf("AggregationStrategy() = %v, want %v", h.AggregationStrategy(), AggregationStrategyHistogram)
		}
		if h.Precision() != PrecisionFloat64 {
			t.Errorf("Precision() = %v, want %v", h.Precision(), PrecisionFloat64)
		}
	})
}
