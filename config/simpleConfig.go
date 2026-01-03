package config

import (
	"context"
	"fmt"
	"strings"
)

// SimpleConfig is an alias for *simpleConfig, providing an in-memory implementation of the Config interface.
// It supports nested configuration maps accessed via dot-notation keys.
type SimpleConfig = *simpleConfig

type simpleConfig struct {
	configurationMap map[string]any
	delimiter        string
}

// Name returns the provider name for SimpleConfig.
// This implements the Config interface's Name() method.
func (cfg *simpleConfig) Name() ProviderName {
	return "SimpleConfig"
}

// GetValue retrieves a value from the configuration using dot-notation keys.
// Returns the value and an error if the config is nil, the key is empty, or the value is not found.
//
// Example:
//
//	port, err := cfg.GetValue(ctx, "server.port")
//	if err != nil {
//	    // handle error
//	}
//	portInt := port.(int)
func (cfg *simpleConfig) GetValue(ctx context.Context, key string) (any, error) {
	if cfg == nil || cfg.configurationMap == nil {
		return nil, ErrConfigNilConfig.New("nil config or map", "key", key)
	}

	if key == "" {
		return nil, ErrConfigEmptyKey.New("empty key", "key", key)
	}

	keyParts := strings.Split(key, cfg.delimiter)
	var currentVal any = cfg.configurationMap

	for i, part := range keyParts {
		if currentVal == nil {
			return nil, ErrConfigMissingValue.New("nil value at intermediate key", "key", key, "intermediate_key", strings.Join(keyParts[:i], DefaultKeyDelimiter))
		}

		if m, ok := currentVal.(map[string]any); ok {
			if val, found := m[part]; found {
				currentVal = val
			} else {
				return nil, ErrConfigMissingValue.New("missing value", "key", key)
			}
		} else {
			return nil, ErrConfigInvalidValue.New("intermediate value is not a map", "key", key, "current_value", currentVal)
		}
	}

	if currentVal == nil {
		return nil, ErrConfigMissingValue.New("value is nil", "key", key)
	}

	return currentVal, nil
}

// GetAsMap converts various map types to map[string]any.
// Supports:
//   - map[string]any (returned as-is)
//   - map[string]string (converted)
//   - map[any]any (converted if all keys are strings)
//   - nil (returns nil)
//
// Returns an error for unsupported types or non-string keys in map[any]any.
//
// Example:
//
//	m, err := config.GetAsMap(ctx, map[string]string{"key": "value"})
func GetAsMap(ctx context.Context, input any) (map[string]any, error) {
	applicableConfigMap := map[string]any{}
	if input == nil {
		return map[string]any(nil), nil
	}
	switch v := input.(type) {
	case map[string]any:
		applicableConfigMap = v
	case map[string]string:
		for k, val := range v {
			applicableConfigMap[k] = val
		}
	case map[any]any:
		for k, val := range v {
			if strKey, ok := k.(string); ok {
				applicableConfigMap[strKey] = val
			} else {
				return nil, ErrConfigInvalidValue.New("key is not a string", "invalid_key", k)
			}
		}
	default:
		return nil, ErrConfigInvalidValue.New(fmt.Sprintf("conversion of configuration %v of type %T to map[string]any not supported", input, input))
	}
	return applicableConfigMap, nil
}

// GetConfig retrieves a nested configuration as a new Config instance.
// The value at the given key must be convertible to a map.
//
// Example:
//
//	serverCfg, err := cfg.GetConfig(ctx, "server")
//	var port int
//	err = serverCfg.GetValue(ctx, "port", &port)
func (cfg *simpleConfig) GetConfig(ctx context.Context, key string) (Config, error) {
	val, err := cfg.GetValue(ctx, key)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, ErrConfigMissingValue.New("missing value", "key", key)
	}

	mapVal, mapErr := GetAsMap(ctx, val)
	if mapErr != nil {
		return nil, mapErr
	}

	return &simpleConfig{
		configurationMap: mapVal,
		delimiter:        cfg.delimiter,
	}, nil
}

// SetValue sets a configuration value using dot-notation keys.
// Creates intermediate maps as needed.
// Returns an error if the config is nil, the key is empty, or an intermediate value is not a map.
//
// Example:
//
//	err := cfg.SetValue(ctx, "server.port", 8080)
//	err := cfg.SetValue(ctx, "database.host", "localhost")
func (cfg *simpleConfig) SetValue(ctx context.Context, key string, value any) error {
	if cfg == nil || cfg.configurationMap == nil {
		return ErrConfigNilConfig.New("nil config or map", "key", key)
	}

	if key == "" {
		return ErrConfigEmptyKey.New("empty key", "key", key)
	}

	keyParts := strings.Split(key, cfg.delimiter)
	var currentMap = cfg.configurationMap

	for i, part := range keyParts {
		// Last part of the key, set the value
		if i == len(keyParts)-1 {
			currentMap[part] = value
			return nil
		}

		// Intermediate parts
		val, found := currentMap[part]
		if !found {
			// Create new map if missing
			newMap := make(map[string]any)
			currentMap[part] = newMap
			currentMap = newMap
		} else {
			if m, ok := val.(map[string]any); ok {
				currentMap = m
			} else {
				return ErrConfigInvalidValue.New("intermediate value is not a map", "key", key, "current_value", val)
			}
		}
	}

	return nil
}

