package telemetry

import (
	"context"
	"sync"
	"testing"
)

func TestDefault(t *testing.T) {
	// Default should return the NoopProvider
	provider := Default()
	if provider == nil {
		t.Fatal("Default() returned nil")
	}

	// Verify it's the NoopProvider
	_, ok := provider.(*NoopProvider)
	if !ok {
		t.Errorf("Default() returned %T, expected *NoopProvider", provider)
	}
}

func TestAsDefault(t *testing.T) {
	// Save original default
	originalDefault := Default()
	defer AsDefault(originalDefault)

	t.Run("set custom provider", func(t *testing.T) {
		customProvider := &NoopProvider{}
		AsDefault(customProvider)

		got := Default()
		if got != customProvider {
			t.Errorf("Default() = %v, want %v", got, customProvider)
		}
	})

	t.Run("nil provider ignored", func(t *testing.T) {
		currentProvider := Default()
		AsDefault(nil)

		got := Default()
		if got != currentProvider {
			t.Error("AsDefault(nil) should not change the default provider")
		}
	})
}

func TestAsDefault_Concurrent(t *testing.T) {
	originalDefault := Default()
	defer AsDefault(originalDefault)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			AsDefault(&NoopProvider{})
		}()
		go func() {
			defer wg.Done()
			_ = Default()
		}()
	}
	wg.Wait()
}

func TestContextWithTelemetry(t *testing.T) {
	t.Run("with valid context", func(t *testing.T) {
		provider := &NoopProvider{}
		ctx := context.Background()

		newCtx := ContextWithTelemetry(ctx, provider)
		if newCtx == nil {
			t.Fatal("ContextWithTelemetry returned nil context")
		}

		// Verify provider is stored
		got := ContextTelemetry(newCtx, true)
		if got != provider {
			t.Errorf("ContextTelemetry() = %v, want %v", got, provider)
		}
	})

	t.Run("with nil context", func(t *testing.T) {
		provider := &NoopProvider{}
		newCtx := ContextWithTelemetry(nil, provider) //nolint:staticcheck // intentionally testing nil context handling
		if newCtx == nil {
			t.Fatal("ContextWithTelemetry(nil, provider) returned nil")
		}

		// Should still store the provider
		got := ContextTelemetry(newCtx, true)
		if got != provider {
			t.Errorf("ContextTelemetry() = %v, want %v", got, provider)
		}
	})
}

func TestContextTelemetry(t *testing.T) {
	t.Run("provider in context with default fallback", func(t *testing.T) {
		provider := &NoopProvider{}
		ctx := ContextWithTelemetry(context.Background(), provider)

		got := ContextTelemetry(ctx, true)
		if got != provider {
			t.Errorf("ContextTelemetry() = %v, want %v", got, provider)
		}
	})

	t.Run("provider in context without default fallback", func(t *testing.T) {
		provider := &NoopProvider{}
		ctx := ContextWithTelemetry(context.Background(), provider)

		got := ContextTelemetry(ctx, false)
		if got != provider {
			t.Errorf("ContextTelemetry() = %v, want %v", got, provider)
		}
	})

	t.Run("nil context with default fallback returns default", func(t *testing.T) {
		got := ContextTelemetry(nil, true) //nolint:staticcheck // intentionally testing nil context handling
		defaultProvider := Default()
		if got != defaultProvider {
			t.Errorf("ContextTelemetry(nil, true) = %v, want default %v", got, defaultProvider)
		}
	})

	t.Run("nil context without default fallback returns nil", func(t *testing.T) {
		got := ContextTelemetry(nil, false) //nolint:staticcheck // intentionally testing nil context handling
		if got != nil {
			t.Errorf("ContextTelemetry(nil, false) = %v, want nil", got)
		}
	})

	t.Run("no provider in context with default fallback returns default", func(t *testing.T) {
		ctx := context.Background()
		got := ContextTelemetry(ctx, true)
		defaultProvider := Default()
		if got != defaultProvider {
			t.Errorf("ContextTelemetry(emptyCtx, true) = %v, want default %v", got, defaultProvider)
		}
	})

	t.Run("no provider in context without default fallback returns nil", func(t *testing.T) {
		ctx := context.Background()
		got := ContextTelemetry(ctx, false)
		if got != nil {
			t.Errorf("ContextTelemetry(emptyCtx, false) = %v, want nil", got)
		}
	})

	t.Run("nil provider in context with default fallback returns default", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), providerContextKey, nil)
		got := ContextTelemetry(ctx, true)
		defaultProvider := Default()
		if got != defaultProvider {
			t.Errorf("ContextTelemetry(nilProviderCtx, true) = %v, want default %v", got, defaultProvider)
		}
	})

	t.Run("nil provider in context without default fallback returns nil", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), providerContextKey, nil)
		got := ContextTelemetry(ctx, false)
		if got != nil {
			t.Errorf("ContextTelemetry(nilProviderCtx, false) = %v, want nil", got)
		}
	})
}

