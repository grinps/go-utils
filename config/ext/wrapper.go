package ext

import (
	"context"
	"reflect"

	"github.com/grinps/go-utils/config"
)

// ConfigWrapper wraps a config.Config and provides consistent MarshableConfig
// and MutableConfig capabilities across all Config implementations.
//
// It implements:
//   - config.Config: by delegating to the wrapped config
//   - MarshableConfig: using mapstructure if the original doesn't support it
//   - MutableConfig: by delegating if supported, otherwise returns an error
//
// Example:
//
//	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
//	wrapper := ext.NewConfigWrapper(cfg)
//
//	// Now use Unmarshal consistently
//	var server ServerConfig
//	err := wrapper.Unmarshal(ctx, "server", &server)
type ConfigWrapper struct {
	wrapped config.Config

	// Cached interface checks
	marshable MarshableConfig
	mutable   MutableConfig
}

// Ensure ConfigWrapper implements all required interfaces
var (
	_ config.Config   = (*ConfigWrapper)(nil)
	_ MarshableConfig = (*ConfigWrapper)(nil)
	_ MutableConfig   = (*ConfigWrapper)(nil)
)

// NewConfigWrapper creates a new ConfigWrapper that wraps the given config.
// If the wrapped config already implements MarshableConfig or MutableConfig,
// those implementations will be used. Otherwise, fallback behavior is provided.
//
// Example:
//
//	wrapper := ext.NewConfigWrapper(cfg)
//	var server ServerConfig
//	err := wrapper.Unmarshal(ctx, "server", &server)
func NewConfigWrapper(cfg config.Config) *ConfigWrapper {
	if cfg == nil {
		return nil
	}

	wrapper := &ConfigWrapper{
		wrapped: cfg,
	}

	// Cache interface checks for performance
	if mc, ok := cfg.(MarshableConfig); ok {
		wrapper.marshable = mc
	}
	if mc, ok := cfg.(MutableConfig); ok {
		wrapper.mutable = mc
	}

	return wrapper
}

// Unwrap returns the underlying config.Config.
// This can be used to access the original config if needed.
func (w *ConfigWrapper) Unwrap() config.Config {
	if w == nil {
		return nil
	}
	return w.wrapped
}

// GetValue delegates to the wrapped config's GetValue method.
// Implements config.Config interface.
func (w *ConfigWrapper) GetValue(ctx context.Context, key string, returnValue any) error {
	if w == nil || w.wrapped == nil {
		return ErrExtNilConfig.New("wrapper or wrapped config is nil", "key", key)
	}
	return w.wrapped.GetValue(ctx, key, returnValue)
}

// GetConfig delegates to the wrapped config's GetConfig method.
// Returns a new ConfigWrapper wrapping the sub-config.
// Implements config.Config interface.
func (w *ConfigWrapper) GetConfig(ctx context.Context, key string) (config.Config, error) {
	if w == nil || w.wrapped == nil {
		return nil, ErrExtNilConfig.New("wrapper or wrapped config is nil", "key", key)
	}

	subCfg, err := w.wrapped.GetConfig(ctx, key)
	if err != nil {
		return nil, err
	}

	// Wrap the sub-config to maintain consistent behavior
	return NewConfigWrapper(subCfg), nil
}

// Unmarshal unmarshals configuration at the given key into the target struct.
// If the wrapped config implements MarshableConfig, its method is used.
// Otherwise, falls back to mapstructure-based unmarshalling.
// Implements MarshableConfig interface.
//
// Example:
//
//	type ServerConfig struct {
//	    Host string `config:"host"`
//	    Port int    `config:"port"`
//	}
//	var server ServerConfig
//	err := wrapper.Unmarshal(ctx, "server", &server)
func (w *ConfigWrapper) Unmarshal(ctx context.Context, key string, target any, options ...UnmarshalOption) error {
	if w == nil || w.wrapped == nil {
		return ErrExtNilConfig.New("wrapper or wrapped config is nil", "key", key)
	}

	if target == nil {
		return ErrExtInvalidTarget.New("target is nil", "key", key)
	}

	// Validate target is a pointer to struct
	targetType := reflect.TypeOf(target)
	if targetType.Kind() != reflect.Ptr {
		return ErrExtInvalidTarget.New("target must be a pointer", "key", key, "type", targetType.Kind().String())
	}

	elemType := targetType.Elem()
	if elemType.Kind() != reflect.Struct {
		return ErrExtInvalidTarget.New("target must be a pointer to struct", "key", key, "type", elemType.Kind().String())
	}

	// Use native MarshableConfig if available
	if w.marshable != nil {
		return w.marshable.Unmarshal(ctx, key, target, options...)
	}

	// Fallback to mapstructure-based unmarshalling
	return unmarshalAny(ctx, w.wrapped, key, target, options...)
}

