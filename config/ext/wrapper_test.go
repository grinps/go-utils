package ext_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/grinps/go-utils/config"
	"github.com/grinps/go-utils/config/ext"
)

// WrapperServerConfig for wrapper tests (avoid redeclaration)
type WrapperServerConfig struct {
	Host    string        `config:"host"`
	Port    int           `config:"port"`
	Timeout time.Duration `config:"timeout"`
}

// mockReadOnlyConfig implements only config.Config (not MutableConfig)
type mockReadOnlyConfig struct {
	data map[string]any
}

func newMockReadOnlyConfig(data map[string]any) *mockReadOnlyConfig {
	return &mockReadOnlyConfig{data: data}
}

func (m *mockReadOnlyConfig) Name() config.ProviderName {
	return "mockReadOnlyConfig"
}

func (m *mockReadOnlyConfig) GetValue(ctx context.Context, key string) (any, error) {
	if key == "" {
		return nil, config.ErrConfigEmptyKey.New("empty key")
	}
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	return val, nil
}

func (m *mockReadOnlyConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if subMap, ok := val.(map[string]any); ok {
		return newMockReadOnlyConfig(subMap), nil
	}
	return nil, config.ErrConfigInvalidValue.New("not a map", "key", key)
}

// mockMutableConfig implements both config.Config and ext.MutableConfig
type mockMutableConfig struct {
	data map[string]any
}

func newMockMutableConfig(data map[string]any) *mockMutableConfig {
	return &mockMutableConfig{data: data}
}

func (m *mockMutableConfig) Name() config.ProviderName {
	return "mockMutableConfig"
}

func (m *mockMutableConfig) GetValue(ctx context.Context, key string) (any, error) {
	if key == "" {
		return nil, config.ErrConfigEmptyKey.New("empty key")
	}
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	return val, nil
}

func (m *mockMutableConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if subMap, ok := val.(map[string]any); ok {
		return newMockMutableConfig(subMap), nil
	}
	return nil, config.ErrConfigInvalidValue.New("not a map", "key", key)
}

func (m *mockMutableConfig) SetValue(ctx context.Context, key string, newValue any) error {
	if key == "" {
		return ext.ErrExtSetValueFailed.New("empty key")
	}
	m.data[key] = newValue
	return nil
}

func (m *mockMutableConfig) All(ctx context.Context) map[string]any {
	return m.data
}

// mockMarshableConfig implements both config.Config and ext.MarshableConfig
type mockMarshableConfig struct {
	data           map[string]any
	unmarshalCalls int
}

func newMockMarshableConfig(data map[string]any) *mockMarshableConfig {
	return &mockMarshableConfig{data: data}
}

func (m *mockMarshableConfig) Name() config.ProviderName {
	return "mockMarshableConfig"
}

func (m *mockMarshableConfig) GetValue(ctx context.Context, key string) (any, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	return val, nil
}

func (m *mockMarshableConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if subMap, ok := val.(map[string]any); ok {
		return newMockMarshableConfig(subMap), nil
	}
	return nil, config.ErrConfigInvalidValue.New("not a map", "key", key)
}

func (m *mockMarshableConfig) Unmarshal(ctx context.Context, key string, target any, options ...any) error {
	m.unmarshalCalls++
	// Simplified implementation - just check it's called
	return nil
}

func (m *mockMarshableConfig) All(ctx context.Context) map[string]any {
	return m.data
}

// Tests for ConfigWrapper

func TestNewConfigWrapper(t *testing.T) {
	t.Run("with nil config", func(t *testing.T) {
		wrapper := ext.NewConfigWrapper(nil)
		if wrapper != nil {
			t.Error("Expected nil wrapper for nil config")
		}
	})

	t.Run("with valid config", func(t *testing.T) {
		cfg := config.NewSimpleConfig(context.Background())
		wrapper := ext.NewConfigWrapper(cfg)
		if wrapper == nil {
			t.Error("Expected non-nil wrapper")
		}
	})
}

