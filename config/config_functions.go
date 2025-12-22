package config

import (
	"context"
	"reflect"
)

var defaultConfig Config = NewSimpleConfig(context.Background())

// SetAsDefault sets the provided Config as the package-level default configuration.
// This default is used when ContextConfig is called with defaultIfNotAvailable=true
// and no config is found in the context.
//
// Example:
//
//	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
//	config.SetAsDefault(cfg)
func SetAsDefault(cfg Config) {
	defaultConfig = cfg
}

// GetValueE retrieves a configuration value from the context and stores it in returnValue.
// If returnValue already contains a value, it acts as the default and is preserved if the key is not found.
// Returns an error if:
//   - The returnValue pointer is nil
//   - The key is empty
//   - The config is nil
//   - The value cannot be converted to type T
//
// Example:
//
//	path := "./logs" // default value
//	err := config.GetValueE(ctx, "server.log_path", &path)
//	// path will be "./logs" if key not found, or the configured value if found
func GetValueE[T any](ctx context.Context, key string, returnValue *T) error {
	applicableConfig := ContextConfig(ctx, true)
	return GetValueWithConfig(ctx, applicableConfig, key, returnValue)
}

// GetValueWithConfig retrieves a configuration value from the provided config and stores it in returnValue.
// If returnValue already contains a value, it acts as the default and is preserved if the key is not found.
// Returns an error if:
//   - The returnValue pointer is nil
//   - The key is empty
//   - The config is nil
//   - The value cannot be converted to type T
//
// Example:
//
//	path := "./logs" // default value
//	err := config.GetValueWithConfig(ctx, cfg, "server.log_path", &path)
//	// path will be "./logs" if key not found, or the configured value if found
func GetValueWithConfig[T any](ctx context.Context, cfg Config, key string, returnValue *T) error {
	if returnValue == nil {
		return ErrConfigNilReturnValue.New("nil return value", "key", key)
	}

	if key == "" {
		return ErrConfigEmptyKey.New("empty key", "key", key)
	}

	if cfg == nil {
		return ErrConfigNilConfig.New("nil config", "key", key)
	}

	// Get the value
	val, err := cfg.GetValue(ctx, key)
	if err != nil {
		return err
	}

	// Type assertion and assignment
	typedVal, ok := val.(T)
	if !ok {
		return ErrConfigInvalidValueType.New("value type mismatch", "key", key, "expected_type", reflect.TypeOf(returnValue).Elem().String(), "actual_type", reflect.TypeOf(val).String())
	}

	*returnValue = typedVal
	return nil
}

type contextConfigType struct{}

// ContextConfigReference is the key used to store Config in context.Context.
var ContextConfigReference contextConfigType = contextConfigType{}

// ContextWithConfig returns a new context with the given Config attached.
// If ctx is nil, a background context is created automatically.
//
// Example:
//
//	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
//	ctx = config.ContextWithConfig(ctx, cfg)
func ContextWithConfig(ctx context.Context, config Config) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ContextConfigReference, config)
}

// ContextConfig retrieves the Config from the given context.
// If defaultIfNotAvailable is true and no config is found, returns the default config.
// If defaultIfNotAvailable is false and no config is found, returns nil.
//
// Example:
//
//	cfg := config.ContextConfig(ctx, true) // Returns default if not in context
//	cfg := config.ContextConfig(ctx, false) // Returns nil if not in context
func ContextConfig(ctx context.Context, defaultIfNotAvailable bool) Config {
	if ctx != nil {
		config := ctx.Value(ContextConfigReference)
		if config != nil {
			if asConfig, isConfig := config.(Config); isConfig {
				return asConfig
			}
		}
	}
	if defaultIfNotAvailable {
		return defaultConfig
	} else {
		return nil
	}
}

// Default returns the default Config instance.
// This is a package-level SimpleConfig that can be used when no custom config is needed.
//
// Example:
//
//	cfg := config.Default()
//	val, err := cfg.GetValue(ctx, "key")
func Default() Config {
	return defaultConfig
}

// GetConfig retrieves a nested configuration from the context config at the given key.
// The value at the key must be convertible to a configuration map.
//
// Example:
//
//	ctx = config.ContextWithConfig(ctx, cfg)
//	serverCfg, err := config.GetConfig(ctx, "server")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	val, _ := serverCfg.GetValue(ctx, "port")
func GetConfig(ctx context.Context, key string) (Config, error) {
	applicableConfig := ContextConfig(ctx, true)
	return GetConfigWithConfig(ctx, applicableConfig, key)
}

// GetConfigWithConfig retrieves a nested configuration from the provided config at the given key.
// The value at the key must be convertible to a configuration map.
//
// Example:
//
//	serverCfg, err := config.GetConfigWithConfig(ctx, cfg, "server")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	val, _ := serverCfg.GetValue(ctx, "port")
func GetConfigWithConfig(ctx context.Context, cfg Config, key string) (Config, error) {
	if cfg == nil {
		return nil, ErrConfigNilConfig.New("nil config", "key", key)
	}

	if key == "" {
		return nil, ErrConfigEmptyKey.New("empty key", "key", key)
	}

	return cfg.GetConfig(ctx, key)
}
