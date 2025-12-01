package config

import (
	"context"
	"reflect"
)

var defaultConfig Config = NewSimpleConfig(context.Background())

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
	if returnValue == nil {
		return ErrConfigNilReturnValue.New("nil return value", "key", key)
	}

	if key == "" {
		return ErrConfigEmptyKey.New("empty key", "key", key)
	}

	applicableConfig := ContextConfig(ctx, true)
	if applicableConfig == nil {
		return ErrConfigNilConfig.New("nil config", "key", key)
	}

	// Get the value
	val, err := applicableConfig.GetValue(ctx, key)
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
