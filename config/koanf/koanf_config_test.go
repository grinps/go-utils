package koanf

import (
	"context"
	"testing"

	"github.com/grinps/go-utils/config"
	"github.com/knadh/koanf/parsers/json"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
)

func TestNewKoanfConfig(t *testing.T) {
	ctx := context.Background()

	t.Run("creates empty config", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx)
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}

		kCfg, ok := cfg.(*KoanfConfig)
		if !ok {
			t.Fatal("expected *KoanfConfig type")
		}

		if kCfg.delimiter != config.DefaultKeyDelimiter {
			t.Errorf("expected delimiter %q, got %q", config.DefaultKeyDelimiter, kCfg.delimiter)
		}
	})

	t.Run("creates config with custom delimiter", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx, WithDelimiter("/"))
		kCfg := cfg.(*KoanfConfig)

		if kCfg.delimiter != "/" {
			t.Errorf("expected delimiter %q, got %q", "/", kCfg.delimiter)
		}
	})

	t.Run("creates config with provider", func(t *testing.T) {
		data := map[string]any{
			"server": map[string]any{
				"port": 8080,
				"host": "localhost",
			},
		}

		cfg := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), json.Parser()))
		kCfg := cfg.(*KoanfConfig)

		val, err := kCfg.GetValue(ctx, "server.port")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// koanf may return different numeric types
		switch v := val.(type) {
		case int:
			if v != 8080 {
				t.Errorf("expected port 8080, got %d", v)
			}
		case int64:
			if v != 8080 {
				t.Errorf("expected port 8080, got %d", v)
			}
		case float64:
			if v != 8080 {
				t.Errorf("expected port 8080, got %f", v)
			}
		default:
			t.Errorf("unexpected type for port: %T", val)
		}
	})
}

func TestFromKoanf(t *testing.T) {
	t.Run("wraps existing koanf instance", func(t *testing.T) {
		k := koanf.New(".")
		_ = k.Load(confmap.Provider(map[string]any{"key": "value"}, "."), nil)

		cfg := FromKoanf(k)
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}

		val, err := cfg.GetValue(context.Background(), "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val != "value" {
			t.Errorf("expected value %q, got %q", "value", val)
		}
	})

	t.Run("returns nil for nil koanf", func(t *testing.T) {
		cfg := FromKoanf(nil)
		if cfg != nil {
			t.Error("expected nil config")
		}
	})
}

func TestKoanfConfig_GetValue(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
			"host": "localhost",
		},
		"database": map[string]any{
			"host": "db.example.com",
			"port": 5432,
		},
		"string": "value",
		"number": 42,
	}

	cfg := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil))

	t.Run("retrieves simple value", func(t *testing.T) {
		val, err := cfg.GetValue(ctx, "string")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val != "value" {
			t.Errorf("expected %q, got %q", "value", val)
		}
	})

	t.Run("retrieves nested value", func(t *testing.T) {
		val, err := cfg.GetValue(ctx, "server.host")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val != "localhost" {
			t.Errorf("expected %q, got %q", "localhost", val)
		}
	})

	t.Run("returns error for empty key", func(t *testing.T) {
		_, err := cfg.GetValue(ctx, "")
		if err == nil {
			t.Fatal("expected error for empty key")
		}
	})

	t.Run("returns error for missing key", func(t *testing.T) {
		_, err := cfg.GetValue(ctx, "nonexistent")
		if err == nil {
			t.Fatal("expected error for missing key")
		}
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		_, err := nilCfg.GetValue(ctx, "key")
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})
}

func TestKoanfConfig_GetConfig(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
			"host": "localhost",
		},
	}

	cfg := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil))

	t.Run("retrieves sub-config", func(t *testing.T) {
		subCfg, err := cfg.GetConfig(ctx, "server")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, err := subCfg.GetValue(ctx, "host")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val != "localhost" {
			t.Errorf("expected %q, got %q", "localhost", val)
		}
	})

	t.Run("returns error for empty key", func(t *testing.T) {
		_, err := cfg.GetConfig(ctx, "")
		if err == nil {
			t.Fatal("expected error for empty key")
		}
	})

	t.Run("returns error for missing key", func(t *testing.T) {
		_, err := cfg.GetConfig(ctx, "nonexistent")
		if err == nil {
			t.Fatal("expected error for missing key")
		}
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		_, err := nilCfg.GetConfig(ctx, "key")
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})
}

func TestKoanfConfig_SetValue(t *testing.T) {
	ctx := context.Background()

	t.Run("sets simple value", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		err := cfg.SetValue(ctx, "key", "value")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, err := cfg.GetValue(ctx, "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val != "value" {
			t.Errorf("expected %q, got %q", "value", val)
		}
	})

	t.Run("sets nested value", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		err := cfg.SetValue(ctx, "server.port", 8080)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, err := cfg.GetValue(ctx, "server.port")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val != 8080 {
			t.Errorf("expected %d, got %v", 8080, val)
		}
	})

	t.Run("overwrites existing value", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		_ = cfg.SetValue(ctx, "key", "old")
		err := cfg.SetValue(ctx, "key", "new")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, err := cfg.GetValue(ctx, "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val != "new" {
			t.Errorf("expected %q, got %q", "new", val)
		}
	})

	t.Run("returns error for empty key", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		err := cfg.SetValue(ctx, "", "value")
		if err == nil {
			t.Fatal("expected error for empty key")
		}
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		err := nilCfg.SetValue(ctx, "key", "value")
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})
}

