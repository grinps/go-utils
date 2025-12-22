package otel

import (
	"context"
	"testing"

	"github.com/grinps/go-utils/config"
	"github.com/grinps/go-utils/config/ext"
	"github.com/grinps/go-utils/errext"
)

func TestLoadConfiguration(t *testing.T) {
	ctx := context.Background()

	t.Run("with valid otelconf structure", func(t *testing.T) {
		simpleCfg := ext.NewConfigWrapper(config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
			"opentelemetry": map[string]any{
				"file_format": "0.3",
				"disabled":    false,
			},
		})))

		cfg, err := LoadConfiguration(ctx, simpleCfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cfg == nil {
			t.Fatal("expected config to be non-nil")
		}
		if cfg.FileFormat != "0.3" {
			t.Errorf("expected file_format '0.3', got %s", cfg.FileFormat)
		}
	})

	t.Run("empty config uses defaults", func(t *testing.T) {
		simpleCfg := config.NewSimpleConfig(ctx)

		cfg, err := LoadConfiguration(ctx, simpleCfg)
		if cfg != nil {
			t.Fatalf("expected config to be nil, error: %v", err)
		}
		if err == nil {
			t.Fatalf("expected error, got nil")
		}
		if !errext.Is(err, ErrCodeConfigLoadFailed) {
			t.Fatalf("expected ErrCodeConfigLoadFailed, got %v", err)
		}
	})

	t.Run("non-marshable config uses defaults", func(t *testing.T) {
		// Create a mock config that doesn't implement MarshableConfig
		cfg, err := LoadConfiguration(ctx, &nonMarshableConfig{})
		if !errext.Is(err, ErrCodeConfigLoadFailed) {
			t.Fatalf("expected ErrCodeConfigLoadFailed, got %v", err)
		}
		if cfg != nil {
			t.Fatalf("expected config to be nil, error: %v", err)
		}
	})
}

// nonMarshableConfig is a mock config that doesn't implement MarshableConfig
type nonMarshableConfig struct{}

func (c *nonMarshableConfig) GetValue(ctx context.Context, key string) (any, error) {
	return nil, nil
}

func (c *nonMarshableConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	return nil, nil
}

func TestDefaultConfiguration(t *testing.T) {
	cfg := DefaultConfiguration()

	if cfg == nil {
		t.Fatal("expected config to be non-nil")
	}
	if cfg.FileFormat != "0.3" {
		t.Errorf("expected default file format '0.3', got %s", cfg.FileFormat)
	}
}

func TestConfigKey(t *testing.T) {
	if ConfigKey != "opentelemetry" {
		t.Errorf("expected ConfigKey 'opentelemetry', got %s", ConfigKey)
	}
}

func TestStatusCodeConstants(t *testing.T) {
	if StatusUnset != 0 {
		t.Errorf("expected StatusUnset 0, got %d", StatusUnset)
	}
	if StatusOK != 1 {
		t.Errorf("expected StatusOK 1, got %d", StatusOK)
	}
	if StatusError != 2 {
		t.Errorf("expected StatusError 2, got %d", StatusError)
	}
}

func TestLoadConfigurationWithFileFormat(t *testing.T) {
	ctx := context.Background()

	simpleCfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"opentelemetry": map[string]any{
			"file_format": "0.3",
			"resource": map[string]any{
				"attributes_list": "service.name=test-service",
			},
		},
	}))

	cfg, err := LoadConfiguration(ctx, simpleCfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config to be non-nil")
	}
	if cfg.FileFormat != "0.3" {
		t.Errorf("expected file_format '0.3', got %s", cfg.FileFormat)
	}
}

func TestLoadConfigurationSetsDefaultFileFormat(t *testing.T) {
	ctx := context.Background()

	// Config without file_format should get default "0.3"
	simpleCfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
		"opentelemetry": map[string]any{
			"disabled": false,
		},
	}))

	cfg, err := LoadConfiguration(ctx, simpleCfg)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg == nil {
		t.Fatal("expected config to be non-nil")
	}
	if cfg.FileFormat != "0.3" {
		t.Errorf("expected default file_format '0.3', got %s", cfg.FileFormat)
	}
}

func TestLoadConfigurationAllBranches(t *testing.T) {
	ctx := context.Background()

	t.Run("successful parse sets default file_format", func(t *testing.T) {
		// Config with minimal fields - parse should succeed and file_format gets default
		simpleCfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
			"opentelemetry": map[string]any{
				"disabled": false,
			},
		}))

		cfg, err := LoadConfiguration(ctx, simpleCfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.FileFormat != "0.3" {
			t.Errorf("expected default file_format '0.3', got %s", cfg.FileFormat)
		}
	})

	t.Run("config without opentelemetry key", func(t *testing.T) {
		simpleCfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(map[string]any{
			"other_key": "value",
		}))

		cfg, err := LoadConfiguration(ctx, simpleCfg)
		if !errext.Is(err, ErrCodeConfigLoadFailed) {
			t.Fatalf("expected ErrCodeConfigLoadFailed, got %v", err)
		}
		if cfg != nil {
			t.Fatal("expected nil config")
		}
	})
}
