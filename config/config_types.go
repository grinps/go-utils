package config

import (
	"context"
)

// DefaultKeyDelimiter is the default separator used for nested key access.
// For example, "server.port" accesses config["server"]["port"].
const DefaultKeyDelimiter = "."

// Config defines the interface for configuration management.
// Implementations should support nested configuration access via dot-notation keys.
type Config interface {
	// GetValue retrieves a configuration value by key and stores it in the provided pointer.
	// Keys can use dot-notation for nested access (e.g., "server.port").
	// The returnValue parameter must be a non-nil pointer to store the result.
	// Returns an error if the key is empty, value is not found, or type conversion fails.
	//
	// Example:
	//   var port int
	//   err := cfg.GetValue(ctx, "server.port", &port)
	GetValue(ctx context.Context, key string, returnValue any) error

	// GetConfig retrieves a nested configuration as a new Config instance.
	// The value at the key must be convertible to a configuration map.
	GetConfig(ctx context.Context, key string) (Config, error)
}
