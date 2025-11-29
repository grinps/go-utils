package config_test

import (
	"context"
	"testing"

	"github.com/grinps/go-utils/config"
)

func TestConfigFunctions(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"app": map[string]any{
			"port":    8080,
			"name":    "test-app",
			"enabled": true,
		},
		"db": map[string]any{
			"host": "localhost",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test ContextWithConfig and ContextConfig
	ctxWithConfig := config.ContextWithConfig(ctx, cfg)
	retrievedConfig := config.ContextConfig(ctxWithConfig, false)
	if retrievedConfig == nil {
		t.Error("Expected config from context, got nil")
	}

	// Test GetValueE via package functions (using context config)
	var val int
	err := config.GetValueE(ctxWithConfig, "app.port", &val)
	if err != nil {
		t.Errorf("GetValueE failed: %v", err)
	}
	if val != 8080 {
		t.Errorf("Expected 8080, got %v", val)
	}

	// Test GetValueE with string
	var name string
	err = config.GetValueE(ctxWithConfig, "app.name", &name)
	if err != nil {
		t.Errorf("GetValueE failed: %v", err)
	}
	if name != "test-app" {
		t.Errorf("Expected test-app, got %v", name)
	}

	// Test GetValueE with missing key
	var missing string
	err = config.GetValueE(ctxWithConfig, "missing.key", &missing)
	if err == nil {
		t.Error("Expected error for missing key")
	}

	// Test Default()
	defCfg := config.Default()
	if defCfg == nil {
		t.Error("Default config should not be nil")
	}
}

func TestGetValueWithDefault(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	ctxWithConfig := config.ContextWithConfig(ctx, cfg)

	// Test with existing value
	port := 3000 // default
	err := config.GetValueE(ctxWithConfig, "server.port", &port)
	if err != nil {
		t.Errorf("GetValueE failed: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected 8080, got %v", port)
	}

	// Test with missing key (returns error, default is preserved)
	timeout := 30 // default
	err = config.GetValueE(ctxWithConfig, "server.timeout", &timeout)
	if err == nil {
		t.Error("Expected error for missing key")
	}
	// Default value is preserved when error occurs
	if timeout != 30 {
		t.Errorf("Expected default 30, got %v", timeout)
	}

	// Test nil pointer
	err = config.GetValueE[int](ctxWithConfig, "server.port", nil)
	if err == nil {
		t.Error("Expected error for nil pointer")
	}
}

func TestSimpleConfig_SetValue(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx, config.WithDelimiter("."))

	// Test SetValue
	err := cfg.(interface {
		SetValue(context.Context, string, any) error
	}).SetValue(ctx, "server.host", "0.0.0.0")
	if err != nil {
		t.Errorf("SetValue failed: %v", err)
	}

	var host string
	err = cfg.GetValue(ctx, "server.host", &host)
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if host != "0.0.0.0" {
		t.Errorf("Expected 0.0.0.0, got %v", host)
	}

	// Test nested set
	err = cfg.(interface {
		SetValue(context.Context, string, any) error
	}).SetValue(ctx, "nested.deep.value", 123)
	if err != nil {
		t.Errorf("SetValue nested failed: %v", err)
	}

	var val int
	err = cfg.GetValue(ctx, "nested.deep.value", &val)
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if val != 123 {
		t.Errorf("Expected 123, got %v", val)
	}
}

func TestSimpleConfig_Errors(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)

	var val string
	err := cfg.GetValue(ctx, "", &val)
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Test nil return value
	err = cfg.GetValue(ctx, "key", nil)
	if err == nil {
		t.Error("Expected error for nil return value")
	}

	// Test non-pointer return value
	err = cfg.GetValue(ctx, "key", "not a pointer")
	if err == nil {
		t.Error("Expected error for non-pointer return value")
	}

	// Test GetConfig with non-map value
	err = cfg.(interface {
		SetValue(context.Context, string, any) error
	}).SetValue(ctx, "invalid_map", 123)

	_, err = cfg.GetConfig(ctx, "invalid_map")
	if err == nil {
		t.Error("Expected error when getting config from non-map value")
	}
}

