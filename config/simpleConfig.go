package config

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// SimpleConfig is an alias for *simpleConfig, providing an in-memory implementation of the Config interface.
// It supports nested configuration maps accessed via dot-notation keys.
type SimpleConfig = *simpleConfig

type simpleConfig struct {
	configurationMap map[string]any
	delimiter        string
}

// GetValue retrieves a value from the configuration using dot-notation keys and stores it in returnValue.
// The returnValue parameter must be a non-nil pointer.
// Returns an error if the config is nil, the key is empty, the value is not found, or type conversion fails.
//
// Example:
//
//	var port int
//	err := cfg.GetValue(ctx, "server.port", &port)
//
//	var host string
//	err := cfg.GetValue(ctx, "database.host", &host)
func (cfg *simpleConfig) GetValue(ctx context.Context, key string, returnValue any) error {
	if returnValue == nil {
		return ErrConfigNilReturnValue.New("nil return value", "key", key)
	}

	// Check if returnValue is a pointer
	rv := reflect.ValueOf(returnValue)
	if rv.Kind() != reflect.Ptr {
		return ErrConfigNilReturnValue.New("return value must be a pointer", "key", key, "type", rv.Kind().String())
	}

	if rv.IsNil() {
		return ErrConfigNilReturnValue.New("nil return value pointer", "key", key)
	}

	if cfg == nil || cfg.configurationMap == nil {
		return ErrConfigNilConfig.New("nil config or map", "key", key)
	}

	if key == "" {
		return ErrConfigEmptyKey.New("empty key", "key", key)
	}

	keyParts := strings.Split(key, cfg.delimiter)
	var currentVal any = cfg.configurationMap

	for i, part := range keyParts {
		if currentVal == nil {
			return ErrConfigMissingValue.New("nil value at intermediate key", "key", key, "intermediate_key", strings.Join(keyParts[:i], DefaultKeyDelimiter))
		}

		if m, ok := currentVal.(map[string]any); ok {
			if val, found := m[part]; found {
				currentVal = val
			} else {
				return ErrConfigMissingValue.New("missing value", "key", key)
			}
		} else {
			return ErrConfigInvalidValue.New("intermediate value is not a map", "key", key, "current_value", currentVal)
		}
	}

	// Type-safe assignment using reflection
	if currentVal == nil {
		return ErrConfigMissingValue.New("value is nil", "key", key)
	}

	targetValue := rv.Elem()
	sourceValue := reflect.ValueOf(currentVal)

	if !sourceValue.Type().AssignableTo(targetValue.Type()) {
		return ErrConfigInvalidValueType.New("value type mismatch", "key", key, "expected_type", targetValue.Type().String(), "actual_type", sourceValue.Type().String(), "actual_value", currentVal)
	}

	targetValue.Set(sourceValue)
	return nil
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
	var val any
	err := cfg.GetValue(ctx, key, &val)
	if err != nil {
		return nil, err
	}

	if val == nil {
		return nil, ErrConfigMissingValue.New("missing value", "key", key)
	}

	var mapVal map[string]any
	var mapErr error

	mapVal, mapErr = GetAsMap(ctx, val)

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