func TestConfigWrapper_Unwrap(t *testing.T) {
	t.Run("unwrap returns original config", func(t *testing.T) {
		ctx := context.Background()
		cfg := config.NewSimpleConfig(ctx)
		wrapper := ext.NewConfigWrapper(cfg)

		unwrapped := wrapper.Unwrap()
		if unwrapped != cfg {
			t.Error("Unwrap should return the original config")
		}
	})

	t.Run("unwrap nil wrapper", func(t *testing.T) {
		var wrapper *ext.ConfigWrapper
		if wrapper.Unwrap() != nil {
			t.Error("Expected nil from nil wrapper")
		}
	})
}

func TestConfigWrapper_GetValue(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"host": "localhost",
		"port": 8080,
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	t.Run("get existing value", func(t *testing.T) {
		val, err := wrapper.GetValue(ctx, "host")
		if err != nil {
			t.Fatalf("GetValue failed: %v", err)
		}
		host, ok := val.(string)
		if !ok || host != "localhost" {
			t.Errorf("Expected 'localhost', got '%v'", val)
		}
	})

	t.Run("get missing value", func(t *testing.T) {
		_, err := wrapper.GetValue(ctx, "missing")
		if err == nil {
			t.Error("Expected error for missing key")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var nilWrapper *ext.ConfigWrapper
		_, err := nilWrapper.GetValue(ctx, "key")
		if err == nil {
			t.Error("Expected error for nil wrapper")
		}
	})
}

func TestConfigWrapper_GetConfig(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	t.Run("get nested config", func(t *testing.T) {
		subCfg, err := wrapper.GetConfig(ctx, "server")
		if err != nil {
			t.Fatalf("GetConfig failed: %v", err)
		}
		if subCfg == nil {
			t.Fatal("Expected non-nil sub-config")
		}

		// Sub-config should also be a wrapper
		val, err := subCfg.GetValue(ctx, "host")
		if err != nil {
			t.Fatalf("GetValue on sub-config failed: %v", err)
		}
		host, ok := val.(string)
		if !ok || host != "localhost" {
			t.Errorf("Expected 'localhost', got '%v'", val)
		}
	})

	t.Run("get missing config", func(t *testing.T) {
		_, err := wrapper.GetConfig(ctx, "missing")
		if err == nil {
			t.Error("Expected error for missing key")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var nilWrapper *ext.ConfigWrapper
		_, err := nilWrapper.GetConfig(ctx, "key")
		if err == nil {
			t.Error("Expected error for nil wrapper")
		}
	})
}

func TestConfigWrapper_Unmarshal(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host":    "localhost",
			"port":    8080,
			"timeout": "30s",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	t.Run("unmarshal nested config", func(t *testing.T) {
		var server WrapperServerConfig
		err := wrapper.Unmarshal(ctx, "server", &server)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if server.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got '%s'", server.Host)
		}
		if server.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", server.Port)
		}
		if server.Timeout != 30*time.Second {
			t.Errorf("Expected timeout 30s, got %v", server.Timeout)
		}
	})

	t.Run("unmarshal with nil target", func(t *testing.T) {
		err := wrapper.Unmarshal(ctx, "server", nil)
		if err == nil {
			t.Error("Expected error for nil target")
		}
	})

	t.Run("unmarshal with non-pointer target", func(t *testing.T) {
		err := wrapper.Unmarshal(ctx, "server", WrapperServerConfig{})
		if err == nil {
			t.Error("Expected error for non-pointer target")
		}
	})

	t.Run("unmarshal with non-struct pointer", func(t *testing.T) {
		var str string
		err := wrapper.Unmarshal(ctx, "server", &str)
		if err == nil {
			t.Error("Expected error for non-struct pointer target")
		}
	})

	t.Run("unmarshal missing key", func(t *testing.T) {
		var server WrapperServerConfig
		err := wrapper.Unmarshal(ctx, "nonexistent", &server)
		if err == nil {
			t.Error("Expected error for missing key")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var nilWrapper *ext.ConfigWrapper
		var server WrapperServerConfig
		err := nilWrapper.Unmarshal(ctx, "server", &server)
		if err == nil {
			t.Error("Expected error for nil wrapper")
		}
	})
}

func TestConfigWrapper_Unmarshal_WithMarshableConfig(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}
	mockCfg := newMockMarshableConfig(data)
	wrapper := ext.NewConfigWrapper(mockCfg)

	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "server", &server)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Verify that the native Unmarshal was called
	if mockCfg.unmarshalCalls != 1 {
		t.Errorf("Expected 1 Unmarshal call, got %d", mockCfg.unmarshalCalls)
	}
}

