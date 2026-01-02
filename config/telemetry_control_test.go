package config_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/grinps/go-utils/config"
)

func TestSetTelemetryEnabled(t *testing.T) {
	// Save original state
	originalState := config.IsTelemetryEnabled()
	defer config.SetTelemetryEnabled(originalState)

	// Test enabling telemetry
	config.SetTelemetryEnabled(true)
	if !config.IsTelemetryEnabled() {
		t.Error("Expected telemetry to be enabled")
	}

	// Test disabling telemetry
	config.SetTelemetryEnabled(false)
	if config.IsTelemetryEnabled() {
		t.Error("Expected telemetry to be disabled")
	}

	// Re-enable for subsequent tests
	config.SetTelemetryEnabled(true)
	if !config.IsTelemetryEnabled() {
		t.Error("Expected telemetry to be re-enabled")
	}
}

func TestTelemetryDisabledSkipsInstrumentation(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Save original state and disable telemetry
	originalState := config.IsTelemetryEnabled()
	defer config.SetTelemetryEnabled(originalState)

	config.SetTelemetryEnabled(false)

	// Operations should still work when telemetry is disabled
	var port int
	err := config.GetValueWithConfig(ctx, cfg, "server.port", &port)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected 8080, got %v", port)
	}

	// Test SetValue with telemetry disabled
	err = config.SetValueWithConfig(ctx, cfg, "server.host", "localhost")
	if err != nil {
		t.Errorf("SetValueWithConfig failed: %v", err)
	}

	// Test GetConfig with telemetry disabled
	serverCfg, err := config.GetConfigWithConfig(ctx, cfg, "server")
	if err != nil {
		t.Errorf("GetConfigWithConfig failed: %v", err)
	}
	if serverCfg == nil {
		t.Error("Expected non-nil config")
	}
}

// telemetryAwareConfig is a test config that implements TelemetryAware
type telemetryAwareConfig struct {
	data             map[string]any
	shouldInstrument bool
	customAttrs      []any
}

func (c *telemetryAwareConfig) Name() config.ProviderName {
	return "TelemetryAwareConfig"
}

func (c *telemetryAwareConfig) GetValue(ctx context.Context, key string) (any, error) {
	if c.data == nil {
		return nil, errors.New("nil data")
	}
	parts := strings.Split(key, ".")
	var current any = c.data
	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return nil, errors.New("key not found")
		}
	}
	if current == nil {
		return nil, errors.New("value not found")
	}
	return current, nil
}

func (c *telemetryAwareConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, err := c.GetValue(ctx, key)
	if err != nil {
		return nil, err
	}
	if m, ok := val.(map[string]any); ok {
		return &telemetryAwareConfig{data: m, shouldInstrument: c.shouldInstrument}, nil
	}
	return nil, errors.New("not a map")
}

func (c *telemetryAwareConfig) ShouldInstrument(ctx context.Context, key string, op string) bool {
	return c.shouldInstrument
}

func (c *telemetryAwareConfig) GenerateTelemetryAttributes(ctx context.Context, op string, attrs []any) []any {
	if c.customAttrs != nil {
		return append(attrs, c.customAttrs...)
	}
	return attrs
}

func TestTelemetryAwareInterface(t *testing.T) {
	ctx := context.Background()

	// Test with TelemetryAware config that opts out
	cfgOptOut := &telemetryAwareConfig{
		data:             map[string]any{"key": "value"},
		shouldInstrument: false,
	}

	// Should work but skip telemetry
	var val string
	err := config.GetValueWithConfig(ctx, cfgOptOut, "key", &val)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}

	// Test with TelemetryAware config that opts in
	cfgOptIn := &telemetryAwareConfig{
		data:             map[string]any{"server": map[string]any{"port": 9090}},
		shouldInstrument: true,
		customAttrs:      []any{"custom.attr", "custom-value"},
	}

	var port int
	err = config.GetValueWithConfig(ctx, cfgOptIn, "server.port", &port)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}
	if port != 9090 {
		t.Errorf("Expected 9090, got %v", port)
	}
}