func TestContextWithTracer(t *testing.T) {
	t.Run("with valid context", func(t *testing.T) {
		tracer := &noopTracer{}
		ctx := context.Background()

		newCtx := ContextWithTracer(ctx, tracer)
		if newCtx == nil {
			t.Fatal("ContextWithTracer returned nil context")
		}

		got := ContextTracer(newCtx, true)
		if got != tracer {
			t.Errorf("ContextTracer() = %v, want %v", got, tracer)
		}
	})

	t.Run("with nil context", func(t *testing.T) {
		tracer := &noopTracer{}
		newCtx := ContextWithTracer(nil, tracer) //nolint:staticcheck // intentionally testing nil context handling
		if newCtx == nil {
			t.Fatal("ContextWithTracer(nil, tracer) returned nil")
		}

		got := ContextTracer(newCtx, true)
		if got != tracer {
			t.Errorf("ContextTracer() = %v, want %v", got, tracer)
		}
	})
}

func TestContextTracer(t *testing.T) {
	t.Run("tracer in context", func(t *testing.T) {
		tracer := &noopTracer{}
		ctx := ContextWithTracer(context.Background(), tracer)

		got := ContextTracer(ctx, true)
		if got != tracer {
			t.Errorf("ContextTracer() = %v, want %v", got, tracer)
		}
	})

	t.Run("no tracer with default fallback returns noop tracer", func(t *testing.T) {
		ctx := context.Background()
		got := ContextTracer(ctx, true)
		if got == nil {
			t.Error("ContextTracer(ctx, true) should return noop tracer, got nil")
		}
	})

	t.Run("no tracer without default fallback returns nil", func(t *testing.T) {
		ctx := context.Background()
		got := ContextTracer(ctx, false)
		if got != nil {
			t.Errorf("ContextTracer(ctx, false) = %v, want nil", got)
		}
	})

	t.Run("nil context with default fallback returns noop tracer", func(t *testing.T) {
		got := ContextTracer(nil, true) //nolint:staticcheck // intentionally testing nil context handling
		if got == nil {
			t.Error("ContextTracer(nil, true) should return noop tracer, got nil")
		}
	})

	t.Run("nil context without default fallback returns nil", func(t *testing.T) {
		got := ContextTracer(nil, false) //nolint:staticcheck // intentionally testing nil context handling
		if got != nil {
			t.Errorf("ContextTracer(nil, false) = %v, want nil", got)
		}
	})
}