func TestConfigWrapper_SetValue(t *testing.T) {
	ctx := context.Background()

	t.Run("with mutable config", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		mockCfg := newMockMutableConfig(data)
		wrapper := ext.NewConfigWrapper(mockCfg)

		err := wrapper.SetValue(ctx, "newkey", "newvalue")
		if err != nil {
			t.Fatalf("SetValue failed: %v", err)
		}

		if mockCfg.data["newkey"] != "newvalue" {
			t.Error("Value was not set")
		}
	})

	t.Run("with non-mutable config", func(t *testing.T) {
		// Use read-only mock that doesn't implement MutableConfig
		readOnlyCfg := newMockReadOnlyConfig(map[string]any{})
		wrapper := ext.NewConfigWrapper(readOnlyCfg)

		err := wrapper.SetValue(ctx, "key", "value")
		if err == nil {
			t.Error("Expected error for non-mutable config")
		}
		// Check error message contains expected text
		if err != nil && !strings.Contains(err.Error(), "does not implement MutableConfig") {
			t.Errorf("Expected error about MutableConfig, got %v", err)
		}
	})

	t.Run("with empty key", func(t *testing.T) {
		data := map[string]any{}
		mockCfg := newMockMutableConfig(data)
		wrapper := ext.NewConfigWrapper(mockCfg)

		err := wrapper.SetValue(ctx, "", "value")
		if err == nil {
			t.Error("Expected error for empty key")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var nilWrapper *ext.ConfigWrapper
		err := nilWrapper.SetValue(ctx, "key", "value")
		if err == nil {
			t.Error("Expected error for nil wrapper")
		}
	})
}