func TestKoanfConfig_Unmarshal(t *testing.T) {
	ctx := context.Background()

	type ServerConfig struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	type DatabaseConfig struct {
		Host     string `koanf:"host"`
		Port     int    `koanf:"port"`
		Username string `koanf:"username"`
	}

	type AppConfig struct {
		Server   ServerConfig   `koanf:"server"`
		Database DatabaseConfig `koanf:"database"`
	}

	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
		"database": map[string]any{
			"host":     "db.example.com",
			"port":     5432,
			"username": "admin",
		},
	}

	cfg := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil))

	t.Run("unmarshals entire config", func(t *testing.T) {
		var app AppConfig
		err := cfg.(*KoanfConfig).Unmarshal(ctx, "", &app)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if app.Server.Host != "localhost" {
			t.Errorf("expected server host %q, got %q", "localhost", app.Server.Host)
		}

		if app.Server.Port != 8080 {
			t.Errorf("expected server port %d, got %d", 8080, app.Server.Port)
		}

		if app.Database.Host != "db.example.com" {
			t.Errorf("expected database host %q, got %q", "db.example.com", app.Database.Host)
		}
	})

	t.Run("unmarshals sub-config", func(t *testing.T) {
		var server ServerConfig
		err := cfg.(*KoanfConfig).Unmarshal(ctx, "server", &server)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if server.Host != "localhost" {
			t.Errorf("expected host %q, got %q", "localhost", server.Host)
		}

		if server.Port != 8080 {
			t.Errorf("expected port %d, got %d", 8080, server.Port)
		}
	})

	t.Run("unmarshals with JSON tag", func(t *testing.T) {
		type JSONConfig struct {
			Host string `json:"host"`
			Port int    `json:"port"`
		}

		var jsonCfg JSONConfig
		err := cfg.(*KoanfConfig).Unmarshal(ctx, "server", &jsonCfg, WithJSONTag())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if jsonCfg.Host != "localhost" {
			t.Errorf("expected host %q, got %q", "localhost", jsonCfg.Host)
		}
	})

	t.Run("returns error for nil target", func(t *testing.T) {
		err := cfg.(*KoanfConfig).Unmarshal(ctx, "server", nil)
		if err == nil {
			t.Fatal("expected error for nil target")
		}
	})

	t.Run("returns error for non-pointer target", func(t *testing.T) {
		var server ServerConfig
		err := cfg.(*KoanfConfig).Unmarshal(ctx, "server", server)
		if err == nil {
			t.Fatal("expected error for non-pointer target")
		}
	})

	t.Run("returns error for non-struct target", func(t *testing.T) {
		var str string
		err := cfg.(*KoanfConfig).Unmarshal(ctx, "server", &str)
		if err == nil {
			t.Fatal("expected error for non-struct target")
		}
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		var server ServerConfig
		err := nilCfg.Unmarshal(ctx, "server", &server)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})
}

func TestKoanfConfig_Load(t *testing.T) {
	ctx := context.Background()

	t.Run("loads from provider", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		data := map[string]any{
			"key": "value",
		}

		err := cfg.Load(ctx, confmap.Provider(data, "."), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, err := cfg.GetValue(ctx, "key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if val != "value" {
			t.Errorf("expected %q, got %q", "value", val)
		}
	})

	t.Run("returns error for nil provider", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		err := cfg.Load(ctx, nil, nil)
		if err == nil {
			t.Fatal("expected error for nil provider")
		}
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		err := nilCfg.Load(ctx, confmap.Provider(map[string]any{}, "."), nil)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})
}

func TestKoanfConfig_Merge(t *testing.T) {
	ctx := context.Background()

	t.Run("merges configs", func(t *testing.T) {
		cfg1 := NewKoanfConfig(ctx, WithProvider(confmap.Provider(map[string]any{
			"key1": "value1",
			"key2": "old",
		}, "."), nil)).(*KoanfConfig)

		cfg2 := NewKoanfConfig(ctx, WithProvider(confmap.Provider(map[string]any{
			"key2": "new",
			"key3": "value3",
		}, "."), nil)).(*KoanfConfig)

		err := cfg1.Merge(ctx, cfg2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Check key1 is preserved
		val, _ := cfg1.GetValue(ctx, "key1")
		if val != "value1" {
			t.Errorf("expected key1 %q, got %q", "value1", val)
		}

		// Check key2 is overwritten
		val, _ = cfg1.GetValue(ctx, "key2")
		if val != "new" {
			t.Errorf("expected key2 %q, got %q", "new", val)
		}

		// Check key3 is added
		val, _ = cfg1.GetValue(ctx, "key3")
		if val != "value3" {
			t.Errorf("expected key3 %q, got %q", "value3", val)
		}
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		cfg2 := NewKoanfConfig(ctx).(*KoanfConfig)

		err := nilCfg.Merge(ctx, cfg2)
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})

	t.Run("returns error for nil other config", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		err := cfg.Merge(ctx, nil)
		if err == nil {
			t.Fatal("expected error for nil other config")
		}
	})
}