func TestErrorCodes(t *testing.T) {
	// Test error instance creation
	errInstance := config.ErrConfigMissingValue.New("missing value", "key", "test.key")

	// Check that error is not nil
	if errInstance == nil {
		t.Error("Expected error instance, got nil")
	}

	// Check error message
	if errInstance.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestGetValueTypeSafety(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"port":   8080,
		"host":   "localhost",
		"active": true,
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test correct type
	var port int
	err := cfg.GetValue(ctx, "port", &port)
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected 8080, got %v", port)
	}

	// Test type mismatch
	var wrongType string
	err = cfg.GetValue(ctx, "port", &wrongType)
	if err == nil {
		t.Error("Expected error for type mismatch")
	}

	// Test bool
	var active bool
	err = cfg.GetValue(ctx, "active", &active)
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if !active {
		t.Error("Expected true")
	}
}

func TestNestedConfiguration(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"database": map[string]any{
			"primary": map[string]any{
				"host": "db1.example.com",
				"port": 5432,
			},
			"replica": map[string]any{
				"host": "db2.example.com",
				"port": 5433,
			},
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test nested access
	var host string
	err := cfg.GetValue(ctx, "database.primary.host", &host)
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if host != "db1.example.com" {
		t.Errorf("Expected db1.example.com, got %v", host)
	}

	// Test GetConfig
	dbCfg, err := cfg.GetConfig(ctx, "database")
	if err != nil {
		t.Errorf("GetConfig failed: %v", err)
	}

	var port int
	err = dbCfg.GetValue(ctx, "primary.port", &port)
	if err != nil {
		t.Errorf("GetValue on sub-config failed: %v", err)
	}
	if port != 5432 {
		t.Errorf("Expected 5432, got %v", port)
	}
}

func TestCoverageContextConfig(t *testing.T) {
	ctx := context.Background()

	// Test with no config in context
	cfg := config.ContextConfig(ctx, false)
	if cfg != nil {
		t.Error("Expected nil config when not in context and defaultIfNotAvailable=false")
	}

	// Test with default
	cfg = config.ContextConfig(ctx, true)
	if cfg == nil {
		t.Error("Expected default config when not in context and defaultIfNotAvailable=true")
	}

	// Test with config in context
	data := map[string]any{"key": "value"}
	customCfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	ctxWithCfg := config.ContextWithConfig(ctx, customCfg)

	retrievedCfg := config.ContextConfig(ctxWithCfg, false)
	if retrievedCfg == nil {
		t.Error("Expected config from context")
	}

	var val string
	err := retrievedCfg.GetValue(ctx, "key", &val)
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if val != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}
}