func TestConfigWrapper_IsMutable(t *testing.T) {
	t.Run("with mutable config", func(t *testing.T) {
		mockCfg := newMockMutableConfig(map[string]any{})
		wrapper := ext.NewConfigWrapper(mockCfg)

		if !wrapper.IsMutable() {
			t.Error("Expected IsMutable to be true")
		}
	})

	t.Run("with non-mutable config", func(t *testing.T) {
		// Use read-only mock that doesn't implement MutableConfig
		readOnlyCfg := newMockReadOnlyConfig(map[string]any{})
		wrapper := ext.NewConfigWrapper(readOnlyCfg)

		if wrapper.IsMutable() {
			t.Error("Expected IsMutable to be false")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var nilWrapper *ext.ConfigWrapper
		if nilWrapper.IsMutable() {
			t.Error("Expected IsMutable to be false for nil wrapper")
		}
	})
}

func TestConfigWrapper_IsMarshable(t *testing.T) {
	ctx := context.Background()

	t.Run("with marshable config", func(t *testing.T) {
		mockCfg := newMockMarshableConfig(map[string]any{})
		wrapper := ext.NewConfigWrapper(mockCfg)

		if !wrapper.IsMarshable() {
			t.Error("Expected IsMarshable to be true")
		}
	})

	t.Run("with non-marshable config", func(t *testing.T) {
		cfg := config.NewSimpleConfig(ctx)
		wrapper := ext.NewConfigWrapper(cfg)

		if wrapper.IsMarshable() {
			t.Error("Expected IsMarshable to be false")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var nilWrapper *ext.ConfigWrapper
		if nilWrapper.IsMarshable() {
			t.Error("Expected IsMarshable to be false for nil wrapper")
		}
	})
}

func TestConfigWrapper_All(t *testing.T) {
	ctx := context.Background()

	t.Run("with config that supports All", func(t *testing.T) {
		data := map[string]any{"key": "value"}
		mockCfg := newMockMutableConfig(data)
		wrapper := ext.NewConfigWrapper(mockCfg)

		all := wrapper.All(ctx)
		if all == nil {
			t.Fatal("Expected non-nil map")
		}
		if all["key"] != "value" {
			t.Error("Expected key to be 'value'")
		}
	})

	t.Run("with config that doesn't support All", func(t *testing.T) {
		cfg := config.NewSimpleConfig(ctx)
		wrapper := ext.NewConfigWrapper(cfg)

		all := wrapper.All(ctx)
		// SimpleConfig may or may not implement All - just verify no panic
		_ = all
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var nilWrapper *ext.ConfigWrapper
		all := nilWrapper.All(ctx)
		if all == nil {
			t.Error("Expected non-nil map for nil wrapper")
		}
		if len(all) != 0 {
			t.Error("Expected empty map for nil wrapper")
		}
	})
}

func TestConfigWrapper_UnmarshalWithOptions(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": "8080", // String that should be converted
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	t.Run("with weak type conversion", func(t *testing.T) {
		var server WrapperServerConfig
		err := wrapper.Unmarshal(ctx, "server", &server, ext.WithWeaklyTypedInput(true))
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if server.Port != 8080 {
			t.Errorf("Expected port 8080, got %d", server.Port)
		}
	})

	t.Run("with JSON tag", func(t *testing.T) {
		type JSONConfig struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		}
		var cfg JSONConfig
		err := wrapper.Unmarshal(ctx, "server", &cfg, ext.WithJSONTag())
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if cfg.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got '%s'", cfg.Host)
		}
	})
}

// mockAllConfigWrapper implements config.Config with All() for wrapper tests
type mockAllConfigWrapper struct {
	data map[string]any
}

func (m *mockAllConfigWrapper) Name() config.ProviderName {
	return "mockAllConfigWrapper"
}

func (m *mockAllConfigWrapper) GetValue(ctx context.Context, key string) (any, error) {
	if key == "" {
		return m.data, nil
	}
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	return val, nil
}

func (m *mockAllConfigWrapper) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if subMap, ok := val.(map[string]any); ok {
		return &mockAllConfigWrapper{data: subMap}, nil
	}
	return nil, config.ErrConfigInvalidValue.New("not a map", "key", key)
}

func (m *mockAllConfigWrapper) All(ctx context.Context) map[string]any {
	return m.data
}

func TestConfigWrapper_UnmarshalWithAllMethod(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host":    "localhost",
			"port":    8080,
			"timeout": "30s",
		},
	}
	mockCfg := &mockAllConfigWrapper{data: data}
	wrapper := ext.NewConfigWrapper(mockCfg)

	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "server", &server)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if server.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", server.Host)
	}
	if server.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.Port)
	}
}

func TestConfigWrapper_UnmarshalRootConfig(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"host":    "localhost",
		"port":    8080,
		"timeout": "30s",
	}
	mockCfg := &mockAllConfigWrapper{data: data}
	wrapper := ext.NewConfigWrapper(mockCfg)

	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "", &server)
	if err != nil {
		t.Fatalf("Unmarshal root failed: %v", err)
	}

	if server.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", server.Host)
	}
}

func TestConfigWrapper_UnmarshalNonMapValue(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"value": "not a map",
	}
	mockCfg := &mockAllConfigWrapper{data: data}
	wrapper := ext.NewConfigWrapper(mockCfg)

	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "value", &server)
	if err == nil {
		t.Error("Expected error for non-map value")
	}
}