// SetValue sets a configuration value at the specified key.
// If the wrapped config implements MutableConfig, its method is used.
// Otherwise, returns ErrExtSetValueFailed.
// Implements MutableConfig interface.
//
// Example:
//
//	err := wrapper.SetValue(ctx, "server.port", 9090)
//	if errors.Is(err, ext.ErrExtSetValueFailed) {
//	    // Wrapped config doesn't support mutation
//	}
func (w *ConfigWrapper) SetValue(ctx context.Context, key string, newValue any) error {
	if w == nil || w.wrapped == nil {
		return ErrExtNilConfig.New("wrapper or wrapped config is nil", "key", key)
	}

	if key == "" {
		return ErrExtSetValueFailed.New("empty key", "key", key)
	}

	// Use native MutableConfig if available
	if w.mutable != nil {
		return w.mutable.SetValue(ctx, key, newValue)
	}

	// Wrapped config doesn't support mutation
	return ErrExtSetValueFailed.New(
		"wrapped config does not implement MutableConfig",
		"key", key,
		"config_type", reflect.TypeOf(w.wrapped).String(),
	)
}

// IsMutable returns true if the wrapped config supports SetValue.
func (w *ConfigWrapper) IsMutable() bool {
	return w != nil && w.mutable != nil
}

// IsMarshable returns true if the wrapped config has native Unmarshal support.
// Note: ConfigWrapper always provides Unmarshal capability via mapstructure fallback.
func (w *ConfigWrapper) IsMarshable() bool {
	return w != nil && w.marshable != nil
}

// All returns all configuration as a map if the wrapped config supports it.
// Returns nil if not supported.
func (w *ConfigWrapper) All(ctx context.Context) map[string]any {
	if w == nil || w.wrapped == nil {
		return nil
	}

	if allGetter, ok := w.wrapped.(interface {
		All(context.Context) map[string]any
	}); ok {
		return allGetter.All(ctx)
	}

	return nil
}

// unmarshalAny is a non-generic version of unmarshalWithMapstructure
// that works with any pointer type. Used by ConfigWrapper.Unmarshal.
func unmarshalAny(ctx context.Context, cfg config.Config, key string, target any, options ...UnmarshalOption) error {
	// Get the configuration value as a map
	var configMap map[string]any

	if key == "" {
		// Get the entire config
		var rawConfig any
		err := cfg.GetValue(ctx, "", &rawConfig)
		if err != nil {
			// Try alternative: the config itself might expose All()
			if allGetter, ok := cfg.(interface {
				All(context.Context) map[string]any
			}); ok {
				configMap = allGetter.All(ctx)
			} else {
				return ErrExtKeyNotFound.New("cannot get root config", "key", key, "error", err)
			}
		} else {
			var ok bool
			configMap, ok = rawConfig.(map[string]any)
			if !ok {
				return ErrExtUnmarshalFailed.New("root config is not a map", "key", key, "type", reflect.TypeOf(rawConfig).String())
			}
		}
	} else {
		// Get nested config as a map
		subCfg, err := cfg.GetConfig(ctx, key)
		if err != nil {
			return ErrExtKeyNotFound.New("key not found", "key", key, "error", err)
		}

		// Try to get the underlying map from the sub-config
		if allGetter, ok := subCfg.(interface {
			All(context.Context) map[string]any
		}); ok {
			configMap = allGetter.All(ctx)
		} else {
			// Try to retrieve as a map value
			var rawMap any
			err = cfg.GetValue(ctx, key, &rawMap)
			if err != nil {
				return ErrExtKeyNotFound.New("cannot get config value", "key", key, "error", err)
			}
			var ok bool
			configMap, ok = rawMap.(map[string]any)
			if !ok {
				return ErrExtUnmarshalFailed.New("config value is not a map", "key", key, "type", reflect.TypeOf(rawMap).String())
			}
		}
	}

	// Apply options and decode
	return decodeWithMapstructure(configMap, target, options...)
}

// decodeWithMapstructure decodes a map into the target struct using mapstructure.
// This is a shared helper used by both unmarshalWithMapstructure and unmarshalAny.
func decodeWithMapstructure(configMap map[string]any, target any, options ...UnmarshalOption) error {
	// Apply options
	unmarshalCfg := defaultUnmarshalConfig()
	for _, opt := range options {
		opt(unmarshalCfg)
	}

	// Build mapstructure decoder config
	decoderConfig := buildDecoderConfig(target, unmarshalCfg)

	decoder, err := newDecoder(decoderConfig)
	if err != nil {
		return ErrExtUnmarshalFailed.New("failed to create decoder", "error", err)
	}

	if err := decoder.Decode(configMap); err != nil {
		return ErrExtUnmarshalFailed.New("failed to decode config", "error", err)
	}

	return nil
}