func TestTelemetryWithErrors(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"key": "value",
	}))

	// Test error paths that trigger finishTelemetry with errors

	// Missing key error
	var val string
	err := config.GetValueWithConfig(ctx, cfg, "missing.key", &val)
	if err == nil {
		t.Error("Expected error for missing key")
	}

	// Empty key error
	err = config.GetValueWithConfig(ctx, cfg, "", &val)
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Nil config error
	err = config.GetValueWithConfig[string](ctx, nil, "key", &val)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Type mismatch error
	data := map[string]any{"port": 8080}
	cfg2 := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	var strVal string
	err = config.GetValueWithConfig(ctx, cfg2, "port", &strVal)
	if err == nil {
		t.Error("Expected error for type mismatch")
	}
}

func TestTelemetryWithUnmarshal(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)

	type TestStruct struct {
		Name string `config:"name"`
	}

	// Test unmarshal with non-MarshableConfig (triggers telemetry with error)
	var ts TestStruct
	err := config.UnmarshalWithConfig(ctx, cfg, "test", &ts)
	if err == nil {
		t.Error("Expected error for non-MarshableConfig")
	}
}

func TestTelemetryWithSetValue(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)

	// Successful SetValue
	err := config.SetValueWithConfig(ctx, cfg, "new.key", "new-value")
	if err != nil {
		t.Errorf("SetValueWithConfig failed: %v", err)
	}

	// Verify value was set
	val, err := cfg.GetValue(ctx, "new.key")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if val != "new-value" {
		t.Errorf("Expected 'new-value', got %v", val)
	}
}

func TestTelemetryWithGetConfig(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"database": map[string]any{
			"host": "localhost",
			"port": 5432,
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Successful GetConfig
	dbCfg, err := config.GetConfigWithConfig(ctx, cfg, "database")
	if err != nil {
		t.Errorf("GetConfigWithConfig failed: %v", err)
	}

	val, err := dbCfg.GetValue(ctx, "host")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if val != "localhost" {
		t.Errorf("Expected 'localhost', got %v", val)
	}
}

func TestTelemetryAwareWithNilConfig(t *testing.T) {
	ctx := context.Background()

	// Test with nil config - should handle gracefully
	var val string
	err := config.GetValueWithConfig[string](ctx, nil, "key", &val)
	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestSimpleConfigName(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)

	// Verify Name() returns expected value
	name := cfg.Name()
	if name != "SimpleConfig" {
		t.Errorf("Expected 'SimpleConfig', got %v", name)
	}
}

// configWithoutName is a config that doesn't implement Name() properly
// to test the fallback path in startTelemetry
type configWithoutName struct {
	data map[string]any
}

func (c *configWithoutName) Name() config.ProviderName {
	return "ConfigWithoutName"
}

func (c *configWithoutName) GetValue(ctx context.Context, key string) (any, error) {
	if val, ok := c.data[key]; ok {
		return val, nil
	}
	return nil, errors.New("not found")
}

func (c *configWithoutName) GetConfig(ctx context.Context, key string) (config.Config, error) {
	return nil, errors.New("not implemented")
}

func TestConfigWithName(t *testing.T) {
	ctx := context.Background()
	cfg := &configWithoutName{data: map[string]any{"key": "value"}}

	var val string
	err := config.GetValueWithConfig(ctx, cfg, "key", &val)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

// nonMutableConfig is a config that doesn't implement MutableConfig
type nonMutableConfig struct {
	data map[string]any
}

func (c *nonMutableConfig) Name() config.ProviderName {
	return "NonMutableConfig"
}

func (c *nonMutableConfig) GetValue(ctx context.Context, key string) (any, error) {
	if val, ok := c.data[key]; ok {
		return val, nil
	}
	return nil, errors.New("not found")
}

func (c *nonMutableConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	return nil, errors.New("not implemented")
}

func TestSetValueWithNonMutableConfig(t *testing.T) {
	ctx := context.Background()
	cfg := &nonMutableConfig{data: map[string]any{"key": "value"}}

	// SetValueWithConfig should fail for non-MutableConfig
	err := config.SetValueWithConfig(ctx, cfg, "key", "new-value")
	if err == nil {
		t.Error("Expected error for non-MutableConfig")
	}
}

func TestUnmarshalWithNonPointerTarget(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)

	// Test with various invalid targets
	type TestStruct struct {
		Name string
	}

	// The function requires a pointer, so we test with nil
	err := config.UnmarshalWithConfig[TestStruct](ctx, cfg, "test", nil)
	if err == nil {
		t.Error("Expected error for nil target")
	}
}

func TestTelemetryProviderNameType(t *testing.T) {
	// Test TelemetryProviderName type
	var name config.ProviderName = "TestProvider"
	if string(name) != "TestProvider" {
		t.Errorf("Expected 'TestProvider', got %v", name)
	}
}

func TestTelemetryWithNestedKeys(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"deeply": map[string]any{
			"nested": map[string]any{
				"config": map[string]any{
					"value": "found",
				},
			},
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test with deeply nested key (tests extractKeyPrefix)
	var val string
	err := config.GetValueWithConfig(ctx, cfg, "deeply.nested.config.value", &val)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}
	if val != "found" {
		t.Errorf("Expected 'found', got %v", val)
	}
}

func TestTelemetryWithSinglePartKey(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"simple": "value",
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test with single part key (no dots)
	var val string
	err := config.GetValueWithConfig(ctx, cfg, "simple", &val)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

func TestTelemetryAwareWithCustomAttributes(t *testing.T) {
	ctx := context.Background()

	cfgWithAttrs := &telemetryAwareConfig{
		data: map[string]any{
			"key1": "value1",
			"key2": "value2",
		},
		shouldInstrument: true,
		customAttrs:      []any{"source", "test", "version", "1.0"},
	}

	// Multiple operations to exercise telemetry paths
	var val1 string
	err := config.GetValueWithConfig(ctx, cfgWithAttrs, "key1", &val1)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}

	var val2 string
	err = config.GetValueWithConfig(ctx, cfgWithAttrs, "key2", &val2)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}
}

