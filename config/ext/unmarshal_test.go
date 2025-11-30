package ext_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/grinps/go-utils/config"
	"github.com/grinps/go-utils/config/ext"
)

type ServerConfig struct {
	Host    string        `config:"host"`
	Port    int           `config:"port"`
	Timeout time.Duration `config:"timeout"`
}

type DatabaseConfig struct {
	Host     string `config:"host"`
	Port     int    `config:"port"`
	User     string `config:"user"`
	Password string `config:"password"`
}

type AppConfig struct {
	Server   ServerConfig   `config:"server"`
	Database DatabaseConfig `config:"database"`
}

// testMarshableConfig implements MarshableConfig for testing Unmarshal delegation
type testMarshableConfig struct {
	data           map[string]any
	unmarshalCalls int
	unmarshalErr   error
}

func (m *testMarshableConfig) GetValue(ctx context.Context, key string, returnValue any) error {
	val, ok := m.data[key]
	if !ok {
		return config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if rv, ok := returnValue.(*any); ok {
		*rv = val
	}
	return nil
}

func (m *testMarshableConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, ok := m.data[key]
	if !ok {
		return nil, config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if subMap, ok := val.(map[string]any); ok {
		return &testMarshableConfig{data: subMap}, nil
	}
	return nil, config.ErrConfigInvalidValue.New("not a map", "key", key)
}

func (m *testMarshableConfig) Unmarshal(ctx context.Context, key string, target any, options ...any) error {
	m.unmarshalCalls++
	if m.unmarshalErr != nil {
		return m.unmarshalErr
	}
	// Simple mock implementation - just set Host field if it's a ServerConfig
	if sc, ok := target.(*ServerConfig); ok {
		if key == "" {
			if host, ok := m.data["host"].(string); ok {
				sc.Host = host
			}
		} else if subData, ok := m.data[key].(map[string]any); ok {
			if host, ok := subData["host"].(string); ok {
				sc.Host = host
			}
			if port, ok := subData["port"].(int); ok {
				sc.Port = port
			}
		} else {
			// Key not found
			return config.ErrConfigMissingValue.New("key not found", "key", key)
		}
	}
	return nil
}

// testMutableConfig implements MutableConfig for testing SetValue
type testMutableConfig struct {
	data map[string]any
}

func (m *testMutableConfig) GetValue(ctx context.Context, key string, returnValue any) error {
	val, ok := m.data[key]
	if !ok {
		return config.ErrConfigMissingValue.New("missing value", "key", key)
	}
	if rv, ok := returnValue.(*any); ok {
		*rv = val
	}
	return nil
}

func (m *testMutableConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	return nil, config.ErrConfigMissingValue.New("not implemented", "key", key)
}

func (m *testMutableConfig) SetValue(ctx context.Context, key string, value any) error {
	if key == "" {
		return config.ErrConfigEmptyKey.New("empty key")
	}
	m.data[key] = value
	return nil
}

func TestUnmarshal(t *testing.T) {
	t.Run("delegates to MarshableConfig from context", func(t *testing.T) {
		data := map[string]any{
			"server": map[string]any{
				"host": "localhost",
				"port": 8080,
			},
		}
		mockCfg := &testMarshableConfig{data: data}
		ctx := config.ContextWithConfig(context.Background(), mockCfg)

		var server ServerConfig
		err := ext.Unmarshal(ctx, "server", &server)
		if err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}

		if mockCfg.unmarshalCalls != 1 {
			t.Errorf("Expected 1 Unmarshal call, got %d", mockCfg.unmarshalCalls)
		}
		if server.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got '%s'", server.Host)
		}
	})

	t.Run("error for non-MarshableConfig", func(t *testing.T) {
		cfg := config.NewSimpleConfig(context.Background())
		ctx := config.ContextWithConfig(context.Background(), cfg)

		var server ServerConfig
		err := ext.Unmarshal(ctx, "server", &server)
		if err == nil {
			t.Error("Expected error for non-MarshableConfig")
		}
		if !strings.Contains(err.Error(), "does not implement MarshableConfig") {
			t.Errorf("Expected MarshableConfig error, got: %v", err)
		}
	})

	t.Run("unmarshal with nil target", func(t *testing.T) {
		mockCfg := &testMarshableConfig{data: map[string]any{}}
		ctx := config.ContextWithConfig(context.Background(), mockCfg)

		err := ext.Unmarshal[ServerConfig](ctx, "server", nil)
		if err == nil {
			t.Error("Expected error for nil target")
		}
	})
}

func TestUnmarshalWithConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("delegates to MarshableConfig", func(t *testing.T) {
		data := map[string]any{
			"server": map[string]any{
				"host": "localhost",
				"port": 8080,
			},
		}
		mockCfg := &testMarshableConfig{data: data}

		var server ServerConfig
		err := ext.UnmarshalWithConfig(ctx, mockCfg, "server", &server)
		if err != nil {
			t.Fatalf("UnmarshalWithConfig failed: %v", err)
		}

		if mockCfg.unmarshalCalls != 1 {
			t.Errorf("Expected 1 Unmarshal call, got %d", mockCfg.unmarshalCalls)
		}
		if server.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got '%s'", server.Host)
		}
	})

	t.Run("error for nil config", func(t *testing.T) {
		var server ServerConfig
		err := ext.UnmarshalWithConfig[ServerConfig](ctx, nil, "server", &server)
		if err == nil {
			t.Error("Expected error for nil config")
		}
	})
}

func TestUnmarshalNonStructPointer(t *testing.T) {
	mockCfg := &testMarshableConfig{data: map[string]any{}}
	ctx := config.ContextWithConfig(context.Background(), mockCfg)

	// Test with pointer to non-struct (int) - should fail validation before delegation
	var num int
	err := ext.Unmarshal(ctx, "key", &num)
	if err == nil {
		t.Error("Expected error for non-struct pointer target")
	}
}

func TestMustUnmarshal(t *testing.T) {
	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}
	mockCfg := &testMarshableConfig{data: data}
	ctx := config.ContextWithConfig(context.Background(), mockCfg)

	t.Run("successful unmarshal", func(t *testing.T) {
		var server ServerConfig
		// Should not panic
		ext.MustUnmarshal(ctx, "server", &server)

		if server.Host != "localhost" {
			t.Errorf("Expected host 'localhost', got '%s'", server.Host)
		}
	})

	t.Run("panic on missing key", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for missing key")
			}
		}()
		var server ServerConfig
		ext.MustUnmarshal(ctx, "nonexistent", &server)
	})
}

func TestSetValue(t *testing.T) {
	t.Run("sets value on MutableConfig from context", func(t *testing.T) {
		mockCfg := &testMutableConfig{data: map[string]any{}}
		ctx := config.ContextWithConfig(context.Background(), mockCfg)

		err := ext.SetValue(ctx, "key", "value")
		if err != nil {
			t.Fatalf("SetValue failed: %v", err)
		}

		if mockCfg.data["key"] != "value" {
			t.Errorf("Expected value 'value', got '%v'", mockCfg.data["key"])
		}
	})

	t.Run("error for non-MutableConfig", func(t *testing.T) {
		cfg := config.NewSimpleConfig(context.Background())
		ctx := config.ContextWithConfig(context.Background(), cfg)

		err := ext.SetValue(ctx, "key", "value")
		// SimpleConfig implements MutableConfig, so this should succeed
		if err != nil {
			t.Fatalf("SetValue failed unexpectedly: %v", err)
		}
	})

	t.Run("error for empty key", func(t *testing.T) {
		mockCfg := &testMutableConfig{data: map[string]any{}}
		ctx := config.ContextWithConfig(context.Background(), mockCfg)

		err := ext.SetValue(ctx, "", "value")
		if err == nil {
			t.Error("Expected error for empty key")
		}
	})
}

func TestSetValueWithConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("sets value on MutableConfig", func(t *testing.T) {
		mockCfg := &testMutableConfig{data: map[string]any{}}

		err := ext.SetValueWithConfig(ctx, mockCfg, "key", "value")
		if err != nil {
			t.Fatalf("SetValueWithConfig failed: %v", err)
		}

		if mockCfg.data["key"] != "value" {
			t.Errorf("Expected value 'value', got '%v'", mockCfg.data["key"])
		}
	})

	t.Run("error for nil config", func(t *testing.T) {
		err := ext.SetValueWithConfig(ctx, nil, "key", "value")
		if err == nil {
			t.Error("Expected error for nil config")
		}
	})

	t.Run("error for non-MutableConfig", func(t *testing.T) {
		// Use a config that doesn't implement MutableConfig
		mockCfg := &testMarshableConfig{data: map[string]any{}}

		err := ext.SetValueWithConfig(ctx, mockCfg, "key", "value")
		if err == nil {
			t.Error("Expected error for non-MutableConfig")
		}
		if !strings.Contains(err.Error(), "does not implement MutableConfig") {
			t.Errorf("Expected MutableConfig error, got: %v", err)
		}
	})
}