func TestKoanfConfig_All(t *testing.T) {
	ctx := context.Background()

	t.Run("returns all config", func(t *testing.T) {
		data := map[string]any{
			"key1": "value1",
			"key2": "value2",
		}

		cfg := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil)).(*KoanfConfig)

		all := cfg.All(ctx)
		if all == nil {
			t.Fatal("expected non-nil map")
		}

		if all["key1"] != "value1" {
			t.Errorf("expected key1 %q, got %q", "value1", all["key1"])
		}

		if all["key2"] != "value2" {
			t.Errorf("expected key2 %q, got %q", "value2", all["key2"])
		}
	})

	t.Run("returns nil for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		all := nilCfg.All(context.Background())
		if all != nil {
			t.Error("expected nil map")
		}
	})
}

func TestKoanfConfig_Exists(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"key": "value",
	}

	cfg := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil)).(*KoanfConfig)

	t.Run("returns true for existing key", func(t *testing.T) {
		if !cfg.Exists("key") {
			t.Error("expected key to exist")
		}
	})

	t.Run("returns false for missing key", func(t *testing.T) {
		if cfg.Exists("nonexistent") {
			t.Error("expected key to not exist")
		}
	})

	t.Run("returns false for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		if nilCfg.Exists("key") {
			t.Error("expected false for nil config")
		}
	})
}

func TestKoanfConfig_Keys(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
			"host": "localhost",
		},
		"database": map[string]any{
			"port": 5432,
		},
	}

	cfg := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil)).(*KoanfConfig)

	t.Run("returns all keys", func(t *testing.T) {
		keys := cfg.Keys("")
		if len(keys) == 0 {
			t.Fatal("expected non-empty keys")
		}

		// Check that some expected keys are present
		hasServer := false
		hasDatabase := false
		for _, key := range keys {
			if key == "server" || key == "server.port" || key == "server.host" {
				hasServer = true
			}
			if key == "database" || key == "database.port" {
				hasDatabase = true
			}
		}

		if !hasServer {
			t.Error("expected server keys")
		}
		if !hasDatabase {
			t.Error("expected database keys")
		}
	})

	t.Run("returns keys with prefix", func(t *testing.T) {
		keys := cfg.Keys("server")
		if len(keys) == 0 {
			t.Fatal("expected non-empty keys")
		}

		for _, key := range keys {
			if key != "server" && !hasPrefix(key, "server.") {
				t.Errorf("unexpected key %q", key)
			}
		}
	})

	t.Run("returns nil for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		keys := nilCfg.Keys("")
		if keys != nil {
			t.Error("expected nil keys")
		}
	})
}

func TestKoanfConfig_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("deletes existing key", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx, WithProvider(confmap.Provider(map[string]any{
			"key": "value",
		}, "."), nil)).(*KoanfConfig)

		err := cfg.Delete("key")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if cfg.Exists("key") {
			t.Error("expected key to be deleted")
		}
	})

	t.Run("returns error for missing key", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		err := cfg.Delete("nonexistent")
		if err == nil {
			t.Fatal("expected error for missing key")
		}
	})

	t.Run("returns error for empty key", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		err := cfg.Delete("")
		if err == nil {
			t.Fatal("expected error for empty key")
		}
	})

	t.Run("returns error for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		err := nilCfg.Delete("key")
		if err == nil {
			t.Fatal("expected error for nil config")
		}
	})
}

func TestKoanfConfig_Koanf(t *testing.T) {
	ctx := context.Background()

	t.Run("returns underlying koanf instance", func(t *testing.T) {
		cfg := NewKoanfConfig(ctx).(*KoanfConfig)

		k := cfg.Koanf()
		if k == nil {
			t.Fatal("expected non-nil koanf instance")
		}
	})

	t.Run("returns nil for nil config", func(t *testing.T) {
		var nilCfg *KoanfConfig
		k := nilCfg.Koanf()
		if k != nil {
			t.Error("expected nil koanf instance")
		}
	})
}

func TestNewMutableKoanfConfig(t *testing.T) {
	ctx := context.Background()

	cfg := NewMutableKoanfConfig(ctx)
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	// Verify it implements MutableConfig
	_, ok := cfg.(config.MutableConfig)
	if !ok {
		t.Fatal("expected MutableConfig interface")
	}
}

func TestNewMarshableKoanfConfig(t *testing.T) {
	ctx := context.Background()

	cfg := NewMarshableKoanfConfig(ctx)
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}

	// Verify it implements MarshableConfig
	_, ok := cfg.(config.MarshableConfig)
	if !ok {
		t.Fatal("expected MarshableConfig interface")
	}
}

// Helper function
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
