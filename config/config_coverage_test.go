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

	val, err := cfg.GetValue(ctx, "server.host")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	host, ok := val.(string)
	if !ok || host != "0.0.0.0" {
		t.Errorf("Expected 0.0.0.0, got %v", val)
	}

	// Test nested set
	err = cfg.(interface {
		SetValue(context.Context, string, any) error
	}).SetValue(ctx, "nested.deep.value", 123)
	if err != nil {
		t.Errorf("SetValue nested failed: %v", err)
	}

	val, err = cfg.GetValue(ctx, "nested.deep.value")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	intVal, ok := val.(int)
	if !ok || intVal != 123 {
		t.Errorf("Expected 123, got %v", val)
	}
}

func TestSimpleConfig_Errors(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)

	_, err := cfg.GetValue(ctx, "")
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Test GetConfig with non-map value
	err = cfg.(interface {
		SetValue(context.Context, string, any) error
	}).SetValue(ctx, "invalid_map", 123)
	if err != nil {
		t.Errorf("SetValue failed: %v", err)
	}

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
	val, err := cfg.GetValue(ctx, "port")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	port, ok := val.(int)
	if !ok || port != 8080 {
		t.Errorf("Expected 8080, got %v", val)
	}

	// Test bool
	val, err = cfg.GetValue(ctx, "active")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	active, ok := val.(bool)
	if !ok || !active {
		t.Errorf("Expected true, got %v", val)
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
	val, err := cfg.GetValue(ctx, "database.primary.host")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	host, ok := val.(string)
	if !ok || host != "db1.example.com" {
		t.Errorf("Expected db1.example.com, got %v", val)
	}

	// Test GetConfig
	dbCfg, err := cfg.GetConfig(ctx, "database")
	if err != nil {
		t.Errorf("GetConfig failed: %v", err)
	}

	val, err = dbCfg.GetValue(ctx, "primary.port")
	if err != nil {
		t.Errorf("GetValue on sub-config failed: %v", err)
	}
	port, ok := val.(int)
	if !ok || port != 5432 {
		t.Errorf("Expected 5432, got %v", val)
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

	val, err := retrievedCfg.GetValue(ctx, "key")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	strVal, ok := val.(string)
	if !ok || strVal != "value" {
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

	val, err := cfg.GetValue(ctx, "a.b.c.d")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	strVal, ok := val.(string)
	if !ok || strVal != "deep" {
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
	_, err := nilCfg.GetValue(ctx, "key")
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Test with nested nil value
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"parent": map[string]any{
			"child": nil,
		},
	}))

	_, err = cfg.GetValue(ctx, "parent.child.grandchild")
	if err == nil {
		t.Error("Expected error for nil intermediate value")
	}

	// Test with non-map intermediate value
	cfg2 := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"scalar": "value",
	}))

	_, err = cfg2.GetValue(ctx, "scalar.nested")
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

	val, err := cfg.GetValue(ctx, "server/port")
	if err != nil {
		t.Errorf("GetValue with custom delimiter failed: %v", err)
	}
	port, ok := val.(int)
	if !ok || port != 8080 {
		t.Errorf("Expected 8080, got %v", val)
	}

	// Test that dot notation doesn't work with custom delimiter
	_, err = cfg.GetValue(ctx, "server.port")
	if err == nil {
		t.Error("Expected error when using wrong delimiter")
	}
}

func TestUnmarshalFunctions(t *testing.T) {
	ctx := context.Background()

	// Test Unmarshal with non-MarshableConfig
	simpleCfg := config.NewSimpleConfig(ctx)
	ctxWithCfg := config.ContextWithConfig(ctx, simpleCfg)

	type TestStruct struct {
		Name string `config:"name"`
	}
	var ts TestStruct
	err := config.Unmarshal(ctxWithCfg, "test", &ts)
	if err == nil {
		t.Error("Expected error for non-MarshableConfig")
	}

	// Test UnmarshalWithConfig with nil config
	err = config.UnmarshalWithConfig(ctx, nil, "test", &ts)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Test UnmarshalWithConfig with non-pointer
	err = config.UnmarshalWithConfig[TestStruct](ctx, simpleCfg, "test", &ts)
	if err == nil {
		t.Error("Expected error for non-MarshableConfig")
	}

	// Test UnmarshalWithConfig with nil target
	err = config.UnmarshalWithConfig[TestStruct](ctx, simpleCfg, "test", nil)
	if err == nil {
		t.Error("Expected error for nil target")
	}

	// Test UnmarshalWithConfig with non-struct pointer
	var num int
	err = config.UnmarshalWithConfig(ctx, simpleCfg, "test", &num)
	if err == nil {
		t.Error("Expected error for non-struct pointer")
	}
}

