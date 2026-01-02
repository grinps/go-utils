package config

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"github.com/grinps/go-utils/telemetry"
)

// Global feature flag for telemetry
var telemetryEnabled atomic.Bool

func init() {
	telemetryEnabled.Store(true)
}

// SetTelemetryEnabled globally toggles telemetry for the config package.
// When disabled, no spans or metrics are recorded.
//
// Example:
//
//	config.SetTelemetryEnabled(false) // Disable all config telemetry
func SetTelemetryEnabled(enabled bool) {
	telemetryEnabled.Store(enabled)
}

// IsTelemetryEnabled returns the current state of the global telemetry flag.
func IsTelemetryEnabled() bool {
	return telemetryEnabled.Load()
}

// TelemetryAware is an optional interface for Config implementations.
// It allows providers to control instrumentation and provide context-specific attributes.
type TelemetryAware interface {
	// ShouldInstrument allows opting out of telemetry based on context, key, and operation.
	// Return false to skip telemetry for sensitive keys (e.g., "database.password").
	//
	// Example:
	//   func (cfg *MyConfig) ShouldInstrument(ctx context.Context, key, op string) bool {
	//       return !strings.HasPrefix(key, "secrets.")
	//   }
	ShouldInstrument(ctx context.Context, key string, op string) bool

	// GenerateTelemetryAttributes allows the provider to generate/augment attributes.
	// It receives the base attributes (pre-populated by package) and returns the final slice.
	// The provider can append its own attributes (e.g., "config.source", "config.version").
	//
	// Example:
	//   func (cfg *MyConfig) GenerateTelemetryAttributes(ctx context.Context, op string, attrs []any) []any {
	//       return append(attrs, "config.source", "etcd", "config.version", "v2")
	//   }
	GenerateTelemetryAttributes(ctx context.Context, op string, attrs []any) []any
}

// telemetryState holds state between startTelemetry and finishTelemetry.
type telemetryState struct {
	span     telemetry.Span
	meter    telemetry.Meter
	start    time.Time
	attrs    []any
	active   bool
	op       string
	implType string
}

// startTelemetry initiates the telemetry lifecycle for an operation.
// It returns a context with the active span and the state for finishing.
// The extraAttrs parameter allows operation-specific attributes (e.g., target_type for Unmarshal).
func startTelemetry(ctx context.Context, cfg Config, op, key string, extraAttrs ...any) (context.Context, telemetryState) {
	// 1. Global Kill Switch
	if !telemetryEnabled.Load() {
		return ctx, telemetryState{active: false}
	}

	// 2. Check TelemetryAware (Single Cast)
	var aware TelemetryAware
	isAware := false
	if cfg != nil {
		aware, isAware = cfg.(TelemetryAware)
	}

	// 3. Provider Opt-Out
	if isAware && !aware.ShouldInstrument(ctx, key, op) {
		return ctx, telemetryState{active: false}
	}

	// 4. Determine implementation name
	implType := "nil"
	if cfg != nil {
		implType = string(cfg.Name())
	}

	// 5. Base Attributes
	// Pre-allocate: 6 base + 4 result + 10 provider + extraAttrs
	attrs := make([]any, 0, 20+len(extraAttrs))
	attrs = append(attrs,
		"config.key_prefix", extractKeyPrefix(key),
		"config.impl_type", implType,
	)

	// 6. Operation-specific extra attributes (e.g., target_type for Unmarshal)
	attrs = append(attrs, extraAttrs...)

	// 7. Generate/Augment Attributes via provider
	if isAware {
		attrs = aware.GenerateTelemetryAttributes(ctx, op, attrs)
	}

	// 8. Get Telemetry Components from context
	tracer := telemetry.ContextTracer(ctx, true)
	meter := telemetry.ContextMeter(ctx, true)

	// 9. Start Span (Crucial for parenting - returns new ctx)
	ctx, span := tracer.Start(ctx, "config."+op, attrs...)

	return ctx, telemetryState{
		span:     span,
		meter:    meter,
		start:    time.Now(),
		attrs:    attrs,
		active:   true,
		op:       op,
		implType: implType,
	}
}

// finishTelemetry completes the telemetry lifecycle.
// It records duration, success status, ends the span, and records metrics.
func finishTelemetry(ctx context.Context, state telemetryState, err error) {
	if !state.active {
		return
	}

	duration := time.Since(state.start).Milliseconds()
	success := err == nil

	// Update span with result
	state.span.SetAttributes("config.success", success)
	if err != nil {
		state.span.RecordError(err)
	}
	state.span.End()

	// Build final attributes for metrics
	metricAttrs := append(state.attrs, "config.success", success)

	// Operation-specific count metric (e.g., config.get_value.count)
	countMetricName := "config." + toSnakeCase(state.op) + ".count"
	recordCounter(ctx, state.meter, countMetricName, 1, metricAttrs...)

	// Operation-specific duration metric (e.g., config.get_value.duration_ms)
	durationMetricName := "config." + toSnakeCase(state.op) + ".duration_ms"
	recordHistogram(ctx, state.meter, durationMetricName, duration, metricAttrs...)

	// Dedicated error counter
	if err != nil {
		errorAttrs := []any{
			"config.operation", state.op,
			"config.impl_type", state.implType,
		}
		if codeErr, ok := err.(interface{ Code() int }); ok {
			errorAttrs = append(errorAttrs, "config.error_code", codeErr.Code())
		}
		recordCounter(ctx, state.meter, "config.errors.count", 1, errorAttrs...)
	}
}

// extractKeyPrefix returns the first segment of a dotted key for cardinality control.
func extractKeyPrefix(key string) string {
	if idx := strings.Index(key, "."); idx > 0 {
		return key[:idx]
	}
	return key
}

// toSnakeCase converts operation names to snake_case for metric names.
func toSnakeCase(s string) string {
	return strings.ToLower(strings.ReplaceAll(s, " ", "_"))
}

// recordCounter records a counter metric using the provided meter.
func recordCounter(ctx context.Context, meter telemetry.Meter, name string, value int64, attrs ...any) {
	if meter == nil {
		return
	}
	counter, err := meter.NewInstrument(name, telemetry.InstrumentTypeCounter, telemetry.CounterTypeMonotonic)
	if err != nil {
		return
	}
	if c, ok := counter.(telemetry.Counter[int64]); ok {
		c.Add(ctx, value, attrs...)
	}
}

// recordHistogram records a histogram metric using the provided meter.
func recordHistogram(ctx context.Context, meter telemetry.Meter, name string, value int64, attrs ...any) {
	if meter == nil {
		return
	}
	recorder, err := meter.NewInstrument(name, telemetry.InstrumentTypeRecorder, telemetry.AggregationStrategyHistogram)
	if err != nil {
		return
	}
	if r, ok := recorder.(telemetry.Recorder[int64]); ok {
		r.Record(ctx, value, attrs...)
	}
}