func TestConfigWrapper_UnmarshalWithDecodeHook(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"config": map[string]any{
			"timeout": "30s",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type TimeoutConfig struct {
		Timeout time.Duration `config:"timeout"`
	}

	var result TimeoutConfig
	err := wrapper.Unmarshal(ctx, "config", &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", result.Timeout)
	}
}

// mockRootErrorConfig returns error for root and doesn't implement All()
type mockRootErrorConfig struct {
	data map[string]any
}

func (m *mockRootErrorConfig) Name() config.ProviderName {
	return "mockRootErrorConfig"
}

func (m *mockRootErrorConfig) GetValue(ctx context.Context, key string) (any, error) {
	if key == "" {
		return nil, config.ErrConfigEmptyKey.New("empty key not supported")
	}
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	return val, nil
}

func (m *mockRootErrorConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if subMap, ok := val.(map[string]any); ok {
		return &mockRootErrorConfig{data: subMap}, nil
	}
	return nil, config.ErrConfigInvalidValue.New("not a map", "key", key)
}

func TestConfigWrapper_UnmarshalRootError(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"host": "localhost",
	}
	mockCfg := &mockRootErrorConfig{data: data}
	wrapper := ext.NewConfigWrapper(mockCfg)

	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "", &server)
	if err == nil {
		t.Error("Expected error for root config without All() support")
	}
}

// mockRootNonMapConfig returns non-map value for root
type mockRootNonMapConfig struct{}

func (m *mockRootNonMapConfig) Name() config.ProviderName {
	return "mockRootNonMapConfig"
}

func (m *mockRootNonMapConfig) GetValue(ctx context.Context, key string) (any, error) {
	if key == "" {
		return "not a map", nil // Return non-map value
	}
	return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
}

func (m *mockRootNonMapConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
}

func TestConfigWrapper_UnmarshalRootNonMap(t *testing.T) {
	ctx := context.Background()
	mockCfg := &mockRootNonMapConfig{}
	wrapper := ext.NewConfigWrapper(mockCfg)

	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "", &server)
	if err == nil {
		t.Error("Expected error for non-map root config")
	}
}

// mockGetConfigErrorConfig returns error from GetConfig
type mockGetConfigErrorConfig struct {
	data map[string]any
}

func (m *mockGetConfigErrorConfig) Name() config.ProviderName {
	return "mockGetConfigErrorConfig"
}

func (m *mockGetConfigErrorConfig) GetValue(ctx context.Context, key string) (any, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	return val, nil
}

func (m *mockGetConfigErrorConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	return nil, config.ErrConfigMissingValue.New("always fails", "key", key)
}

func TestConfigWrapper_UnmarshalGetConfigError(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
		},
	}
	mockCfg := &mockGetConfigErrorConfig{data: data}
	wrapper := ext.NewConfigWrapper(mockCfg)

	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "server", &server)
	if err == nil {
		t.Error("Expected error when GetConfig fails")
	}
}

func TestConfigWrapper_Name(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)
	wrapper := ext.NewConfigWrapper(cfg)

	name := wrapper.Name()
	if name != "ConfigWrapper" {
		t.Errorf("Expected name 'ConfigWrapper', got '%s'", name)
	}
}

func TestConfigWrapper_ShouldInstrument(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)
	wrapper := ext.NewConfigWrapper(cfg)

	// ShouldInstrument should always return true
	if !wrapper.ShouldInstrument(ctx, "any.key", "get_value") {
		t.Error("Expected ShouldInstrument to return true")
	}
	if !wrapper.ShouldInstrument(ctx, "", "set_value") {
		t.Error("Expected ShouldInstrument to return true for empty key")
	}
}

func TestConfigWrapper_GenerateTelemetryAttributes(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)
	wrapper := ext.NewConfigWrapper(cfg)

	attrs := []any{"key1", "value1", "key2", "value2"}
	result := wrapper.GenerateTelemetryAttributes(ctx, "get_value", attrs)

	// Should return attrs as-is
	if len(result) != len(attrs) {
		t.Errorf("Expected %d attrs, got %d", len(attrs), len(result))
	}
	for i, v := range attrs {
		if result[i] != v {
			t.Errorf("Expected attr[%d] = %v, got %v", i, v, result[i])
		}
	}
}

