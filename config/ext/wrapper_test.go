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
		if all != nil {
			t.Error("Expected nil for nil wrapper")
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
