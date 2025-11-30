// Package ext provides extended configuration interfaces and utilities
// that build upon the base config package.
//
// This package defines additional interfaces for mutable configurations
// and struct unmarshalling capabilities, along with helper functions
// that work with any config.Config implementation.
package ext

import (
	"context"
)

// MutableConfig defines the ability to set values.
// Implementations should handle nested key creation automatically.
type MutableConfig interface {

	// SetValue sets a configuration value at the specified key.
	// Keys can use dot-notation for nested access (e.g., "server.port").
	// Intermediate maps are created automatically if they don't exist.
	// Returns an error if the key is empty or if an intermediate value
	// is not a map and cannot be replaced.
	//
	// Example:
	//   err := cfg.SetValue(ctx, "server.port", 8080)
	//   err := cfg.SetValue(ctx, "database.credentials.password", "secret")
	SetValue(ctx context.Context, key string, newValue any) error
}

// MarshableConfig provides struct unmarshalling capabilities.
// This interface is for configurations that can efficiently unmarshal
// their data directly to structs using their internal representation.
type MarshableConfig interface {

	// Unmarshal unmarshals the configuration at the given key into the target struct.
	// If key is empty, the entire configuration is unmarshalled.
	// The target must be a pointer to a struct.
	// Options are implementation-specific; each implementation defines its own option types.
	//
	// Example:
	//   type ServerConfig struct {
	//       Host string `config:"host"`
	//       Port int    `config:"port"`
	//   }
	//   var server ServerConfig
	//   err := cfg.Unmarshal(ctx, "server", &server)
	Unmarshal(ctx context.Context, key string, target any, options ...any) error
}