func TestConfigWrapper_Keys(t *testing.T) {
	ctx := context.Background()

	t.Run("with config that supports Keys", func(t *testing.T) {
		data := map[string]any{
			"server": map[string]any{
				"port": 8080,
				"host": "localhost",
			},
			"app": map[string]any{
				"name": "test",
			},
		}
		cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
		wrapper := ext.NewConfigWrapper(cfg)

		// Get all keys
		keys := wrapper.Keys("")
		if keys == nil {
			t.Error("Expected non-nil keys")
		}
		if len(keys) < 4 {
			t.Errorf("Expected at least 4 keys, got %d", len(keys))
		}

		// Get keys with prefix
		serverKeys := wrapper.Keys("server")
		if serverKeys == nil {
			t.Error("Expected non-nil server keys")
		}
	})

	t.Run("with config that doesn't support Keys", func(t *testing.T) {
		mockCfg := newMockReadOnlyConfig(map[string]any{"key": "value"})
		wrapper := ext.NewConfigWrapper(mockCfg)

		keys := wrapper.Keys("")
		if keys == nil {
			t.Error("Expected non-nil keys for config without AllKeysProvider")
		}
		if len(keys) != 0 {
			t.Error("Expected empty keys for config without AllKeysProvider")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var wrapper *ext.ConfigWrapper
		keys := wrapper.Keys("")
		if keys == nil {
			t.Error("Expected non-nil keys for nil wrapper")
		}
		if len(keys) != 0 {
			t.Error("Expected empty keys for nil wrapper")
		}
	})
}

func TestConfigWrapper_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("with config that supports Delete", func(t *testing.T) {
		data := map[string]any{
			"server": map[string]any{
				"port":  8080,
				"debug": true,
			},
		}
		cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
		wrapper := ext.NewConfigWrapper(cfg)

		// Delete a key
		err := wrapper.Delete("server.debug")
		if err != nil {
			t.Errorf("Delete failed: %v", err)
		}

		// Verify key is deleted
		_, err = wrapper.GetValue(ctx, "server.debug")
		if err == nil {
			t.Error("Expected error when getting deleted key")
		}
	})

	t.Run("with config that doesn't support Delete", func(t *testing.T) {
		mockCfg := newMockReadOnlyConfig(map[string]any{"key": "value"})
		wrapper := ext.NewConfigWrapper(mockCfg)

		err := wrapper.Delete("key")
		if err == nil {
			t.Error("Expected error for config without Deleter")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var wrapper *ext.ConfigWrapper
		err := wrapper.Delete("key")
		if err == nil {
			t.Error("Expected error for nil wrapper")
		}
	})
}

func TestConfigWrapper_HasAllGetter(t *testing.T) {
	ctx := context.Background()

	t.Run("with config that supports AllGetter", func(t *testing.T) {
		cfg := config.NewSimpleConfig(ctx)
		wrapper := ext.NewConfigWrapper(cfg)

		if !wrapper.HasAllGetter() {
			t.Error("Expected HasAllGetter to return true for SimpleConfig")
		}
	})

	t.Run("with config that doesn't support AllGetter", func(t *testing.T) {
		mockCfg := newMockReadOnlyConfig(map[string]any{})
		wrapper := ext.NewConfigWrapper(mockCfg)

		if wrapper.HasAllGetter() {
			t.Error("Expected HasAllGetter to return false for mock config")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var wrapper *ext.ConfigWrapper
		if wrapper.HasAllGetter() {
			t.Error("Expected HasAllGetter to return false for nil wrapper")
		}
	})
}

func TestConfigWrapper_HasAllKeys(t *testing.T) {
	ctx := context.Background()

	t.Run("with config that supports AllKeysProvider", func(t *testing.T) {
		cfg := config.NewSimpleConfig(ctx)
		wrapper := ext.NewConfigWrapper(cfg)

		if !wrapper.HasAllKeys() {
			t.Error("Expected HasAllKeys to return true for SimpleConfig")
		}
	})

	t.Run("with config that doesn't support AllKeysProvider", func(t *testing.T) {
		mockCfg := newMockReadOnlyConfig(map[string]any{})
		wrapper := ext.NewConfigWrapper(mockCfg)

		if wrapper.HasAllKeys() {
			t.Error("Expected HasAllKeys to return false for mock config")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var wrapper *ext.ConfigWrapper
		if wrapper.HasAllKeys() {
			t.Error("Expected HasAllKeys to return false for nil wrapper")
		}
	})
}

func TestConfigWrapper_HasDeleter(t *testing.T) {
	ctx := context.Background()

	t.Run("with config that supports Deleter", func(t *testing.T) {
		cfg := config.NewSimpleConfig(ctx)
		wrapper := ext.NewConfigWrapper(cfg)

		if !wrapper.HasDeleter() {
			t.Error("Expected HasDeleter to return true for SimpleConfig")
		}
	})

	t.Run("with config that doesn't support Deleter", func(t *testing.T) {
		mockCfg := newMockReadOnlyConfig(map[string]any{})
		wrapper := ext.NewConfigWrapper(mockCfg)

		if wrapper.HasDeleter() {
			t.Error("Expected HasDeleter to return false for mock config")
		}
	})

	t.Run("nil wrapper", func(t *testing.T) {
		var wrapper *ext.ConfigWrapper
		if wrapper.HasDeleter() {
			t.Error("Expected HasDeleter to return false for nil wrapper")
		}
	})
}

func TestConfigWrapper_InterfaceCompileCheck(t *testing.T) {
	ctx := context.Background()
	cfg := config.NewSimpleConfig(ctx)
	wrapper := ext.NewConfigWrapper(cfg)

	// Runtime interface checks
	var _ config.Config = wrapper
	var _ config.MutableConfig = wrapper
	var _ config.MarshableConfig = wrapper
	var _ config.AllGetter = wrapper
	var _ config.AllKeysProvider = wrapper
	var _ config.Deleter = wrapper
}

// mockConfigNoAllGetter implements config.Config but NOT AllGetter
// and returns error for empty key GetValue
type mockConfigNoAllGetter struct {
	data map[string]any
}

func (m *mockConfigNoAllGetter) Name() config.ProviderName {
	return "mockConfigNoAllGetter"
}

func (m *mockConfigNoAllGetter) GetValue(ctx context.Context, key string) (any, error) {
	if key == "" {
		return nil, config.ErrConfigMissingValue.New("empty key not supported")
	}
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	return val, nil
}

func (m *mockConfigNoAllGetter) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if subMap, ok := val.(map[string]any); ok {
		return &mockConfigNoAllGetter{data: subMap}, nil
	}
	return nil, config.ErrConfigInvalidValue.New("not a map", "key", key)
}

func TestConfigWrapper_Unmarshal_RootWithoutAllGetter(t *testing.T) {
	ctx := context.Background()

	// Mock config that doesn't support AllGetter and returns error for empty key
	mockCfg := &mockConfigNoAllGetter{
		data: map[string]any{
			"server": map[string]any{
				"host": "localhost",
				"port": 8080,
			},
		},
	}
	wrapper := ext.NewConfigWrapper(mockCfg)

	// This should fail because root config can't be retrieved without AllGetter
	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "", &server)
	if err == nil {
		t.Error("Expected error when unmarshalling root without AllGetter support")
	}
}

// mockConfigWithSubConfigNoAllGetter returns a sub-config that doesn't implement AllGetter
type mockConfigWithSubConfigNoAllGetter struct {
	data map[string]any
}

func (m *mockConfigWithSubConfigNoAllGetter) Name() config.ProviderName {
	return "mockConfigWithSubConfigNoAllGetter"
}

func (m *mockConfigWithSubConfigNoAllGetter) GetValue(ctx context.Context, key string) (any, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	return val, nil
}

func (m *mockConfigWithSubConfigNoAllGetter) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if subMap, ok := val.(map[string]any); ok {
		// Return a sub-config that doesn't implement AllGetter
		return &mockConfigNoAllGetter{data: subMap}, nil
	}
	return nil, config.ErrConfigInvalidValue.New("not a map", "key", key)
}

