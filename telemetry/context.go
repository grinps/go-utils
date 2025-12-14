package telemetry

import (
	"context"
	"reflect"
	"sync"
)

// contextKey is a type used for context keys to avoid collisions.
type contextKey string

const (
	// providerContextKey is the context key for storing the telemetry Provider.
	providerContextKey contextKey = "telemetry.provider"
	tracerContextKey   contextKey = "telemetry.tracer"
	meterContextKey    contextKey = "telemetry.meter"
	DefaultTracerName  string     = "default-tracer"
	DefaultMeterName   string     = "default-meter"
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
// If no provider is found in the context, it returns the Default provider if defaultIfNotAvailable is true.
// If no provider is found in the context and defaultIfNotAvailable is false, it returns nil.
//
// Example:
//
//	provider := telemetry.ContextTelemetry(ctx, true)
//	tracer := provider.Tracer("my-service")
func ContextTelemetry(ctx context.Context, defaultIfNotAvailable bool) Provider {
	if ctx == nil {
		if defaultIfNotAvailable {
			return Default()
		}
		return nil
	}
	if provider, ok := ctx.Value(providerContextKey).(Provider); ok && provider != nil {
		return provider
	}
	if defaultIfNotAvailable {
		return Default()
	}
	return nil
}

// ContextWithTracer returns a new context with the given tracer stored in it.
// This allows passing a specific tracer through the call chain.
//
// Example:
//
//	ctx := telemetry.ContextWithTracer(ctx, tracer)
func ContextWithTracer(ctx context.Context, tracer Tracer) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, tracerContextKey, tracer)
}

// ContextTracer retrieves the tracer from the context.
// If no tracer is found in the context, it returns the noop tracer if defaultIfNotAvailable is true.
// If no tracer is found in the context and defaultIfNotAvailable is false, it returns nil.
// See `ContextTracerE` for more details.
//
// Example:
//
//	tracer := telemetry.ContextTracer(ctx, true)
//	span := tracer.Start("my-service")
func ContextTracer(ctx context.Context, defaultIfNotAvailable bool) Tracer {
	tracer, _ := ContextTracerE(ctx, defaultIfNotAvailable)
	if tracer == nil {
		if defaultIfNotAvailable {
			tracer = &noopTracer{}
		}
	}
	return tracer
}

// ContextTracerE returns the tracer from the context.
// If no tracer is found in the context and defaultIfNotAvailable is true, it uses available or default provider to get default tracer.
// If no tracer is found in the context and defaultIfNotAvailable is false, it returns nil.
// If an error occurs, it returns an error.
//
// Example:
//
// tracer, err := telemetry.ContextTracerE(ctx, true)
//
//	if err != nil {
//		return nil, err
//	}
//
// span := tracer.Start("my-service")
func ContextTracerE(ctx context.Context, defaultIfNotAvailable bool) (Tracer, error) {
	if ctx != nil {
		if availableTracer, ok := ctx.Value(tracerContextKey).(Tracer); ok && availableTracer != nil {
			return availableTracer, nil
		}
	}
	if !defaultIfNotAvailable {
		return nil, nil
	}
	// ContextTelemetry with defaultIfNotAvailable=true always returns non-nil provider
	return ContextTelemetry(ctx, true).Tracer(DefaultTracerName)
}

// ContextWithMeter returns a new context with the given meter stored in it.
// This allows passing a specific meter through the call chain.
//
// Example:
//
//	ctx := telemetry.ContextWithMeter(ctx, meter)
func ContextWithMeter(ctx context.Context, meter Meter) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, meterContextKey, meter)
}

// ContextMeter retrieves the meter from the context.
// If no meter is found in the context and defaultIfNotAvailable is true, it returns the noop meter.
// If no meter is found in the context and defaultIfNotAvailable is false, it returns nil.
// Incase of error or no meter being available, it returns a Noop meter.
// See `ContextMeterE` for more details.
//
// Example:
//
//	meter := telemetry.ContextMeter(ctx, true)
//	counter, _ := meter.NewInstrument[Counter[int64]]("requests_total", InstrumentTypeCounter, CounterTypeMonotonic)
func ContextMeter(ctx context.Context, defaultIfNotAvailable bool) Meter {
	meter, _ := ContextMeterE(ctx, defaultIfNotAvailable)
	if meter == nil && defaultIfNotAvailable {
		meter = &noopMeter{}
	}
	return meter
}

// ContextMeterE retrieves the meter from the context.
// If no meter is found in the context and defaultIfNotAvailable is true, it uses available or default provider to get default meter.
// If no meter is found in the context and defaultIfNotAvailable is false, it returns nil.
// If an error occurs, it returns an error.
//
// Example:
//
// meter, err := telemetry.ContextMeterE(ctx, true)
//
//	if err != nil {
//	 return nil, err
//	}
//
// counter, _ := meter.NewInstrument[Counter[int64]]("requests_total", InstrumentTypeCounter, CounterTypeMonotonic)
func ContextMeterE(ctx context.Context, defaultIfNotAvailable bool) (Meter, error) {
	if ctx != nil {
		if availableMeter, ok := ctx.Value(meterContextKey).(Meter); ok && availableMeter != nil {
			return availableMeter, nil
		}
	}
	if !defaultIfNotAvailable {
		return nil, nil
	}
	// ContextTelemetry with defaultIfNotAvailable=true always returns non-nil provider
	return ContextTelemetry(ctx, true).Meter(DefaultMeterName)
}

// NewInstrument returns a new instrument from the available or default Meter.
// The name parameter identifies the instrument.
//
// Example:
//
//	counter, _ := meter.NewInstrument[Counter[int64]]("requests_total", InstrumentTypeCounter, CounterTypeMonotonic)
func NewInstrument[T Instrument](ctx context.Context, name string, opts ...any) (T, error) {
	var zero T
	meter, err := ContextMeterE(ctx, true)
	if err != nil {
		return zero, err
	}
	if meter == nil {
		return zero, ErrInstrumentCreation.New(ErrReasonNilMeter, ErrParamName, name)
	}
	newInstrument, err := meter.NewInstrument(name, opts...)
	if err != nil {
		return zero, ErrInstrumentCreation.NewWithError(ErrReasonNewInstrumentFailed, err, ErrParamName, name, ErrParamOption, opts)
	}
	if newInstrument == nil {
		return zero, ErrInstrumentCreation.New(ErrReasonNewInstrumentFailed, ErrParamName, name, ErrParamOption, opts)
	}
	if newInstrumentAsT, ok := newInstrument.(T); !ok {
		var expectedType string
		if t := reflect.TypeOf((*T)(nil)); t != nil {
			expectedType = t.Elem().String()
		}
		return zero, ErrInstrumentCreation.New(ErrReasonInvalidInstrumentType, ErrParamName, name, ErrParamExpectedType, expectedType, ErrParamActualType, reflect.TypeOf(newInstrument).String())
	} else {
		return newInstrumentAsT, nil
	}
}
