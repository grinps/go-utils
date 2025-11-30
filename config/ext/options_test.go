package ext_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/grinps/go-utils/config"
	"github.com/grinps/go-utils/config/ext"
)

// Options tests use ConfigWrapper since options apply to mapstructure-based unmarshalling

func TestWithSquash(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"embedded": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type EmbeddedConfig struct {
		Host string `config:"host"`
		Port int    `config:"port"`
	}
	type ParentConfig struct {
		EmbeddedConfig `config:",squash"`
	}

	var result ParentConfig
	err := wrapper.Unmarshal(ctx, "embedded", &result, ext.WithSquash(true))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", result.Host)
	}
}

func TestWithErrorUnused(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host":  "localhost",
			"port":  8080,
			"extra": "unused", // This key is not in the struct
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type MinimalConfig struct {
		Host string `config:"host"`
	}

	var result MinimalConfig
	err := wrapper.Unmarshal(ctx, "server", &result, ext.WithErrorUnused(true))
	if err == nil {
		t.Error("Expected error for unused keys")
	}

	// Without error on unused should work
	err = wrapper.Unmarshal(ctx, "server", &result, ext.WithErrorUnused(false))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}

func TestWithErrorUnset(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			// "port" is missing
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type RequiredConfig struct {
		Host string `config:"host"`
		Port int    `config:"port"` // Required but not in config
	}

	var result RequiredConfig
	err := wrapper.Unmarshal(ctx, "server", &result, ext.WithErrorUnset(true))
	if err == nil {
		t.Error("Expected error for unset fields")
	}

	// Without error on unset should work
	result = RequiredConfig{}
	err = wrapper.Unmarshal(ctx, "server", &result, ext.WithErrorUnset(false))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
}

func TestWithDecodeHook(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"config": map[string]any{
			"value": "custom_value",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type CustomType struct {
		Processed bool
		Original  string
	}

	type ConfigWithCustom struct {
		Value CustomType `config:"value"`
	}

	// Custom hook to convert string to CustomType
	hook := func(from reflect.Type, to reflect.Type, data any) (any, error) {
		if to == reflect.TypeOf(CustomType{}) && from.Kind() == reflect.String {
			return CustomType{Processed: true, Original: data.(string)}, nil
		}
		return data, nil
	}

	var result ConfigWithCustom
	err := wrapper.Unmarshal(ctx, "config", &result, ext.WithDecodeHook(hook))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !result.Value.Processed {
		t.Error("Expected custom hook to process value")
	}
	if result.Value.Original != "custom_value" {
		t.Errorf("Expected original 'custom_value', got '%s'", result.Value.Original)
	}
}

func TestWithMetadata(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host":  "localhost",
			"extra": "unused",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type SimpleConfig struct {
		Host string `config:"host"`
	}

	var metadata mapstructure.Metadata
	var result SimpleConfig
	err := wrapper.Unmarshal(ctx, "server", &result, ext.WithMetadata(&metadata))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	// Check metadata was populated
	if len(metadata.Keys) == 0 {
		t.Error("Expected metadata to have decoded keys")
	}
	if len(metadata.Unused) == 0 {
		t.Error("Expected metadata to have unused keys")
	}
}

func TestWithMapstructureTag(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type MapstructureConfig struct {
		Host string `mapstructure:"host"`
		Port int    `mapstructure:"port"`
	}

	var result MapstructureConfig
	err := wrapper.Unmarshal(ctx, "server", &result, ext.WithMapstructureTag())
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", result.Host)
	}
}

func TestWithYAMLTag(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host": "localhost",
			"port": 8080,
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type YAMLConfig struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	}

	var result YAMLConfig
	err := wrapper.Unmarshal(ctx, "server", &result, ext.WithYAMLTag())
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Host != "localhost" {
		t.Errorf("Expected host 'localhost', got '%s'", result.Host)
	}
}

func TestWithStrictMode(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"server": map[string]any{
			"host":  "localhost",
			"extra": "unused",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type StrictConfig struct {
		Host string `config:"host"`
		Port int    `config:"port"` // Required but missing
	}

	var result StrictConfig
	err := wrapper.Unmarshal(ctx, "server", &result, ext.WithStrictMode())
	if err == nil {
		t.Error("Expected error in strict mode")
	}
}

func TestDurationConversion(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"config": map[string]any{
			"timeout":  "30s",
			"interval": "5m",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type DurationConfig struct {
		Timeout  time.Duration `config:"timeout"`
		Interval time.Duration `config:"interval"`
	}

	var result DurationConfig
	err := wrapper.Unmarshal(ctx, "config", &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", result.Timeout)
	}
	if result.Interval != 5*time.Minute {
		t.Errorf("Expected interval 5m, got %v", result.Interval)
	}
}

func TestSliceConversion(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"config": map[string]any{
			"tags": "a,b,c",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type SliceConfig struct {
		Tags []string `config:"tags"`
	}

	var result SliceConfig
	err := wrapper.Unmarshal(ctx, "config", &result)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(result.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(result.Tags))
	}
	if result.Tags[0] != "a" || result.Tags[1] != "b" || result.Tags[2] != "c" {
		t.Errorf("Unexpected tags: %v", result.Tags)
	}
}

func TestMultipleDecodeHooks(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"config": map[string]any{
			"timeout": "30s",
			"tags":    "x,y,z",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type MultiConfig struct {
		Timeout time.Duration `config:"timeout"`
		Tags    []string      `config:"tags"`
	}

	// Custom hook that does nothing but exercises the hook path
	noopHook := func(from reflect.Type, to reflect.Type, data any) (any, error) {
		return data, nil
	}

	var result MultiConfig
	err := wrapper.Unmarshal(ctx, "config", &result, ext.WithDecodeHook(noopHook))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", result.Timeout)
	}
	if len(result.Tags) != 3 {
		t.Errorf("Expected 3 tags, got %d", len(result.Tags))
	}
}

func TestWrapperWithDecodeHooks(t *testing.T) {
	ctx := context.Background()
	data := map[string]any{
		"config": map[string]any{
			"duration": "1h30m",
		},
	}
	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
	wrapper := ext.NewConfigWrapper(cfg)

	type DurationConfig struct {
		Duration time.Duration `config:"duration"`
	}

	// Add custom hook
	customHook := func(from reflect.Type, to reflect.Type, data any) (any, error) {
		return data, nil
	}

	var result DurationConfig
	err := wrapper.Unmarshal(ctx, "config", &result, ext.WithDecodeHook(customHook))
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if result.Duration != 90*time.Minute {
		t.Errorf("Expected duration 1h30m, got %v", result.Duration)
	}
}
