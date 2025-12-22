package otel

import (
	"context"
	"testing"

	"github.com/grinps/go-utils/config"
	"github.com/grinps/go-utils/config/ext"
	"github.com/grinps/go-utils/errext"
	"github.com/grinps/go-utils/telemetry"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

func TestNewProvider(t *testing.T) {
	ctx := context.Background()

	t.Run("with default config", func(t *testing.T) {
		provider, err := NewProvider(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
		defer provider.Shutdown(ctx)
	})

	t.Run("with disabled option", func(t *testing.T) {
		disabled := true
		provider, err := NewProvider(ctx, WithDisabled(disabled))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
		defer provider.Shutdown(ctx)
	})
}

func TestNewProviderFromConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("with valid config", func(t *testing.T) {
		simpleCfg := ext.NewConfigWrapper(config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
			"opentelemetry": map[string]any{
				"file_format": "0.3",
			},
		})))

		provider, err := NewProviderFromConfig(ctx, simpleCfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if provider == nil {
			t.Fatal("expected provider to be non-nil")
		}
		defer provider.Shutdown(ctx)
	})

	t.Run("with empty config uses defaults", func(t *testing.T) {
		simpleCfg := ext.NewConfigWrapper(config.NewSimpleConfig(ctx))

		provider, err := NewProviderFromConfig(ctx, simpleCfg)
		if !errext.Is(err, ErrCodeConfigLoadFailed) {
			t.Fatalf("expected ErrCodeConfigLoadFailed error, got %v", err)
		}
		if provider != nil {
			defer provider.Shutdown(ctx)
			t.Fatal("expected provider to be nil")
		}
	})
}

func TestProviderTracer(t *testing.T) {
	ctx := context.Background()
	provider, _ := NewProvider(ctx)
	defer provider.Shutdown(ctx)

	t.Run("basic tracer", func(t *testing.T) {
		tracer, err := provider.Tracer("test-tracer")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tracer == nil {
			t.Fatal("expected tracer to be non-nil")
		}
	})

	t.Run("tracer with version", func(t *testing.T) {
		tracer, err := provider.Tracer("test-tracer", trace.WithInstrumentationVersion("1.0.0"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tracer == nil {
			t.Fatal("expected tracer to be non-nil")
		}
	})

	t.Run("tracer with schemaURL", func(t *testing.T) {
		tracer, err := provider.Tracer("test-tracer-schema", trace.WithSchemaURL("http://example.com"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if tracer == nil {
			t.Fatal("expected tracer to be non-nil")
		}
	})
}

func TestProviderMeter(t *testing.T) {
	ctx := context.Background()
	provider, err := NewProvider(ctx)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Shutdown(ctx)

	t.Run("basic meter", func(t *testing.T) {
		meter, err := provider.Meter("test-meter")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if meter == nil {
			t.Fatal("expected meter to be non-nil")
		}
	})

	t.Run("meter with version", func(t *testing.T) {
		meter, err := provider.Meter("test-meter", metric.WithInstrumentationVersion("1.0.0"))
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if meter == nil {
			t.Fatal("expected meter to be non-nil")
		}
	})
}

func TestProviderShutdown(t *testing.T) {
	ctx := context.Background()

	t.Run("normal shutdown", func(t *testing.T) {
		provider, _ := NewProvider(ctx)
		err := provider.Shutdown(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("double shutdown", func(t *testing.T) {
		provider, _ := NewProvider(ctx)
		_ = provider.Shutdown(ctx)
		err := provider.Shutdown(ctx)
		if !errext.Is(err, ErrCodeShutdown) {
			t.Fatalf("expected ErrCodeShutdown, got %v", err)
		}
	})

	t.Run("tracer after shutdown returns noop", func(t *testing.T) {
		provider, _ := NewProvider(ctx)
		_ = provider.Shutdown(ctx)

		tracer, err := provider.Tracer("test")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// Should return noop tracer
		ctx2, span := tracer.Start(ctx, "test-span")
		if ctx2 == nil {
			t.Error("expected context to be non-nil")
		}
		if span == nil {
			t.Error("expected span to be non-nil")
		}
		if span.IsRecording() {
			t.Error("expected noop span to not be recording")
		}
		span.End()
	})

	t.Run("meter after shutdown returns noop", func(t *testing.T) {
		provider, _ := NewProvider(ctx)
		_ = provider.Shutdown(ctx)

		meter, err := provider.Meter("test")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		// Should return noop meter
		inst, err := meter.NewInstrument("test-counter",
			telemetry.InstrumentTypeCounter,
			telemetry.CounterTypeMonotonic,
		)
		if err != nil {
			t.Fatalf("expected no error from noop meter, got %v", err)
		}
		if inst == nil {
			t.Error("expected instrument to be non-nil")
		}
	})
}

func TestProviderImplementsInterface(t *testing.T) {
	var _ telemetry.Provider = (*Provider)(nil)
}

func TestWithDisabledOption(t *testing.T) {
	cfg := DefaultConfiguration()
	disabled := true
	WithDisabled(disabled)(cfg)

	if cfg.Disabled == nil || *cfg.Disabled != true {
		t.Errorf("expected Disabled to be true, got %v", cfg.Disabled)
	}
}

func TestProviderTracerWithNilSDK(t *testing.T) {
	// Create a provider with nil sdk to test fallback to global tracer
	p := &Provider{sdk: nil, shutdown: false}

	tracer, err := p.Tracer("test-tracer")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tracer == nil {
		t.Fatal("expected tracer to be non-nil")
	}
}

func TestProviderMeterWithNilSDK(t *testing.T) {
	// Create a provider with nil sdk to test fallback to global meter
	p := &Provider{sdk: nil, shutdown: false}

	meter, err := p.Meter("test-meter")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if meter == nil {
		t.Fatal("expected meter to be non-nil")
	}
}

func TestProviderShutdownWithNilSDK(t *testing.T) {
	// Create a provider with nil sdk
	p := &Provider{sdk: nil, shutdown: false}

	err := p.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("expected no error for nil sdk shutdown, got %v", err)
	}
}