func TestContextTracerE(t *testing.T) {
	t.Run("tracer in context", func(t *testing.T) {
		tracer := &noopTracer{}
		ctx := ContextWithTracer(context.Background(), tracer)

		got, err := ContextTracerE(ctx, true)
		if err != nil {
			t.Fatalf("ContextTracerE() error = %v", err)
		}
		if got != tracer {
			t.Errorf("ContextTracerE() = %v, want %v", got, tracer)
		}
	})

	t.Run("no tracer with default fallback uses provider", func(t *testing.T) {
		provider := &NoopProvider{}
		ctx := ContextWithTelemetry(context.Background(), provider)

		got, err := ContextTracerE(ctx, true)
		if err != nil {
			t.Fatalf("ContextTracerE() error = %v", err)
		}
		if got == nil {
			t.Error("ContextTracerE() returned nil tracer")
		}
	})

	t.Run("no tracer without default fallback returns nil", func(t *testing.T) {
		ctx := context.Background()
		got, err := ContextTracerE(ctx, false)
		if err != nil {
			t.Fatalf("ContextTracerE() error = %v", err)
		}
		if got != nil {
			t.Errorf("ContextTracerE(ctx, false) = %v, want nil", got)
		}
	})

	t.Run("nil context with default fallback uses default provider", func(t *testing.T) {
		got, err := ContextTracerE(nil, true) //nolint:staticcheck // intentionally testing nil context handling
		if err != nil {
			t.Fatalf("ContextTracerE(nil, true) error = %v", err)
		}
		if got == nil {
			t.Error("ContextTracerE(nil, true) should return tracer from default provider")
		}
	})

	t.Run("nil context without default fallback returns nil", func(t *testing.T) {
		got, err := ContextTracerE(nil, false) //nolint:staticcheck // intentionally testing nil context handling
		if err != nil {
			t.Fatalf("ContextTracerE() error = %v", err)
		}
		if got != nil {
			t.Errorf("ContextTracerE(nil, false) = %v, want nil", got)
		}
	})
}

func TestContextWithMeter(t *testing.T) {
	t.Run("with valid context", func(t *testing.T) {
		meter := &noopMeter{}
		ctx := context.Background()

		newCtx := ContextWithMeter(ctx, meter)
		if newCtx == nil {
			t.Fatal("ContextWithMeter returned nil context")
		}

		got := ContextMeter(newCtx, true)
		if got != meter {
			t.Errorf("ContextMeter() = %v, want %v", got, meter)
		}
	})

	t.Run("with nil context", func(t *testing.T) {
		meter := &noopMeter{}
		newCtx := ContextWithMeter(nil, meter) //nolint:staticcheck // intentionally testing nil context handling
		if newCtx == nil {
			t.Fatal("ContextWithMeter(nil, meter) returned nil")
		}

		got := ContextMeter(newCtx, true)
		if got != meter {
			t.Errorf("ContextMeter() = %v, want %v", got, meter)
		}
	})
}

func TestContextMeter(t *testing.T) {
	t.Run("meter in context", func(t *testing.T) {
		meter := &noopMeter{}
		ctx := ContextWithMeter(context.Background(), meter)

		got := ContextMeter(ctx, true)
		if got != meter {
			t.Errorf("ContextMeter() = %v, want %v", got, meter)
		}
	})

	t.Run("no meter with default fallback returns noop meter", func(t *testing.T) {
		ctx := context.Background()
		got := ContextMeter(ctx, true)
		if got == nil {
			t.Error("ContextMeter(ctx, true) should return noop meter, got nil")
		}
	})

	t.Run("no meter without default fallback returns nil", func(t *testing.T) {
		ctx := context.Background()
		got := ContextMeter(ctx, false)
		if got != nil {
			t.Errorf("ContextMeter(ctx, false) = %v, want nil", got)
		}
	})

	t.Run("nil context with default fallback returns noop meter", func(t *testing.T) {
		got := ContextMeter(nil, true) //nolint:staticcheck // intentionally testing nil context handling
		if got == nil {
			t.Error("ContextMeter(nil, true) should return noop meter, got nil")
		}
	})

	t.Run("nil context without default fallback returns nil", func(t *testing.T) {
		got := ContextMeter(nil, false) //nolint:staticcheck // intentionally testing nil context handling
		if got != nil {
			t.Errorf("ContextMeter(nil, false) = %v, want nil", got)
		}
	})
}

