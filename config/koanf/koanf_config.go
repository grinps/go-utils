package koanf

import (
	"context"
	"reflect"
	"strings"

	"github.com/grinps/go-utils/config"
	"github.com/knadh/koanf/v2"
)

// KoanfConfig wraps knadh/koanf to implement config.Config, config.MutableConfig,
// and config.MarshableConfig interfaces.
//
// It provides a powerful configuration management solution with support for:
//   - Multiple configuration sources (files, env vars, command-line flags, etc.)
//   - Nested configuration access via dot-notation
//   - Type-safe unmarshalling to structs
//   - Mutable configuration with SetValue
//   - Provider-based configuration loading
//
// Example:
//
//	cfg := koanf.NewKoanfConfig(ctx)
//	err := cfg.Load(ctx, file.Provider("config.json"), json.Parser())
//	var server ServerConfig
//	err = cfg.Unmarshal(ctx, "server", &server)
type KoanfConfig struct {
	k         *koanf.Koanf
	delimiter string
}

// Ensure KoanfConfig implements all required interfaces
var (
	_ config.Config          = (*KoanfConfig)(nil)
	_ config.MutableConfig   = (*KoanfConfig)(nil)
	_ config.MarshableConfig = (*KoanfConfig)(nil)
)

// Name returns the provider name for KoanfConfig.
// Implements config.Config interface.
func (k *KoanfConfig) Name() config.ProviderName {
	return "KoanfConfig"
}

// ShouldInstrument always returns true for KoanfConfig.
// Implements config.TelemetryAware interface.
func (k *KoanfConfig) ShouldInstrument(ctx context.Context, key string, op string) bool {
	return true
}

// GenerateTelemetryAttributes returns the attributes as-is.
// Implements config.TelemetryAware interface.
func (k *KoanfConfig) GenerateTelemetryAttributes(ctx context.Context, op string, attrs []any) []any {
	return attrs
}

// GetValue retrieves a configuration value by key using dot-notation.
// Returns the value and an error if the key is empty or value is not found.
//
// Example:
//
//	port, err := cfg.GetValue(ctx, "server.port")
//	if err != nil {
//	    // handle error
//	}
//	portInt := port.(int)
func (k *KoanfConfig) GetValue(ctx context.Context, key string) (any, error) {
	if k == nil || k.k == nil {
		return nil, ErrKoanfNilConfig.New("nil koanf config", "key", key)
	}

	if key == "" {
		return nil, ErrKoanfEmptyKey.New("empty key", "key", key)
	}

	if !k.k.Exists(key) {
		return nil, ErrKoanfMissingValue.New("key not found", "key", key)
	}

	val := k.k.Get(key)
	if val == nil {
		return nil, ErrKoanfMissingValue.New("value is nil", "key", key)
	}

	return val, nil
}

// GetConfig retrieves a nested configuration as a new Config instance.
// The value at the key must be a map structure.
//
// Example:
//
//	serverCfg, err := cfg.GetConfig(ctx, "server")
//	port, err := serverCfg.GetValue(ctx, "port")
func (k *KoanfConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	if k == nil || k.k == nil {
		return nil, ErrKoanfNilConfig.New("nil koanf config", "key", key)
	}

	if key == "" {
		return nil, ErrKoanfEmptyKey.New("empty key", "key", key)
	}

	if !k.k.Exists(key) {
		return nil, ErrKoanfMissingValue.New("key not found", "key", key)
	}

	// Get the sub-configuration
	subKoanf := k.k.Cut(key)
	if subKoanf == nil {
		return nil, ErrKoanfMissingValue.New("sub-config is nil", "key", key)
	}

	return &KoanfConfig{
		k:         subKoanf,
		delimiter: k.delimiter,
	}, nil
}

// SetValue sets a configuration value at the specified key.
// Keys can use dot-notation for nested access (e.g., "server.port").
// Intermediate maps are created automatically if they don't exist.
//
// Example:
//
//	err := cfg.SetValue(ctx, "server.port", 8080)
//	err := cfg.SetValue(ctx, "database.credentials.password", "secret")
func (k *KoanfConfig) SetValue(ctx context.Context, key string, newValue any) error {
	if k == nil || k.k == nil {
		return ErrKoanfNilConfig.New("nil koanf config", "key", key)
	}

	if key == "" {
		return ErrKoanfEmptyKey.New("empty key", "key", key)
	}

	// Use koanf's Set method which handles nested keys automatically
	err := k.k.Set(key, newValue)
	if err != nil {
		return ErrKoanfSetValueFailed.New("failed to set value", "key", key, "error", err)
	}

	return nil
}

