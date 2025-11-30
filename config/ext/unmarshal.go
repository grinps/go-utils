package ext

import (
	"context"
	"reflect"

	"github.com/go-viper/mapstructure/v2"
	"github.com/grinps/go-utils/config"
)

// Unmarshal unmarshals configuration at the given key into the target struct.
// It extracts the config from context using config.ContextConfig and checks
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
//	ctx = config.ContextWithConfig(ctx, cfg)
//	var server ServerConfig
//	err := ext.Unmarshal(ctx, "server", &server)
//
//	// With options
//	err := ext.Unmarshal(ctx, "server", &server,
//	    ext.WithTagName("json"),
//	    ext.WithStrictMode())
func Unmarshal[T any](ctx context.Context, key string, target *T, options ...any) error {
	cfg := config.ContextConfig(ctx, true)
	return unmarshalWithConfig(ctx, cfg, key, target, options...)
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
//	err := ext.UnmarshalWithConfig(ctx, cfg, "server", &server)
func UnmarshalWithConfig[T any](ctx context.Context, cfg config.Config, key string, target *T, options ...any) error {
	return unmarshalWithConfig(ctx, cfg, key, target, options...)
}

// unmarshalWithConfig is the internal implementation for unmarshalling.
func unmarshalWithConfig[T any](ctx context.Context, cfg config.Config, key string, target *T, options ...any) error {
	if cfg == nil {
		return ErrExtNilConfig.New("config is nil", "key", key)
	}

	if target == nil {
		return ErrExtInvalidTarget.New("target is nil", "key", key)
	}

	// Validate target is a struct pointer
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return ErrExtInvalidTarget.New("target must be a pointer", "key", key, "type", targetType.Kind().String())
	}

	elemType := targetType.Elem()
	if elemType.Kind() != reflect.Struct {
		return ErrExtInvalidTarget.New("target must be a pointer to struct", "key", key, "type", elemType.Kind().String())
	}

	// Check if config implements MarshableConfig
	if mc, ok := cfg.(MarshableConfig); ok {
		return mc.Unmarshal(ctx, key, target, options...)
	}

	return ErrExtUnmarshalFailed.New("config does not implement MarshableConfig", "key", key)
}

// MustUnmarshal is like Unmarshal but panics if unmarshalling fails.
// It extracts the config from context.
// Use this only in initialization code where failure should be fatal.
//
// Example:
//
//	ctx = config.ContextWithConfig(ctx, cfg)
//	var server ServerConfig
//	ext.MustUnmarshal(ctx, "server", &server)
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
//	ctx = config.ContextWithConfig(ctx, cfg)
//	err := ext.SetValue(ctx, "server.port", 8080)
func SetValue(ctx context.Context, key string, value any) error {
	cfg := config.ContextConfig(ctx, true)
	return SetValueWithConfig(ctx, cfg, key, value)
}

// SetValueWithConfig sets a configuration value using the explicitly provided config.
// The config must implement MutableConfig, otherwise an error is returned.
//
// Example:
//
//	err := ext.SetValueWithConfig(ctx, cfg, "server.port", 8080)
func SetValueWithConfig(ctx context.Context, cfg config.Config, key string, value any) error {
	if cfg == nil {
		return ErrExtNilConfig.New("config is nil", "key", key)
	}

	if key == "" {
		return ErrExtSetValueFailed.New("empty key", "key", key)
	}

	// Check if config implements MutableConfig
	if mc, ok := cfg.(MutableConfig); ok {
		return mc.SetValue(ctx, key, value)
	}

	return ErrExtSetValueFailed.New("config does not implement MutableConfig", "key", key)
}

// buildDecoderConfig creates a mapstructure.DecoderConfig from unmarshalConfig.
func buildDecoderConfig(target any, cfg *unmarshalConfig) *mapstructure.DecoderConfig {
	decoderConfig := &mapstructure.DecoderConfig{
		Result:           target,
		TagName:          cfg.tagName,
		WeaklyTypedInput: cfg.weaklyTypedInput,
		Squash:           cfg.squash,
		ErrorUnused:      cfg.errorUnused,
		ErrorUnset:       cfg.errorUnset,
		Metadata:         cfg.metadata,
	}

	// Add decode hooks
	if len(cfg.decodeHooks) > 0 {
		// Compose all hooks with default hooks
		hooks := append([]mapstructure.DecodeHookFunc{
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		}, cfg.decodeHooks...)
		decoderConfig.DecodeHook = mapstructure.ComposeDecodeHookFunc(hooks...)
	} else {
		// Default hooks for common conversions
		decoderConfig.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}

	return decoderConfig
}

// newDecoder creates a new mapstructure decoder with the given config.
func newDecoder(config *mapstructure.DecoderConfig) (*mapstructure.Decoder, error) {
	return mapstructure.NewDecoder(config)
}