func TestContextMeterE(t *testing.T) {
	t.Run("meter in context", func(t *testing.T) {
		meter := &noopMeter{}
		ctx := ContextWithMeter(context.Background(), meter)

		got, err := ContextMeterE(ctx, true)
		if err != nil {
			t.Fatalf("ContextMeterE() error = %v", err)
		}
		if got != meter {
			t.Errorf("ContextMeterE() = %v, want %v", got, meter)
		}
	})

	t.Run("no meter with default fallback uses provider", func(t *testing.T) {
		provider := &NoopProvider{}
		ctx := ContextWithTelemetry(context.Background(), provider)

		got, err := ContextMeterE(ctx, true)
		if err != nil {
			t.Fatalf("ContextMeterE() error = %v", err)
		}
		if got == nil {
			t.Error("ContextMeterE() returned nil meter")
		}
	})

	t.Run("no meter without default fallback returns nil", func(t *testing.T) {
		ctx := context.Background()
		got, err := ContextMeterE(ctx, false)
		if err != nil {
			t.Fatalf("ContextMeterE() error = %v", err)
		}
		if got != nil {
			t.Errorf("ContextMeterE(ctx, false) = %v, want nil", got)
		}
	})

	t.Run("nil context with default fallback uses default provider", func(t *testing.T) {
		got, err := ContextMeterE(nil, true) //nolint:staticcheck // intentionally testing nil context handling
		if err != nil {
			t.Fatalf("ContextMeterE(nil, true) error = %v", err)
		}
		if got == nil {
			t.Error("ContextMeterE(nil, true) should return meter from default provider")
		}
	})

	t.Run("nil context without default fallback returns nil", func(t *testing.T) {
		got, err := ContextMeterE(nil, false) //nolint:staticcheck // intentionally testing nil context handling
		if err != nil {
			t.Fatalf("ContextMeterE() error = %v", err)
		}
		if got != nil {
			t.Errorf("ContextMeterE(nil, false) = %v, want nil", got)
		}
	})
}

// mockNilTracerProvider is a test provider that returns nil tracer
type mockNilTracerProvider struct {
	NoopProvider
}

func (m *mockNilTracerProvider) Tracer(name string, opts ...any) (Tracer, error) {
	return nil, nil
}

// mockNilMeterProvider is a test provider that returns nil meter
type mockNilMeterProvider struct {
	NoopProvider
}

func (m *mockNilMeterProvider) Meter(name string, opts ...any) (Meter, error) {
	return nil, nil
}

// mockErrorMeterProvider is a test provider that returns an error from Meter
type mockErrorMeterProvider struct {
	NoopProvider
}

func (m *mockErrorMeterProvider) Meter(name string, opts ...any) (Meter, error) {
	return nil, ErrMeterCreation.New("test error")
}

func TestContextTracer_NilTracerFromProvider(t *testing.T) {
	// Test the fallback to noopTracer when provider returns nil tracer
	originalDefault := Default()
	defer AsDefault(originalDefault)

	AsDefault(&mockNilTracerProvider{})

	ctx := context.Background()
	tracer := ContextTracer(ctx, true)
	if tracer == nil {
		t.Error("ContextTracer should return noop tracer when provider returns nil")
	}
}

func TestContextMeter_NilMeterFromProvider(t *testing.T) {
	// Test the fallback to noopMeter when provider returns nil meter
	originalDefault := Default()
	defer AsDefault(originalDefault)

	AsDefault(&mockNilMeterProvider{})

	ctx := context.Background()
	meter := ContextMeter(ctx, true)
	if meter == nil {
		t.Error("ContextMeter should return noop meter when provider returns nil")
	}
}

func TestNewInstrument_MeterReturnsNil(t *testing.T) {
	// Test NewInstrument when meter is nil
	originalDefault := Default()
	defer AsDefault(originalDefault)

	AsDefault(&mockNilMeterProvider{})

	ctx := context.Background()
	_, err := NewInstrument[Counter[int64]](ctx, "test", InstrumentTypeCounter, CounterTypeMonotonic)
	if err == nil {
		t.Error("Expected error when meter is nil")
	}
}

func TestNewInstrument_MeterReturnsError(t *testing.T) {
	// Test NewInstrument when ContextMeterE returns an error
	originalDefault := Default()
	defer AsDefault(originalDefault)

	AsDefault(&mockErrorMeterProvider{})

	ctx := context.Background()
	_, err := NewInstrument[Counter[int64]](ctx, "test", InstrumentTypeCounter, CounterTypeMonotonic)
	if err == nil {
		t.Error("Expected error when meter returns error")
	}
}

