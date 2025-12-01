package config

import (
	"context"
	"reflect"
)

// Unmarshal unmarshals configuration at the given key into the target struct.
// It extracts the config from context using ContextConfig and checks
// if it implements MarshableConfig.
//
// The target must be a non-nil pointer to a struct.
// If key is empty, the entire configuration is unmarshalled.
//
// Example:
//
//	type ServerConfig struct {
//	    Host string `config:"host"`
//	    Port int    `config:"port"`
//	}
//	ctx = ContextWithConfig(ctx, cfg)
//	var server ServerConfig
//	err := Unmarshal(ctx, "server", &server)
//
//	// With options (implementation-specific)
//	err := Unmarshal(ctx, "server", &server, someOption)
func Unmarshal[T any](ctx context.Context, key string, target *T, options ...any) error {
	cfg := ContextConfig(ctx, true)
	return UnmarshalWithConfig(ctx, cfg, key, target, options...)
}

// UnmarshalWithConfig unmarshals configuration at the given key into the target struct
// using the explicitly provided config.
// It checks if the config implements MarshableConfig and uses its Unmarshal method.
//
// The target must be a non-nil pointer to a struct.
// If key is empty, the entire configuration is unmarshalled.
//
// Example:
//
//	type ServerConfig struct {
//	    Host string `config:"host"`
//	    Port int    `config:"port"`
//	}
//	var server ServerConfig
//	err := UnmarshalWithConfig(ctx, cfg, "server", &server)
func UnmarshalWithConfig[T any](ctx context.Context, cfg Config, key string, target *T, options ...any) error {
	if cfg == nil {
		return ErrConfigNilConfig.New("config is nil", "key", key)
	}

	if target == nil {
		return ErrConfigInvalidTarget.New("target is nil", "key", key)
	}

	// Validate target is a struct pointer
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return ErrConfigInvalidTarget.New("target must be a pointer", "key", key, "type", targetType.Kind().String())
	}

	elemType := targetType.Elem()
	if elemType.Kind() != reflect.Struct {
		return ErrConfigInvalidTarget.New("target must be a pointer to struct", "key", key, "type", elemType.Kind().String())
	}

	// Check if config implements MarshableConfig
	if mc, ok := cfg.(MarshableConfig); ok {
		return mc.Unmarshal(ctx, key, target, options...)
	}

	return ErrConfigUnmarshalFailed.New("config does not implement MarshableConfig", "key", key)
}

// MustUnmarshal is like Unmarshal but panics if unmarshalling fails.
// It extracts the config from context.
// Use this only in initialization code where failure should be fatal.
//
// Example:
//
//	ctx = ContextWithConfig(ctx, cfg)
//	var server ServerConfig
//	MustUnmarshal(ctx, "server", &server)
func MustUnmarshal[T any](ctx context.Context, key string, target *T, options ...any) {
	if err := Unmarshal(ctx, key, target, options...); err != nil {
		panic("config: failed to unmarshal " + key + ": " + err.Error())
	}
}

// SetValue sets a configuration value using the config from context.
// The config must implement MutableConfig, otherwise an error is returned.
//
// Example:
//
//	ctx = ContextWithConfig(ctx, cfg)
//	err := SetValue(ctx, "server.port", 8080)
func SetValue(ctx context.Context, key string, value any) error {
	cfg := ContextConfig(ctx, true)
	return SetValueWithConfig(ctx, cfg, key, value)
}

// SetValueWithConfig sets a configuration value using the explicitly provided config.
// The config must implement MutableConfig, otherwise an error is returned.
//
// Example:
//
//	err := SetValueWithConfig(ctx, cfg, "server.port", 8080)
func SetValueWithConfig(ctx context.Context, cfg Config, key string, value any) error {
	if cfg == nil {
		return ErrConfigNilConfig.New("config is nil", "key", key)
	}

	if key == "" {
		return ErrConfigSetValueFailed.New("empty key", "key", key)
	}

	// Check if config implements MutableConfig
	if mc, ok := cfg.(MutableConfig); ok {
		return mc.SetValue(ctx, key, value)
	}

	return ErrConfigSetValueFailed.New("config does not implement MutableConfig", "key", key)
}
