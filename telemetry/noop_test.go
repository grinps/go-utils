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

		// Test counter methods
		if counter.Name() != "" {
			t.Error("noopInstrument.Name() should return empty string")
		}
		if counter.Description() != "" {
			t.Error("noopInstrument.Description() should return empty string")
		}
		if counter.Unit() != "" {
			t.Error("noopInstrument.Unit() should return empty string")
		}
		if !counter.IsMonotonic() {
			t.Error("noopInt64Counter.IsMonotonic() should return true")
		}
		counter.Add(context.Background(), 10) // Should not panic
	})

	t.Run("counter with invalid counter type", func(t *testing.T) {
		inst, err := meter.NewInstrument("counter", InstrumentTypeCounter, CounterTypeUnknown, ErrorHandlingStrategyReturn)
		if err == nil {
			t.Fatal("Expected error with invalid counter type")
		}
		if inst != nil {
			t.Error("Expected nil instrument on error")
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

	t.Run("recorder with invalid aggregation", func(t *testing.T) {
		inst, err := meter.NewInstrument("gauge", InstrumentTypeRecorder, AggregationStrategyUnknown, ErrorHandlingStrategyReturn)
		if err == nil {
			t.Fatal("Expected error with invalid aggregation strategy")
		}
		if inst != nil {
			t.Error("Expected nil instrument on error")
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

func TestNoopInstrument(t *testing.T) {
	inst := &noopInstrument{}

	if inst.Name() != "" {
		t.Errorf("Name() = %q, want empty", inst.Name())
	}
	if inst.Description() != "" {
		t.Errorf("Description() = %q, want empty", inst.Description())
	}
	if inst.Unit() != "" {
		t.Errorf("Unit() = %q, want empty", inst.Unit())
	}
}

func TestNoopInt64Counter(t *testing.T) {
	counter := &noopInt64Counter{}

	t.Run("Add does nothing", func(t *testing.T) {
		counter.Add(context.Background(), 1)
		counter.Add(context.Background(), 100, "attr1", "attr2")
	})

	t.Run("IsMonotonic returns true", func(t *testing.T) {
		if !counter.IsMonotonic() {
			t.Error("IsMonotonic() should return true")
		}
	})

	t.Run("inherits noopInstrument", func(t *testing.T) {
		if counter.Name() != "" {
			t.Error("Name() should return empty string")
		}
	})
}

func TestNoopFloat64Gauge(t *testing.T) {
	gauge := &noopFloat64Gauge{}

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

	t.Run("inherits noopInstrument", func(t *testing.T) {
		if gauge.Name() != "" {
			t.Error("Name() should return empty string")
		}
	})
}
