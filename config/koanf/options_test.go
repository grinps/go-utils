package koanf

import (
	"context"
	"testing"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"
)

func TestWithTag(t *testing.T) {
	ctx := context.Background()

	type JSONConfig struct {
		Host string `json:"host"`
		Port int    `json:"port"`
	}

	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}

	cfg, err := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	kcfg := cfg.(*KoanfConfig)

	var jsonCfg JSONConfig
	err = kcfg.Unmarshal(ctx, "server", &jsonCfg, WithJSONTag())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if jsonCfg.Host != "localhost" {
		t.Errorf("expected host %q, got %q", "localhost", jsonCfg.Host)
	}

	if jsonCfg.Port != 8080 {
		t.Errorf("expected port %d, got %d", 8080, jsonCfg.Port)
	}
}

func TestWithYAMLTag(t *testing.T) {
	ctx := context.Background()

	type YAMLConfig struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}

	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}

	cfg, err := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	kcfg := cfg.(*KoanfConfig)

	var yamlCfg YAMLConfig
	err = kcfg.Unmarshal(ctx, "server", &yamlCfg, WithYAMLTag())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if yamlCfg.Host != "localhost" {
		t.Errorf("expected host %q, got %q", "localhost", yamlCfg.Host)
	}
}

func TestWithMapstructureTag(t *testing.T) {
	ctx := context.Background()

	type MapstructureConfig struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}

	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}

	cfg, err := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	kcfg := cfg.(*KoanfConfig)

	var msCfg MapstructureConfig
	err = kcfg.Unmarshal(ctx, "server", &msCfg, WithMapstructureTag())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if msCfg.Host != "localhost" {
		t.Errorf("expected host %q, got %q", "localhost", msCfg.Host)
	}
}

func TestWithKoanfTag(t *testing.T) {
	ctx := context.Background()

	type KoanfTagConfig struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}

	cfg, err := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	kcfg := cfg.(*KoanfConfig)

	var kCfg KoanfTagConfig
	err = kcfg.Unmarshal(ctx, "server", &kCfg, WithKoanfTag())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if kCfg.Host != "localhost" {
		t.Errorf("expected host %q, got %q", "localhost", kCfg.Host)
	}
}

func TestWithFlatPaths(t *testing.T) {
	ctx := context.Background()

	type FlatConfig struct {
		ServerPort int    `koanf:"server.port"`
		ServerHost string `koanf:"server.host"`
	}

	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}

	cfg, err := NewKoanfConfig(ctx, WithProvider(confmap.Provider(data, "."), nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	kcfg := cfg.(*KoanfConfig)

	var flatCfg FlatConfig
	err = kcfg.Unmarshal(ctx, "", &flatCfg, WithFlatPaths(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if flatCfg.ServerHost != "localhost" {
		t.Errorf("expected host %q, got %q", "localhost", flatCfg.ServerHost)
	}

	if flatCfg.ServerPort != 8080 {
		t.Errorf("expected port %d, got %d", 8080, flatCfg.ServerPort)
	}
}

func TestWithKoanfInstance(t *testing.T) {
	ctx := context.Background()

	// Create a koanf instance directly
	k := koanf.New(".")
	_ = k.Load(confmap.Provider(map[string]any{
		"key": "value",
	}, "."), nil)

	// Wrap it with KoanfConfig using WithKoanfInstance
	cfg, err := NewKoanfConfig(ctx, WithKoanfInstance(k))
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
}

func TestWithDelimiter(t *testing.T) {
	ctx := context.Background()

	cfg, err := NewKoanfConfig(ctx, WithDelimiter("/"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	kcfg := cfg.(*KoanfConfig)

	if kcfg.delimiter != "/" {
		t.Errorf("expected delimiter %q, got %q", "/", kcfg.delimiter)
	}

	// Test that the delimiter works
	_ = kcfg.SetValue(ctx, "server/port", 8080)

	val, err := kcfg.GetValue(ctx, "server/port")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if val != 8080 {
		t.Errorf("expected port %d, got %v", 8080, val)
	}
}

func TestBuildUnmarshalConf(t *testing.T) {
	t.Run("builds default config", func(t *testing.T) {
		conf := buildUnmarshalConf()
		if conf.Tag != "koanf" {
			t.Errorf("expected tag %q, got %q", "koanf", conf.Tag)
		}
	})

	t.Run("applies UnmarshalOption", func(t *testing.T) {
		conf := buildUnmarshalConf(WithTag("json"))
		if conf.Tag != "json" {
			t.Errorf("expected tag %q, got %q", "json", conf.Tag)
		}
	})

	t.Run("applies koanf.UnmarshalConf", func(t *testing.T) {
		customConf := koanf.UnmarshalConf{
			Tag:       "custom",
			FlatPaths: true,
		}
		conf := buildUnmarshalConf(customConf)
		if conf.Tag != "custom" {
			t.Errorf("expected tag %q, got %q", "custom", conf.Tag)
		}
		if !conf.FlatPaths {
			t.Error("expected FlatPaths to be true")
		}
	})

	t.Run("applies multiple options", func(t *testing.T) {
		conf := buildUnmarshalConf(WithTag("yaml"), WithFlatPaths(true))
		if conf.Tag != "yaml" {
			t.Errorf("expected tag %q, got %q", "yaml", conf.Tag)
		}
		if !conf.FlatPaths {
			t.Error("expected FlatPaths to be true")
		}
	})
}
