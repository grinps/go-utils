package otel

import (
	"context"

	"github.com/grinps/go-utils/config"
	"github.com/grinps/go-utils/telemetry"
	"go.opentelemetry.io/contrib/otelconf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Provider implements telemetry.Provider using OpenTelemetry.
type Provider struct {
	sdk      *otelconf.SDK
	shutdown bool
}

// NewProviderFromConfig creates a provider from a config.Config.
// It uses LoadConfiguration to unmarshal the OpenTelemetry configuration
// from the config at the ConfigKey ("opentelemetry").
func NewProviderFromConfig(ctx context.Context, cfg config.Config) (*Provider, error) {
	otelCfg, err := LoadConfiguration(ctx, cfg)
	if err != nil {
		return nil, ErrCodeConfigInvalid.NewWithError("failed to load configuration", err)
	}
	return newProviderFromOtelConfig(ctx, otelCfg)
}

// NewProvider creates a provider with options applied to a default configuration.
// This is a convenience function for simple use cases.
func NewProvider(ctx context.Context, opts ...ProviderOption) (*Provider, error) {
	cfg := DefaultConfiguration()
	for _, opt := range opts {
		opt(cfg)
	}
	return newProviderFromOtelConfig(ctx, cfg)
}

func newProviderFromOtelConfig(ctx context.Context, cfg *otelconf.OpenTelemetryConfiguration) (*Provider, error) {
	// Create SDK from configuration
	sdk, err := otelconf.NewSDK(
		otelconf.WithContext(ctx),
		otelconf.WithOpenTelemetryConfiguration(*cfg),
	)
	if err != nil {
		return nil, ErrCodeProviderCreation.NewWithError("failed to create SDK", err)
	}

	// Set global providers
	if tp := sdk.TracerProvider(); tp != nil {
		otel.SetTracerProvider(tp)
	}
	if mp := sdk.MeterProvider(); mp != nil {
		otel.SetMeterProvider(mp)
	}

	return &Provider{
		sdk: &sdk,
	}, nil
}

// Tracer returns a Tracer for creating spans.
// Returns the telemetry.NoopProvider's tracer after shutdown.
func (p *Provider) Tracer(name string, opts ...any) (telemetry.Tracer, error) {
	if p.shutdown {
		return telemetry.Default().Tracer(name, opts...)
	}

	var otelOpts []trace.TracerOption
	for _, opt := range opts {
		if o, ok := opt.(trace.TracerOption); ok {
			otelOpts = append(otelOpts, o)
		}
	}

	var otelTracer trace.Tracer
	if p.sdk != nil {
		if tp := p.sdk.TracerProvider(); tp != nil {
			otelTracer = tp.Tracer(name, otelOpts...)
		} else {
			otelTracer = otel.Tracer(name, otelOpts...)
		}
	} else {
		otelTracer = otel.Tracer(name, otelOpts...)
	}

	return &Tracer{
		provider: p,
		Tracer:   otelTracer,
	}, nil
}

// Meter returns a Meter for creating metric instruments.
// Returns the telemetry.NoopProvider's meter after shutdown.
func (p *Provider) Meter(name string, opts ...any) (telemetry.Meter, error) {
	if p.shutdown {
		return telemetry.Default().Meter(name, opts...)
	}

	var otelOpts []metric.MeterOption
	for _, opt := range opts {
		if o, ok := opt.(metric.MeterOption); ok {
			otelOpts = append(otelOpts, o)
		}
	}

	var otelMeter metric.Meter
	if p.sdk != nil {
		if mp := p.sdk.MeterProvider(); mp != nil {
			otelMeter = mp.Meter(name, otelOpts...)
		} else {
			otelMeter = otel.Meter(name, otelOpts...)
		}
	} else {
		otelMeter = otel.Meter(name, otelOpts...)
	}

	return &Meter{
		Meter: otelMeter,
	}, nil
}

// Shutdown shuts down the provider.
func (p *Provider) Shutdown(ctx context.Context) error {
	if p.shutdown {
		return ErrCodeShutdown.New("provider already shutdown")
	}

	p.shutdown = true

	if p.sdk != nil {
		if err := p.sdk.Shutdown(ctx); err != nil {
			return ErrCodeShutdown.NewWithError("failed to shutdown SDK", err)
		}
	}

	return nil
}

// ProviderOption is a functional option for configuring the Provider via NewProvider.
type ProviderOption func(*otelconf.OpenTelemetryConfiguration)

// WithDisabled disables the SDK.
func WithDisabled(disabled bool) ProviderOption {
	return func(cfg *otelconf.OpenTelemetryConfiguration) {
		cfg.Disabled = &disabled
	}
}

// Ensure Provider implements telemetry.Provider interface.
var _ telemetry.Provider = (*Provider)(nil)