// Unmarshal unmarshals the configuration at the given key into the target struct.
// If key is empty, the entire configuration is unmarshalled.
// The target must be a pointer to a struct.
//
// Options can be provided to customize unmarshalling behavior.
// Supported option types:
//   - UnmarshalOption: Custom unmarshalling options
//   - koanf.UnmarshalConf: Direct koanf unmarshal configuration
//
// Example:
//
//	type ServerConfig struct {
//	    Host string `koanf:"host"`
//	    Port int    `koanf:"port"`
//	}
//	var server ServerConfig
//	err := cfg.Unmarshal(ctx, "server", &server)
//
//	// With options
//	err := cfg.Unmarshal(ctx, "server", &server, WithTag("json"))
func (k *KoanfConfig) Unmarshal(ctx context.Context, key string, target any, options ...any) error {
	if k == nil || k.k == nil {
		return ErrKoanfNilConfig.New("nil koanf config", "key", key)
	}

	if target == nil {
		return ErrKoanfInvalidTarget.New("target is nil", "key", key)
	}

	// Validate target is a pointer to struct
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return ErrKoanfInvalidTarget.New("target must be a pointer", "key", key, "type", targetType.Kind().String())
	}

	elemType := targetType.Elem()
	if elemType.Kind() != reflect.Struct {
		return ErrKoanfInvalidTarget.New("target must be a pointer to struct", "key", key, "type", elemType.Kind().String())
	}

	// Build unmarshal configuration from options
	unmarshalConf := buildUnmarshalConf(options...)

	var err error
	if key == "" {
		// Unmarshal entire config
		err = k.k.UnmarshalWithConf("", target, unmarshalConf)
	} else {
		// Unmarshal specific key
		err = k.k.UnmarshalWithConf(key, target, unmarshalConf)
	}

	if err != nil {
		return ErrKoanfUnmarshalFailed.New("unmarshal failed", "key", key, "error", err)
	}

	return nil
}

// Load loads configuration from a provider with a parser.
// This is a convenience method that wraps koanf's Load functionality.
//
// Example:
//
//	import (
//	    "github.com/knadh/koanf/parsers/json"
//	    "github.com/knadh/koanf/providers/file"
//	)
//
//	err := cfg.Load(ctx, file.Provider("config.json"), json.Parser())
func (k *KoanfConfig) Load(ctx context.Context, provider koanf.Provider, parser koanf.Parser) error {
	if k == nil || k.k == nil {
		return ErrKoanfNilConfig.New("nil koanf config")
	}

	if provider == nil {
		return ErrKoanfInvalidProvider.New("provider is nil")
	}

	err := k.k.Load(provider, parser)
	if err != nil {
		return ErrKoanfLoadFailed.NewWithError("failed to load config", err, "provider", provider, "parser", parser)
	}

	return nil
}

// Merge merges another KoanfConfig into this one.
// The other config's values will override existing values.
//
// Example:
//
//	err := cfg.Merge(ctx, otherCfg)
func (k *KoanfConfig) Merge(ctx context.Context, other *KoanfConfig) error {
	if k == nil || k.k == nil {
		return ErrKoanfNilConfig.New("nil koanf config")
	}

	if other == nil || other.k == nil {
		return ErrKoanfNilConfig.New("other config is nil")
	}

	err := k.k.Merge(other.k)
	if err != nil {
		return ErrKoanfMergeFailed.NewWithError("failed to merge configs", err, "other", other)
	}

	return nil
}

// All returns all configuration as a map.
// This is useful for debugging or when you need the entire config structure.
func (k *KoanfConfig) All(ctx context.Context) map[string]any {
	if k == nil || k.k == nil {
		return nil
	}
	return k.k.All()
}

// Exists checks if a key exists in the configuration.
func (k *KoanfConfig) Exists(key string) bool {
	if k == nil || k.k == nil {
		return false
	}
	return k.k.Exists(key)
}

// Keys returns all keys in the configuration with the given prefix.
// If prefix is empty, all keys are returned.
func (k *KoanfConfig) Keys(prefix string) []string {
	if k == nil || k.k == nil {
		return nil
	}

	allKeys := k.k.Keys()
	if prefix == "" {
		return allKeys
	}

	// Filter keys by prefix
	var filtered []string
	prefixWithDot := prefix + k.delimiter
	for _, key := range allKeys {
		if key == prefix || strings.HasPrefix(key, prefixWithDot) {
			filtered = append(filtered, key)
		}
	}
	return filtered
}

// Delete deletes a key from the configuration.
func (k *KoanfConfig) Delete(key string) error {
	if k == nil || k.k == nil {
		return ErrKoanfNilConfig.New("nil koanf config", "key", key)
	}

	if key == "" {
		return ErrKoanfEmptyKey.New("empty key", "key", key)
	}

	// koanf v2 Delete returns void, so we check if key exists first
	if !k.k.Exists(key) {
		return ErrKoanfMissingValue.New("key not found", "key", key)
	}

	k.k.Delete(key)
	return nil
}

// Koanf returns the underlying koanf.Koanf instance.
// This can be used to access koanf-specific functionality not exposed by the wrapper.
func (k *KoanfConfig) Koanf() *koanf.Koanf {
	if k == nil {
		return nil
	}
	return k.k
}

// buildUnmarshalConf builds a koanf.UnmarshalConf from the provided options.
func buildUnmarshalConf(options ...any) koanf.UnmarshalConf {
	conf := koanf.UnmarshalConf{
		Tag: "koanf",
	}

	for _, opt := range options {
		switch o := opt.(type) {
		case UnmarshalOption:
			o(&conf)
		case koanf.UnmarshalConf:
			conf = o
		}
	}

	return conf
}
