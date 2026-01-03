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
	marshable config.MarshableConfig
	mutable   config.MutableConfig
	allGetter config.AllGetter
	allKeys   config.AllKeysProvider
	deleter   config.Deleter
}

// Ensure ConfigWrapper implements all required interfaces
var (
	_ config.Config          = (*ConfigWrapper)(nil)
	_ config.MarshableConfig = (*ConfigWrapper)(nil)
	_ config.MutableConfig   = (*ConfigWrapper)(nil)
	_ config.AllGetter       = (*ConfigWrapper)(nil)
	_ config.AllKeysProvider = (*ConfigWrapper)(nil)
	_ config.Deleter         = (*ConfigWrapper)(nil)
)

// Name returns the provider name for ConfigWrapper.
// Implements config.Config interface.
func (w *ConfigWrapper) Name() config.ProviderName {
	return "ConfigWrapper"
}

// ShouldInstrument always returns true for ConfigWrapper.
// Implements config.TelemetryAware interface.
func (w *ConfigWrapper) ShouldInstrument(ctx context.Context, key string, op string) bool {
	return true
}

// GenerateTelemetryAttributes returns the attributes as-is.
// Implements config.TelemetryAware interface.
func (w *ConfigWrapper) GenerateTelemetryAttributes(ctx context.Context, op string, attrs []any) []any {
	return attrs
}

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
	if mc, ok := cfg.(config.MarshableConfig); ok {
		wrapper.marshable = mc
	}
	if mc, ok := cfg.(config.MutableConfig); ok {
		wrapper.mutable = mc
	}
	if ag, ok := cfg.(config.AllGetter); ok {
		wrapper.allGetter = ag
	}
	if ak, ok := cfg.(config.AllKeysProvider); ok {
		wrapper.allKeys = ak
	}
	if d, ok := cfg.(config.Deleter); ok {
		wrapper.deleter = d
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
func (w *ConfigWrapper) GetValue(ctx context.Context, key string) (any, error) {
	if w == nil || w.wrapped == nil {
		return nil, ErrExtNilConfig.New("wrapper or wrapped config is nil", "key", key)
	}
	return w.wrapped.GetValue(ctx, key)
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
func (w *ConfigWrapper) Unmarshal(ctx context.Context, key string, target any, options ...any) error {
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
// Implements config.AllGetter interface.
func (w *ConfigWrapper) All(ctx context.Context) map[string]any {
	if w == nil || w.wrapped == nil {
		return map[string]any{}
	}

	if w.allGetter != nil {
		return w.allGetter.All(ctx)
	}

	return map[string]any{}
}

// Keys returns all keys in the configuration with the given prefix.
// If the wrapped config doesn't support AllKeysProvider, returns nil.
// Implements config.AllKeysProvider interface.
func (w *ConfigWrapper) Keys(prefix string) []string {
	if w == nil || w.wrapped == nil {
		return []string{}
	}

	if w.allKeys != nil {
		return w.allKeys.Keys(prefix)
	}

	return []string{}
}

// Delete deletes a key from the configuration.
// If the wrapped config doesn't support Deleter, returns ErrExtDeleteNotSupported.
// Implements config.Deleter interface.
func (w *ConfigWrapper) Delete(key string) error {
	if w == nil || w.wrapped == nil {
		return ErrExtNilConfig.New("wrapper or wrapped config is nil", "key", key)
	}

	if w.deleter != nil {
		return w.deleter.Delete(key)
	}

	return ErrExtDeleteNotSupported.New(
		"wrapped config does not implement Deleter",
		"key", key,
		"config_type", reflect.TypeOf(w.wrapped).String(),
	)
}

// HasAllGetter returns true if the wrapped config supports All().
func (w *ConfigWrapper) HasAllGetter() bool {
	return w != nil && w.allGetter != nil
}

// HasAllKeys returns true if the wrapped config supports Keys().
func (w *ConfigWrapper) HasAllKeys() bool {
	return w != nil && w.allKeys != nil
}

// HasDeleter returns true if the wrapped config supports Delete().
func (w *ConfigWrapper) HasDeleter() bool {
	return w != nil && w.deleter != nil
}

// unmarshalAny is a non-generic version of unmarshalWithMapstructure
// that works with any pointer type. Used by ConfigWrapper.Unmarshal.
func unmarshalAny(ctx context.Context, cfg config.Config, key string, target any, options ...any) error {
	// Get the configuration value as a map
	var configMap map[string]any

	if key == "" {
		// Get the entire config
		rawConfig, err := cfg.GetValue(ctx, "")
		if err != nil {
			// Try alternative: the config itself might expose All()
			if allGetter, ok := cfg.(config.AllGetter); ok {
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
		if allGetter, ok := subCfg.(config.AllGetter); ok {
			configMap = allGetter.All(ctx)
		} else {
			// Try to retrieve as a map value
			rawMap, err := cfg.GetValue(ctx, key)
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