func TestConfigWrapper_Unmarshal_NestedWithoutAllGetter(t *testing.T) {
	ctx := context.Background()

	// Mock config where GetConfig returns a config without AllGetter
	// but GetValue returns the map directly
	mockCfg := &mockConfigWithSubConfigNoAllGetter{
		data: map[string]any{
			"server": map[string]any{
				"host": "localhost",
				"port": 8080,
			},
		},
	}
	wrapper := ext.NewConfigWrapper(mockCfg)

	// This should succeed because GetValue returns the map
	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "server", &server)
	if err != nil {
		t.Errorf("Unmarshal failed: %v", err)
	}
	if server.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", server.Host)
	}
}

// mockConfigNestedGetValueError returns error from GetValue for nested key
type mockConfigNestedGetValueError struct {
	data map[string]any
}

func (m *mockConfigNestedGetValueError) Name() config.ProviderName {
	return "mockConfigNestedGetValueError"
}

func (m *mockConfigNestedGetValueError) GetValue(ctx context.Context, key string) (any, error) {
	// Always return error for any key
	return nil, config.ErrConfigMissingValue.New("always fails", "key", key)
}

func (m *mockConfigNestedGetValueError) GetConfig(ctx context.Context, key string) (config.Config, error) {
	// Return a config that doesn't implement AllGetter
	return &mockConfigNoAllGetter{data: map[string]any{}}, nil
}