func TestContextTracerE_ProviderNil(t *testing.T) {
	// Test case where ContextTelemetry returns nil (defaultIfNotAvailable=true but provider is nil)
	// This tests line 150 in context.go where it falls back to Default().Tracer()
	originalDefault := Default()
	defer AsDefault(originalDefault)

	// Even with no provider in context, should get tracer from default
	ctx := context.Background()
	tracer, err := ContextTracerE(ctx, true)
	if err != nil {
		t.Fatalf("ContextTracerE() error = %v", err)
	}
	if tracer == nil {
		t.Error("Expected tracer from default provider")
	}
}

func TestContextMeterE_ProviderNil(t *testing.T) {
	// Test case where ContextTelemetry returns nil (defaultIfNotAvailable=true but provider is nil)
	// This tests line 210 in context.go where it falls back to Default().Meter()
	originalDefault := Default()
	defer AsDefault(originalDefault)

	// Even with no provider in context, should get meter from default
	ctx := context.Background()
	meter, err := ContextMeterE(ctx, true)
	if err != nil {
		t.Fatalf("ContextMeterE() error = %v", err)
	}
	if meter == nil {
		t.Error("Expected meter from default provider")
	}
}

func TestContextTracer_FallbackToNoopWhenTracerEReturnsNil(t *testing.T) {
	// Test line 117-118 where ContextTracerE returns nil and we fallback to noopTracer
	ctx := context.Background()
	// ContextTracerE with defaultIfNotAvailable=false returns nil, nil
	// Then ContextTracer checks if tracer is nil and defaultIfNotAvailable is true
	tracer := ContextTracer(ctx, true)
	if tracer == nil {
		t.Error("Expected noop tracer fallback")
	}
}

func TestContextMeter_FallbackToNoopWhenMeterEReturnsNil(t *testing.T) {
	// Test line 177-179 where ContextMeterE returns nil and we fallback to noopMeter
	ctx := context.Background()
	meter := ContextMeter(ctx, true)
	if meter == nil {
		t.Error("Expected noop meter fallback")
	}
}

func TestNewInstrument(t *testing.T) {
	t.Run("creates counter successfully", func(t *testing.T) {
		ctx := context.Background()
		counter, err := NewInstrument[Counter[int64]](ctx, "test-counter", InstrumentTypeCounter, CounterTypeMonotonic)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		if counter == nil {
			t.Fatal("NewInstrument() returned nil")
		}
	})

	t.Run("creates recorder successfully", func(t *testing.T) {
		ctx := context.Background()
		recorder, err := NewInstrument[Recorder[float64]](ctx, "test-gauge", InstrumentTypeRecorder, AggregationStrategyNone)
		if err != nil {
			t.Fatalf("NewInstrument() error = %v", err)
		}
		if recorder == nil {
			t.Fatal("NewInstrument() returned nil")
		}
	})

	t.Run("returns error on type mismatch", func(t *testing.T) {
		ctx := context.Background()
		// Try to get a Recorder when creating a Counter
		_, err := NewInstrument[Recorder[float64]](ctx, "test-counter", InstrumentTypeCounter, CounterTypeMonotonic)
		if err == nil {
			t.Fatal("Expected error on type mismatch")
		}
	})

	t.Run("returns error when meter returns error", func(t *testing.T) {
		ctx := context.Background()
		// Invalid instrument type with error handling
		_, err := NewInstrument[Counter[int64]](ctx, "test", InstrumentTypeUnknown, ErrorHandlingStrategyReturn)
		if err == nil {
			t.Fatal("Expected error with invalid instrument type")
		}
	})

	t.Run("returns error when instrument is nil", func(t *testing.T) {
		ctx := context.Background()
		// No valid instrument type specified, returns nil instrument
		_, err := NewInstrument[Counter[int64]](ctx, "test")
		if err == nil {
			t.Fatal("Expected error when instrument is nil")
		}
	})
}
