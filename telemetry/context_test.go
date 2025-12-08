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
		got := ContextTelemetry(newCtx)
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
		got := ContextTelemetry(newCtx)
		if got != provider {
			t.Errorf("ContextTelemetry() = %v, want %v", got, provider)
		}
	})
}

func TestContextTelemetry(t *testing.T) {
	t.Run("provider in context", func(t *testing.T) {
		provider := &NoopProvider{}
		ctx := ContextWithTelemetry(context.Background(), provider)

		got := ContextTelemetry(ctx)
		if got != provider {
			t.Errorf("ContextTelemetry() = %v, want %v", got, provider)
		}
	})

	t.Run("nil context returns default", func(t *testing.T) {
		got := ContextTelemetry(nil) //nolint:staticcheck // intentionally testing nil context handling
		defaultProvider := Default()
		if got != defaultProvider {
			t.Errorf("ContextTelemetry(nil) = %v, want default %v", got, defaultProvider)
		}
	})

	t.Run("no provider in context returns default", func(t *testing.T) {
		ctx := context.Background()
		got := ContextTelemetry(ctx)
		defaultProvider := Default()
		if got != defaultProvider {
			t.Errorf("ContextTelemetry(emptyCtx) = %v, want default %v", got, defaultProvider)
		}
	})

	t.Run("nil provider in context returns default", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), providerContextKey, nil)
		got := ContextTelemetry(ctx)
		defaultProvider := Default()
		if got != defaultProvider {
			t.Errorf("ContextTelemetry(nilProviderCtx) = %v, want default %v", got, defaultProvider)
		}
	})
}

func TestNewTracer(t *testing.T) {
	t.Run("with provider in context", func(t *testing.T) {
		provider := &NoopProvider{}
		ctx := ContextWithTelemetry(context.Background(), provider)

		tracer, err := NewTracer(ctx, "test-tracer")
		if err != nil {
			t.Fatalf("NewTracer() error = %v", err)
		}
		if tracer == nil {
			t.Fatal("NewTracer() returned nil tracer")
		}
	})

	t.Run("with default provider", func(t *testing.T) {
		ctx := context.Background()
		tracer, err := NewTracer(ctx, "test-tracer")
		if err != nil {
			t.Fatalf("NewTracer() error = %v", err)
		}
		if tracer == nil {
			t.Fatal("NewTracer() returned nil tracer")
		}
	})

	t.Run("with nil context uses default", func(t *testing.T) {
		tracer, err := NewTracer(nil, "test-tracer") //nolint:staticcheck // intentionally testing nil context handling
		if err != nil {
			t.Fatalf("NewTracer(nil, ...) error = %v", err)
		}
		if tracer == nil {
			t.Fatal("NewTracer(nil, ...) returned nil tracer")
		}
	})

	t.Run("passes options to provider", func(t *testing.T) {
		ctx := context.Background()
		_, err := NewTracer(ctx, "test", ErrorHandlingStrategyGenerateError)
		if err == nil {
			t.Error("Expected error with ErrorHandlingStrategyGenerateError")
		}
	})
}

func TestNewMeter(t *testing.T) {
	t.Run("with provider in context", func(t *testing.T) {
		provider := &NoopProvider{}
		ctx := ContextWithTelemetry(context.Background(), provider)

		meter, err := NewMeter(ctx, "test-meter")
		if err != nil {
			t.Fatalf("NewMeter() error = %v", err)
		}
		if meter == nil {
			t.Fatal("NewMeter() returned nil meter")
		}
	})

	t.Run("with default provider", func(t *testing.T) {
		ctx := context.Background()
		meter, err := NewMeter(ctx, "test-meter")
		if err != nil {
			t.Fatalf("NewMeter() error = %v", err)
		}
		if meter == nil {
			t.Fatal("NewMeter() returned nil meter")
		}
	})

	t.Run("with nil context uses default", func(t *testing.T) {
		meter, err := NewMeter(nil, "test-meter") //nolint:staticcheck // intentionally testing nil context handling
		if err != nil {
			t.Fatalf("NewMeter(nil, ...) error = %v", err)
		}
		if meter == nil {
			t.Fatal("NewMeter(nil, ...) returned nil tracer")
		}
	})

	t.Run("passes options to provider", func(t *testing.T) {
		ctx := context.Background()
		_, err := NewMeter(ctx, "test", ErrorHandlingStrategyGenerateError)
		if err == nil {
			t.Error("Expected error with ErrorHandlingStrategyGenerateError")
		}
	})
}