func TestConfigWrapper_Unmarshal_NestedGetValueError(t *testing.T) {
	ctx := context.Background()

	mockCfg := &mockConfigNestedGetValueError{
		data: map[string]any{},
	}
	wrapper := ext.NewConfigWrapper(mockCfg)

	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "server", &server)
	if err == nil {
		t.Error("Expected error when GetValue fails for nested key")
	}
}

// mockConfigNestedNonMapValue returns non-map value from GetValue
type mockConfigNestedNonMapValue struct{}

func (m *mockConfigNestedNonMapValue) Name() config.ProviderName {
	return "mockConfigNestedNonMapValue"
}

func (m *mockConfigNestedNonMapValue) GetValue(ctx context.Context, key string) (any, error) {
	// Return a non-map value
	return "not a map", nil
}

func (m *mockConfigNestedNonMapValue) GetConfig(ctx context.Context, key string) (config.Config, error) {
	// Return a config that doesn't implement AllGetter
	return &mockConfigNoAllGetter{data: map[string]any{}}, nil
}

func TestConfigWrapper_Unmarshal_NestedNonMapValue(t *testing.T) {
	ctx := context.Background()

	mockCfg := &mockConfigNestedNonMapValue{}
	wrapper := ext.NewConfigWrapper(mockCfg)

	var server WrapperServerConfig
	err := wrapper.Unmarshal(ctx, "server", &server)
	if err == nil {
		t.Error("Expected error when GetValue returns non-map value")
	}
}

func TestConfigWrapper_All_WithConfigWithoutAllGetter(t *testing.T) {
	ctx := context.Background()

	// Mock config that doesn't implement AllGetter
	mockCfg := &mockConfigNoAllGetter{
		data: map[string]any{"key": "value"},
	}
	wrapper := ext.NewConfigWrapper(mockCfg)

	// All() should return empty map when AllGetter is not supported
	all := wrapper.All(ctx)
	if all == nil {
		t.Error("Expected non-nil map")
	}
	if len(all) != 0 {
		t.Error("Expected empty map for config without AllGetter")
	}
}