func TestSetValueFunctions(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)
	ctxWithCfg := config.ContextWithConfig(ctx, cfg)

	// Test SetValue
	err := config.SetValue(ctxWithCfg, "test.key", "value")
	if err != nil {
		t.Errorf("SetValue failed: %v", err)
	}

	val, err := cfg.GetValue(ctx, "test.key")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	strVal, ok := val.(string)
	if !ok || strVal != "value" {
		t.Errorf("Expected 'value', got %v", val)
	}

	// Test SetValueWithConfig with nil config
	err = config.SetValueWithConfig(ctx, nil, "key", "value")
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Test SetValueWithConfig with empty key
	err = config.SetValueWithConfig(ctx, cfg, "", "value")
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Test SetValueWithConfig
	err = config.SetValueWithConfig(ctx, cfg, "another.key", 123)
	if err != nil {
		t.Errorf("SetValueWithConfig failed: %v", err)
	}

	val, err = cfg.GetValue(ctx, "another.key")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	intVal, ok := val.(int)
	if !ok || intVal != 123 {
		t.Errorf("Expected 123, got %v", val)
	}
}

func TestMustUnmarshal(t *testing.T) {
	ctx := context.Background()
	simpleCfg := config.NewSimpleConfig(ctx)
	ctxWithCfg := config.ContextWithConfig(ctx, simpleCfg)

	type TestStruct struct {
		Name string `config:"name"`
	}

	// Test MustUnmarshal with error - should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for MustUnmarshal with non-MarshableConfig")
		}
	}()

	var ts TestStruct
	config.MustUnmarshal(ctxWithCfg, "test", &ts)
}

func TestGetValueENilValue(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"key": nil,
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	ctxWithCfg := config.ContextWithConfig(ctx, cfg)

	var val string
	err := config.GetValueE(ctxWithCfg, "key", &val)
	if err == nil {
		t.Error("Expected error for nil value")
	}
}

func TestSetValueNilConfig(t *testing.T) {
	ctx := context.Background()
	var nilCfg config.SimpleConfig

	err := nilCfg.SetValue(ctx, "key", "value")
	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestSetAsDefault(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"custom": map[string]any{
			"key": "custom-value",
		},
	}
	customCfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Save original default to restore later
	originalDefault := config.Default()
	defer config.SetAsDefault(originalDefault)

	// Set custom config as default
	config.SetAsDefault(customCfg)

	// Verify default was changed
	newDefault := config.Default()
	val, err := newDefault.GetValue(ctx, "custom.key")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	strVal, ok := val.(string)
	if !ok || strVal != "custom-value" {
		t.Errorf("Expected 'custom-value', got %v", val)
	}

	// Test that ContextConfig uses new default when no config in context
	cfg := config.ContextConfig(ctx, true)
	val, err = cfg.GetValue(ctx, "custom.key")
	if err != nil {
		t.Errorf("GetValue from context default failed: %v", err)
	}
	strVal, ok = val.(string)
	if !ok || strVal != "custom-value" {
		t.Errorf("Expected 'custom-value', got %v", val)
	}
}

func TestGetValueWithConfig(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 9090,
			"host": "example.com",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test successful retrieval
	var port int
	err := config.GetValueWithConfig(ctx, cfg, "server.port", &port)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}
	if port != 9090 {
		t.Errorf("Expected 9090, got %v", port)
	}

	// Test with string
	var host string
	err = config.GetValueWithConfig(ctx, cfg, "server.host", &host)
	if err != nil {
		t.Errorf("GetValueWithConfig failed: %v", err)
	}
	if host != "example.com" {
		t.Errorf("Expected 'example.com', got %v", host)
	}

	// Test with nil config
	err = config.GetValueWithConfig[int](ctx, nil, "server.port", &port)
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Test with nil return value
	err = config.GetValueWithConfig[int](ctx, cfg, "server.port", nil)
	if err == nil {
		t.Error("Expected error for nil return value")
	}

	// Test with empty key
	err = config.GetValueWithConfig(ctx, cfg, "", &port)
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Test with missing key
	var missing string
	err = config.GetValueWithConfig(ctx, cfg, "missing.key", &missing)
	if err == nil {
		t.Error("Expected error for missing key")
	}

	// Test type mismatch
	var wrongType string
	err = config.GetValueWithConfig(ctx, cfg, "server.port", &wrongType)
	if err == nil {
		t.Error("Expected error for type mismatch")
	}
}

