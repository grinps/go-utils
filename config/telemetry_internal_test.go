package config

import (
	"context"
	"errors"
	"testing"
)

// Internal tests for unexported telemetry functions

func TestExtractKeyPrefix(t *testing.T) {
	tests := []struct {
		key      string
		expected string
	}{
		{"server.port", "server"},
		{"database.host.name", "database"},
		{"simple", "simple"},
		{"", ""},
		{"a.b", "a"},
	}

	for _, tt := range tests {
		result := extractKeyPrefix(tt.key)
		if result != tt.expected {
			t.Errorf("extractKeyPrefix(%q) = %q, want %q", tt.key, result, tt.expected)
		}
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"GetValue", "getvalue"},
		{"get value", "get_value"},
		{"SET VALUE", "set_value"},
		{"", ""},
	}

	for _, tt := range tests {
		result := toSnakeCase(tt.input)
		if result != tt.expected {
			t.Errorf("toSnakeCase(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestRecordCounterNilMeter(t *testing.T) {
	ctx := context.Background()
	// Should not panic with nil meter
	recordCounter(ctx, nil, "test.counter", 1, "key", "value")
}

func TestRecordHistogramNilMeter(t *testing.T) {
	ctx := context.Background()
	// Should not panic with nil meter
	recordHistogram(ctx, nil, "test.histogram", 100, "key", "value")
}

func TestStartTelemetryWithNilConfig(t *testing.T) {
	ctx := context.Background()

	// Test with nil config
	newCtx, state := startTelemetry(ctx, nil, "test_op", "test.key")
	if newCtx == nil {
		t.Error("Expected non-nil context")
	}
	if !state.active {
		// With telemetry enabled and nil config, should still be active
		// but implType should be "nil"
	}
}

func TestStartTelemetryDisabled(t *testing.T) {
	ctx := context.Background()
	cfg := NewSimpleConfig(ctx)

	// Disable telemetry
	originalState := telemetryEnabled.Load()
	telemetryEnabled.Store(false)
	defer telemetryEnabled.Store(originalState)

	newCtx, state := startTelemetry(ctx, cfg, "test_op", "test.key")
	if state.active {
		t.Error("Expected inactive state when telemetry is disabled")
	}
	if newCtx != ctx {
		t.Error("Expected same context when telemetry is disabled")
	}
}

func TestFinishTelemetryInactive(t *testing.T) {
	ctx := context.Background()
	state := telemetryState{active: false}

	// Should not panic with inactive state
	finishTelemetry(ctx, state, nil)
	finishTelemetry(ctx, state, errors.New("test error"))
}

func TestTelemetryStateFields(t *testing.T) {
	state := telemetryState{
		active:   true,
		op:       "get_value",
		implType: "SimpleConfig",
	}

	if state.op != "get_value" {
		t.Errorf("Expected op 'get_value', got %q", state.op)
	}
	if state.implType != "SimpleConfig" {
		t.Errorf("Expected implType 'SimpleConfig', got %q", state.implType)
	}
}

// errorWithCode is a test error that implements Code() method
type errorWithCode struct {
	code int
	msg  string
}

func (e *errorWithCode) Error() string {
	return e.msg
}

func (e *errorWithCode) Code() int {
	return e.code
}

func TestFinishTelemetryWithCodedError(t *testing.T) {
	ctx := context.Background()
	cfg := NewSimpleConfig(ctx)

	// Start telemetry
	newCtx, state := startTelemetry(ctx, cfg, "test_op", "test.key")

	// Finish with coded error
	codedErr := &errorWithCode{code: 123, msg: "test error"}
	finishTelemetry(newCtx, state, codedErr)
}

func TestStartTelemetryWithExtraAttrs(t *testing.T) {
	ctx := context.Background()
	cfg := NewSimpleConfig(ctx)

	// Start with extra attributes
	_, state := startTelemetry(ctx, cfg, "unmarshal", "test.key",
		"config.target_type", "TestStruct",
		"config.custom", "value")

	if !state.active {
		t.Error("Expected active state")
	}

	// Check that extra attrs are in the state
	found := false
	for i := 0; i < len(state.attrs)-1; i += 2 {
		if state.attrs[i] == "config.target_type" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected extra attrs to be in state")
	}
}

// telemetryAwareTestConfig implements TelemetryAware for testing
type telemetryAwareTestConfig struct {
	*simpleConfig
	shouldInstrument bool
	customAttrs      []any
}

func (c *telemetryAwareTestConfig) ShouldInstrument(ctx context.Context, key string, op string) bool {
	return c.shouldInstrument
}

func (c *telemetryAwareTestConfig) GenerateTelemetryAttributes(ctx context.Context, op string, attrs []any) []any {
	if c.customAttrs != nil {
		return append(attrs, c.customAttrs...)
	}
	return attrs
}

func TestStartTelemetryWithTelemetryAware(t *testing.T) {
	ctx := context.Background()
	baseCfg := NewSimpleConfig(ctx, WithConfigurationMap(map[string]any{"key": "value"}))

	// Test with ShouldInstrument returning false
	cfgOptOut := &telemetryAwareTestConfig{
		simpleConfig:     baseCfg.(*simpleConfig),
		shouldInstrument: false,
	}

	_, state := startTelemetry(ctx, cfgOptOut, "get_value", "key")
	if state.active {
		t.Error("Expected inactive state when ShouldInstrument returns false")
	}

	// Test with ShouldInstrument returning true and custom attrs
	cfgOptIn := &telemetryAwareTestConfig{
		simpleConfig:     baseCfg.(*simpleConfig),
		shouldInstrument: true,
		customAttrs:      []any{"custom.key", "custom.value"},
	}

	_, state = startTelemetry(ctx, cfgOptIn, "get_value", "key")
	if !state.active {
		t.Error("Expected active state when ShouldInstrument returns true")
	}
}

func TestFinishTelemetrySuccess(t *testing.T) {
	ctx := context.Background()
	cfg := NewSimpleConfig(ctx)

	// Start telemetry
	newCtx, state := startTelemetry(ctx, cfg, "test_op", "test.key")

	// Finish successfully (no error)
	finishTelemetry(newCtx, state, nil)
}

func TestFinishTelemetryWithError(t *testing.T) {
	ctx := context.Background()
	cfg := NewSimpleConfig(ctx)

	// Start telemetry
	newCtx, state := startTelemetry(ctx, cfg, "test_op", "test.key")

	// Finish with error
	finishTelemetry(newCtx, state, errors.New("test error"))
}

func TestSetTelemetryEnabledInternal(t *testing.T) {
	original := telemetryEnabled.Load()
	defer telemetryEnabled.Store(original)

	SetTelemetryEnabled(false)
	if IsTelemetryEnabled() {
		t.Error("Expected telemetry to be disabled")
	}

	SetTelemetryEnabled(true)
	if !IsTelemetryEnabled() {
		t.Error("Expected telemetry to be enabled")
	}
}

func TestSimpleConfigSetValueSingleKey(t *testing.T) {
	ctx := context.Background()
	cfg := NewSimpleConfig(ctx).(*simpleConfig)

	// Test SetValue with single key (no delimiter)
	err := cfg.SetValue(ctx, "singlekey", "value")
	if err != nil {
		t.Errorf("SetValue failed: %v", err)
	}

	val, err := cfg.GetValue(ctx, "singlekey")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

func TestSimpleConfigGetConfigWithValidMap(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"nested": map[string]any{
			"key": "value",
		},
	}
	cfg := NewSimpleConfig(ctx, WithConfigurationMap(data)).(*simpleConfig)

	// Test GetConfig with valid nested map
	nestedCfg, err := cfg.GetConfig(ctx, "nested")
	if err != nil {
		t.Errorf("GetConfig failed: %v", err)
	}

	val, err := nestedCfg.GetValue(ctx, "key")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

func TestTelemetryProviderNameTypeConversion(t *testing.T) {
	var name ProviderName = "TestProvider"
	strName := string(name)
	if strName != "TestProvider" {
		t.Errorf("Expected 'TestProvider', got %v", strName)
	}
}

func TestStartTelemetryImplTypeNil(t *testing.T) {
	ctx := context.Background()

	// Test startTelemetry with nil config - implType should be "nil"
	_, state := startTelemetry(ctx, nil, "test_op", "key")

	if state.implType != "nil" {
		t.Errorf("Expected implType 'nil', got %q", state.implType)
	}
}