func TestMultipleTelemetryOperations(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}))

	// Exercise multiple operations in sequence
	var host string
	_ = config.GetValueWithConfig(ctx, cfg, "server.host", &host)

	var port int
	_ = config.GetValueWithConfig(ctx, cfg, "server.port", &port)

	_ = config.SetValueWithConfig(ctx, cfg, "server.timeout", 30)

	serverCfg, _ := config.GetConfigWithConfig(ctx, cfg, "server")
	if serverCfg != nil {
		var timeout int
		_ = config.GetValueWithConfig(ctx, serverCfg, "timeout", &timeout)
	}
}

func TestTelemetryErrorCodeExtraction(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)

	// Trigger various error types that have error codes
	var val string

	// Missing key error
	err := config.GetValueWithConfig(ctx, cfg, "missing.key", &val)
	if err == nil {
		t.Error("Expected error")
	}

	// Empty key error
	err = config.GetValueWithConfig(ctx, cfg, "", &val)
	if err == nil {
		t.Error("Expected error")
	}

	// SetValue errors
	err = config.SetValueWithConfig(ctx, cfg, "", "value")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestUnmarshalTelemetryPaths(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"data": map[string]any{
			"name": "test",
		},
	}))

	type TestStruct struct {
		Name string `config:"name"`
	}

	// Test all error paths in UnmarshalWithConfig
	var ts TestStruct

	// Non-MarshableConfig error
	err := config.UnmarshalWithConfig(ctx, cfg, "data", &ts)
	if err == nil {
		t.Error("Expected error for non-MarshableConfig")
	}

	// Nil target
	err = config.UnmarshalWithConfig[TestStruct](ctx, cfg, "data", nil)
	if err == nil {
		t.Error("Expected error for nil target")
	}

	// Non-struct pointer
	var num int
	err = config.UnmarshalWithConfig(ctx, cfg, "data", &num)
	if err == nil {
		t.Error("Expected error for non-struct pointer")
	}

	// Nil config
	err = config.UnmarshalWithConfig(ctx, nil, "data", &ts)
	if err == nil {
		t.Error("Expected error for nil config")
	}
}