func TestGetConfig(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"database": map[string]any{
			"primary": map[string]any{
				"host": "primary.db.com",
				"port": 5432,
			},
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	ctxWithCfg := config.ContextWithConfig(ctx, cfg)

	// Test successful retrieval
	dbCfg, err := config.GetConfig(ctxWithCfg, "database")
	if err != nil {
		t.Errorf("GetConfig failed: %v", err)
	}

	val, err := dbCfg.GetValue(ctx, "primary.host")
	if err != nil {
		t.Errorf("GetValue on sub-config failed: %v", err)
	}
	host, ok := val.(string)
	if !ok || host != "primary.db.com" {
		t.Errorf("Expected 'primary.db.com', got %v", val)
	}

	// Test with missing key
	_, err = config.GetConfig(ctxWithCfg, "missing")
	if err == nil {
		t.Error("Expected error for missing key")
	}
}

func TestGetConfigWithConfig(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"services": map[string]any{
			"api": map[string]any{
				"endpoint": "https://api.example.com",
				"timeout":  30,
			},
		},
		"scalar": "not-a-map",
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test successful retrieval
	apiCfg, err := config.GetConfigWithConfig(ctx, cfg, "services.api")
	if err != nil {
		t.Errorf("GetConfigWithConfig failed: %v", err)
	}

	val, err := apiCfg.GetValue(ctx, "endpoint")
	if err != nil {
		t.Errorf("GetValue on sub-config failed: %v", err)
	}
	endpoint, ok := val.(string)
	if !ok || endpoint != "https://api.example.com" {
		t.Errorf("Expected 'https://api.example.com', got %v", val)
	}

	// Test with nil config
	_, err = config.GetConfigWithConfig(ctx, nil, "services")
	if err == nil {
		t.Error("Expected error for nil config")
	}

	// Test with empty key
	_, err = config.GetConfigWithConfig(ctx, cfg, "")
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Test with missing key
	_, err = config.GetConfigWithConfig(ctx, cfg, "missing.key")
	if err == nil {
		t.Error("Expected error for missing key")
	}

	// Test with non-map value
	_, err = config.GetConfigWithConfig(ctx, cfg, "scalar")
	if err == nil {
		t.Error("Expected error for non-map value")
	}
}

func TestGetValueEWithEmptyKey(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)
	ctxWithCfg := config.ContextWithConfig(ctx, cfg)

	var val string
	err := config.GetValueE(ctxWithCfg, "", &val)
	if err == nil {
		t.Error("Expected error for empty key")
	}
}

func TestContextConfigNilContext(t *testing.T) {
	// Test with nil context and defaultIfNotAvailable=true
	cfg := config.ContextConfig(nil, true)
	if cfg == nil {
		t.Error("Expected default config when context is nil and defaultIfNotAvailable=true")
	}

	// Test with nil context and defaultIfNotAvailable=false
	cfg = config.ContextConfig(nil, false)
	if cfg != nil {
		t.Error("Expected nil config when context is nil and defaultIfNotAvailable=false")
	}
}

func TestSimpleConfig_All(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
			"host": "localhost",
		},
		"app": map[string]any{
			"name": "test-app",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test All() returns the configuration map
	allGetter, ok := cfg.(config.AllGetter)
	if !ok {
		t.Fatal("SimpleConfig should implement AllGetter")
	}

	all := allGetter.All(ctx)
	if all == nil {
		t.Error("Expected non-nil map from All()")
	}

	// Verify the data is correct
	server, ok := all["server"].(map[string]any)
	if !ok {
		t.Error("Expected server to be a map")
	}
	if server["port"] != 8080 {
		t.Errorf("Expected port 8080, got %v", server["port"])
	}

	// Test All() with nil config
	var nilCfg config.SimpleConfig
	nilAll := nilCfg.All(ctx)
	if nilAll != nil {
		t.Error("Expected nil from All() on nil config")
	}
}

