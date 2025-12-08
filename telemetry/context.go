package telemetry

import (
	"context"
	"sync"
)

// contextKey is a type used for context keys to avoid collisions.
type contextKey string

const (
	// providerContextKey is the context key for storing the telemetry Provider.
	providerContextKey contextKey = "telemetry.provider"
)

var (
	// defaultProvider holds the global default telemetry provider.
	defaultProvider Provider = &NoopProvider{}
	// defaultProviderMu protects access to defaultProvider.
	defaultProviderMu sync.RWMutex
)

// Default returns the default telemetry Provider.
// If a default provider has been set via AsDefault, it returns that provider.
// If no provider is available, it returns nil.
func Default() Provider {
	defaultProviderMu.RLock()
	defer defaultProviderMu.RUnlock()
	return defaultProvider
}

// AsDefault sets the given Provider as the default Provider.
// The provider must not be nil.
// This function is safe for concurrent use.
//
// Example:
//
//	provider := memory.NewProvider()
//	telemetry.AsDefault(provider)
func AsDefault(provider Provider) {
	if provider == nil {
		return
	}
	defaultProviderMu.Lock()
	defer defaultProviderMu.Unlock()
	defaultProvider = provider
}

// ContextWithTelemetry returns a new context with the given telemetry Provider stored in it.
// This allows passing a specific provider through the call chain.
//
// Example:
//
//	ctx := telemetry.ContextWithTelemetry(ctx, provider)
func ContextWithTelemetry(ctx context.Context, provider Provider) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, providerContextKey, provider)
}

// ContextTelemetry retrieves the telemetry Provider from the context.
// If no provider is found in the context, it returns the Default provider.
//
// Example:
//
//	provider := telemetry.ContextTelemetry(ctx)
//	tracer := provider.Tracer("my-service")
func ContextTelemetry(ctx context.Context) Provider {
	if ctx == nil {
		return Default()
	}
	if provider, ok := ctx.Value(providerContextKey).(Provider); ok && provider != nil {
		return provider
	}
	return Default()
}

// NewTracer returns a Tracer from the Provider in the context.
// If no provider is found in the context, it uses the Default provider.
// The name parameter identifies the instrumentation library or application component.
//
// Example:
//
//	tracer := telemetry.NewTracer(ctx, "my-service")
//	ctx, span := tracer.Start(ctx, "operation")
//	defer span.End()
func NewTracer(ctx context.Context, name string, opts ...any) (Tracer, error) {
	return ContextTelemetry(ctx).Tracer(name, opts...)
}

// NewMeter returns a Meter from the Provider in the context.
// If no provider is found in the context, it uses the Default provider.
// The name parameter identifies the instrumentation library or application component.
//
// Example:
//
//	meter := telemetry.NewMeter(ctx, "my-service")
//	counter, _ := meter.NewInstrument("requests_total")
func NewMeter(ctx context.Context, name string, opts ...any) (Meter, error) {
	return ContextTelemetry(ctx).Meter(name, opts...)
}
