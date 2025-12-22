package otel

import (
	"context"

	"github.com/grinps/go-utils/config"
	"go.opentelemetry.io/contrib/otelconf"
	"gopkg.in/yaml.v3"
)

// ConfigKey is the key used to retrieve OpenTelemetry configuration from config.Config.
const ConfigKey = "opentelemetry"

// LoadConfiguration loads otelconf.OpenTelemetryConfiguration from a config.Config.
// It retrieves the raw config map, converts it to YAML, and uses otelconf.ParseYAML
// to properly parse the configuration (including ResourceJSON and other complex types).
func LoadConfiguration(ctx context.Context, cfg config.Config) (*otelconf.OpenTelemetryConfiguration, error) {
	// Get the raw config value as a map
	rawValue, err := cfg.GetValue(ctx, ConfigKey)
	if err != nil {
		return nil, ErrCodeConfigLoadFailed.NewWithError("failed to get configuration", err)
	}
	if rawValue == nil {
		return nil, ErrCodeConfigLoadFailed.New("configuration key not found")
	}

	// Ensure the config is a map and has file_format (required by otelconf.ParseYAML)
	if rawMap, ok := rawValue.(map[string]any); ok {
		if _, hasFileFormat := rawMap["file_format"]; !hasFileFormat {
			rawMap["file_format"] = "0.3"
		}
	}

	// Convert to YAML bytes
	yamlBytes, err := yaml.Marshal(rawValue)
	if err != nil {
		return nil, ErrCodeConfigLoadFailed.NewWithError("failed to marshal configuration to YAML", err)
	}

	// Use otelconf.ParseYAML to properly parse the configuration
	otelCfg, err := otelconf.ParseYAML(yamlBytes)
	if err != nil {
		return nil, ErrCodeConfigLoadFailed.NewWithError("failed to parse YAML configuration", err)
	}
	return otelCfg, nil
}

// DefaultConfiguration returns a default OpenTelemetryConfiguration.
func DefaultConfiguration() *otelconf.OpenTelemetryConfiguration {
	return &otelconf.OpenTelemetryConfiguration{
		FileFormat: "0.3",
	}
}

// StatusCode represents the status of a span.
type StatusCode int

const (
	StatusUnset StatusCode = iota
	StatusOK
	StatusError
)