func TestSimpleConfig_Keys(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
			"host": "localhost",
		},
		"app": map[string]any{
			"name":    "test-app",
			"version": "1.0.0",
		},
		"debug": true,
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test Keys() returns all keys
	keysProvider, ok := cfg.(config.AllKeysProvider)
	if !ok {
		t.Fatal("SimpleConfig should implement AllKeysProvider")
	}

	allKeys := keysProvider.Keys("")
	if allKeys == nil {
		t.Error("Expected non-nil keys from Keys()")
	}

	// Should have keys for: server, server.port, server.host, app, app.name, app.version, debug
	expectedMinKeys := 7
	if len(allKeys) < expectedMinKeys {
		t.Errorf("Expected at least %d keys, got %d: %v", expectedMinKeys, len(allKeys), allKeys)
	}

	// Test Keys() with prefix
	serverKeys := keysProvider.Keys("server")
	if serverKeys == nil {
		t.Error("Expected non-nil keys from Keys('server')")
	}

	// Should have: server, server.port, server.host
	hasServerPort := false
	hasServerHost := false
	for _, k := range serverKeys {
		if k == "server.port" {
			hasServerPort = true
		}
		if k == "server.host" {
			hasServerHost = true
		}
	}
	if !hasServerPort || !hasServerHost {
		t.Errorf("Expected server.port and server.host in keys, got %v", serverKeys)
	}

	// Test Keys() with nil config
	var nilCfg config.SimpleConfig
	nilKeys := nilCfg.Keys("")
	if nilKeys != nil {
		t.Error("Expected nil from Keys() on nil config")
	}
}

func TestSimpleConfig_Delete(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port":  8080,
			"host":  "localhost",
			"debug": true,
		},
		"app": map[string]any{
			"name": "test-app",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	// Test Delete() interface
	deleter, ok := cfg.(config.Deleter)
	if !ok {
		t.Fatal("SimpleConfig should implement Deleter")
	}

	// Delete a nested key
	err := deleter.Delete("server.debug")
	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	// Verify key is deleted
	_, err = cfg.GetValue(ctx, "server.debug")
	if err == nil {
		t.Error("Expected error when getting deleted key")
	}

	// Verify other keys still exist
	val, err := cfg.GetValue(ctx, "server.port")
	if err != nil {
		t.Errorf("GetValue failed: %v", err)
	}
	if val.(int) != 8080 {
		t.Errorf("Expected 8080, got %v", val)
	}

	// Delete a top-level key
	err = deleter.Delete("app")
	if err != nil {
		t.Errorf("Delete top-level key failed: %v", err)
	}

	// Verify key is deleted
	_, err = cfg.GetValue(ctx, "app.name")
	if err == nil {
		t.Error("Expected error when getting deleted key")
	}

	// Test Delete() with empty key
	err = deleter.Delete("")
	if err == nil {
		t.Error("Expected error for empty key")
	}

	// Test Delete() with missing key
	err = deleter.Delete("nonexistent.key")
	if err == nil {
		t.Error("Expected error for missing key")
	}

	// Test Delete() with nil config
	var nilCfg config.SimpleConfig
	err = nilCfg.Delete("key")
	if err == nil {
		t.Error("Expected error for nil config")
	}
}

func TestSimpleConfig_Delete_NestedMissing(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

	deleter := cfg.(config.Deleter)

	// Test Delete() with missing intermediate key
	err := deleter.Delete("missing.nested.key")
	if err == nil {
		t.Error("Expected error for missing intermediate key")
	}

	// Test Delete() with non-map intermediate value
	cfg2 := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"scalar": "value",
	}))
	deleter2 := cfg2.(config.Deleter)

	err = deleter2.Delete("scalar.nested")
	if err == nil {
		t.Error("Expected error for non-map intermediate value")
	}
}

func TestSimpleConfig_InterfaceCompileCheck(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)

	// Runtime interface checks (NewSimpleConfig returns config.Config interface)
	if _, ok := cfg.(config.MutableConfig); !ok {
		t.Error("SimpleConfig should implement MutableConfig")
	}
	if _, ok := cfg.(config.AllGetter); !ok {
		t.Error("SimpleConfig should implement AllGetter")
	}
	if _, ok := cfg.(config.AllKeysProvider); !ok {
		t.Error("SimpleConfig should implement AllKeysProvider")
	}
	if _, ok := cfg.(config.Deleter); !ok {
		t.Error("SimpleConfig should implement Deleter")
	}
}