func TestCoverageGetAsMap(t *testing.T) {
	ctx := context.Background()

	// Test map[string]any
	input1 := map[string]any{"key": "value"}
	result1, err := config.GetAsMap(ctx, input1)
	if err != nil {
		t.Errorf("GetAsMap failed: %v", err)
	}
	if result1["key"] != "value" {
		t.Error("Expected value to be preserved")
	}

	// Test map[string]string
	input2 := map[string]string{"key": "value"}
	result2, err := config.GetAsMap(ctx, input2)
	if err != nil {
		t.Errorf("GetAsMap failed: %v", err)
	}
	if result2["key"] != "value" {
		t.Error("Expected value to be converted")
	}

	// Test map[any]any with string keys
	input3 := map[any]any{"key": "value"}
	result3, err := config.GetAsMap(ctx, input3)
	if err != nil {
		t.Errorf("GetAsMap failed: %v", err)
	}
	if result3["key"] != "value" {
		t.Error("Expected value to be converted")
	}

	// Test map[any]any with non-string key
	input4 := map[any]any{123: "value"}
	_, err = config.GetAsMap(ctx, input4)
	if err == nil {
		t.Error("Expected error for non-string key")
	}

	// Test nil
	result5, err := config.GetAsMap(ctx, nil)
	if err != nil {
		t.Errorf("GetAsMap failed: %v", err)
	}
	if result5 != nil {
		t.Error("Expected nil result for nil input")
	}

	// Test unsupported type
	_, err = config.GetAsMap(ctx, "string")
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

func TestGetValueE(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"port": 8080,
		"host": "localhost",
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	ctxWithCfg := config.ContextWithConfig(ctx, cfg)

	// Test successful retrieval
	var port int
	err := config.GetValueE(ctxWithCfg, "port", &port)
	if err != nil {
		t.Errorf("GetValueE failed: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected 8080, got %v", port)
	}

	// Test missing key
	var missing string
	err = config.GetValueE(ctxWithCfg, "missing", &missing)
	if err == nil {
		t.Error("Expected error for missing key")
	}

	// Test type mismatch
	var wrongType string
	err = config.GetValueE(ctxWithCfg, "port", &wrongType)
	if err == nil {
		t.Error("Expected error for type mismatch")
	}
}

func TestContextWithConfigNil(t *testing.T) {
	// Test with nil context
	cfg := config.NewSimpleConfig(context.Background())
	ctx := config.ContextWithConfig(nil, cfg)
	if ctx == nil {
		t.Error("Expected non-nil context")
	}

	retrieved := config.ContextConfig(ctx, false)
	if retrieved == nil {
		t.Error("Expected config to be stored in context")
	}
}

func TestSetValueEdgeCases(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)

	// Test SetValue with empty key
	err := cfg.(interface {
		SetValue(context.Context, string, any) error
	}).SetValue(ctx, "", "value")
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Test SetValue creating nested maps
	err = cfg.(interface {
		SetValue(context.Context, string, any) error
	}).SetValue(ctx, "a.b.c.d", "deep")
	if err != nil {
		t.Errorf("SetValue failed: %v", err)
	}

	var val string
	err = cfg.GetValue(ctx, "a.b.c.d", &val)
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if val != "deep" {
		t.Errorf("Expected 'deep', got %v", val)
	}

	// Test SetValue with existing non-map intermediate value
	err = cfg.(interface {
		SetValue(context.Context, string, any) error
	}).SetValue(ctx, "scalar", "value")
	if err != nil {
		t.Errorf("SetValue failed: %v", err)
	}

	err = cfg.(interface {
		SetValue(context.Context, string, any) error
	}).SetValue(ctx, "scalar.nested", "value")
	if err == nil {
		t.Error("Expected error when trying to set nested value on scalar")
	}
}

func TestGetValueEdgeCases(t *testing.T) {
	ctx := context.Background()

	// Test with nil config
	var nilCfg config.SimpleConfig
	var val string
	err := nilCfg.GetValue(ctx, "key", &val)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Test with nested nil value
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"parent": map[string]any{
			"child": nil,
		},
	}))

	err = cfg.GetValue(ctx, "parent.child.grandchild", &val)
	if err == nil {
		t.Error("Expected error for nil intermediate value")
	}

	// Test with non-map intermediate value
	cfg2 := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"scalar": "value",
	}))

	err = cfg2.GetValue(ctx, "scalar.nested", &val)
	if err == nil {
		t.Error("Expected error for non-map intermediate value")
	}
}

func TestGetConfigEdgeCases(t *testing.T) {
	ctx := context.Background()

	// Test GetConfig with nil value
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"key": nil,
	}))

	_, err := cfg.GetConfig(ctx, "key")
	if err == nil {
		t.Error("Expected error for nil value in GetConfig")
	}

	// Test GetConfig with missing key
	_, err = cfg.GetConfig(ctx, "missing")
	if err == nil {
		t.Error("Expected error for missing key in GetConfig")
	}
}

func TestCustomDelimiter(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
		},
	}

	// Test with custom delimiter
	cfg := config.NewSimpleConfig(ctx,
		config.WithConfigurationMap(data),
		config.WithDelimiter("/"))

	var port int
	err := cfg.GetValue(ctx, "server/port", &port)
	if err != nil {
		t.Errorf("GetValue with custom delimiter failed: %v", err)
	}
	if port != 8080 {
		t.Errorf("Expected 8080, got %v", port)
	}

	// Test that dot notation doesn't work with custom delimiter
	err = cfg.GetValue(ctx, "server.port", &port)
	if err == nil {
		t.Error("Expected error when using wrong delimiter")
	}
}