// All returns all configuration as a map.
// Implements config.AllGetter interface.
//
// Example:
//
//	all := cfg.All(ctx)
//	for key, value := range all {
//	    fmt.Printf("%s: %v\n", key, value)
//	}
func (cfg *simpleConfig) All(ctx context.Context) map[string]any {
	if cfg == nil || cfg.configurationMap == nil {
		return map[string]any{}
	}
	return cfg.configurationMap
}

// Keys returns all keys in the configuration with the given prefix.
// If prefix is empty, all keys are returned.
// Implements config.AllKeysProvider interface.
//
// Example:
//
//	allKeys := cfg.Keys("")           // Get all keys
//	serverKeys := cfg.Keys("server")  // Get keys starting with "server"
func (cfg *simpleConfig) Keys(prefix string) []string {
	if cfg == nil || cfg.configurationMap == nil {
		return []string{}
	}

	keys := cfg.collectKeys(cfg.configurationMap, "")

	if prefix == "" {
		return keys
	}

	// Filter keys by prefix
	var filtered []string
	prefixWithDot := prefix + cfg.delimiter
	for _, key := range keys {
		if key == prefix || strings.HasPrefix(key, prefixWithDot) {
			filtered = append(filtered, key)
		}
	}
	return filtered
}

// collectKeys recursively collects all keys from a nested map.
func (cfg *simpleConfig) collectKeys(m map[string]any, prefix string) []string {
	var keys []string
	for k, v := range m {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + cfg.delimiter + k
		}

		keys = append(keys, fullKey)

		// Recurse into nested maps
		if nested, ok := v.(map[string]any); ok {
			keys = append(keys, cfg.collectKeys(nested, fullKey)...)
		}
	}
	return keys
}

// Delete deletes a key from the configuration.
// Returns an error if the key is empty or not found.
// Implements config.Deleter interface.
//
// Example:
//
//	err := cfg.Delete("server.debug")
func (cfg *simpleConfig) Delete(key string) error {
	if cfg == nil || cfg.configurationMap == nil {
		return ErrConfigNilConfig.New("nil config or map", "key", key)
	}

	if key == "" {
		return ErrConfigEmptyKey.New("empty key", "key", key)
	}

	keyParts := strings.Split(key, cfg.delimiter)
	currentMap := cfg.configurationMap

	// Navigate to the parent map
	for i := 0; i < len(keyParts)-1; i++ {
		part := keyParts[i]
		val, found := currentMap[part]
		if !found {
			return ErrConfigMissingValue.New("key not found", "key", key)
		}
		if m, ok := val.(map[string]any); ok {
			currentMap = m
		} else {
			return ErrConfigMissingValue.New("key not found", "key", key)
		}
	}

	// Delete the final key
	finalKey := keyParts[len(keyParts)-1]
	if _, found := currentMap[finalKey]; !found {
		return ErrConfigMissingValue.New("key not found", "key", key)
	}

	delete(currentMap, finalKey)
	return nil
}

// SimpleConfigOption is a function that configures a simpleConfig instance.
type SimpleConfigOption func(cfg *simpleConfig)

// WithDelimiter sets a custom delimiter for key parsing (default is ".").
//
// Example:
//
//	cfg := config.NewSimpleConfig(ctx, config.WithDelimiter("/"))
//	val, _ := cfg.GetValue(ctx, "server/port") // Uses / instead of .
func WithDelimiter(delimiter string) SimpleConfigOption {
	return func(cfg *simpleConfig) {
		cfg.delimiter = delimiter
	}
}

// WithConfigurationMap initializes the config with an existing map.
//
// Example:
//
//	data := map[string]any{"server": map[string]any{"port": 8080}}
//	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
func WithConfigurationMap(configurationMap map[string]any) SimpleConfigOption {
	return func(cfg *simpleConfig) {
		cfg.configurationMap = configurationMap
	}
}

// NewSimpleConfig creates a new in-memory Config instance with optional configuration.
// By default, uses "." as the key delimiter and an empty configuration map.
//
// Example:
//
//	// Empty config
//	cfg := config.NewSimpleConfig(ctx)
//
//	// With initial data
//	data := map[string]any{"key": "value"}
//	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
//
//	// With custom delimiter
//	cfg := config.NewSimpleConfig(ctx, config.WithDelimiter("/"))
func NewSimpleConfig(ctx context.Context, options ...SimpleConfigOption) Config {
	cfg := &simpleConfig{
		configurationMap: make(map[string]any),
		delimiter:        DefaultKeyDelimiter,
	}
	for _, option := range options {
		option(cfg)
	}
	return cfg
}
