package config

import (
	"context"
)

// DefaultKeyDelimiter is the default separator used for nested key access.
// For example, "server.port" accesses config["server"]["port"].
const DefaultKeyDelimiter = "."

// ProviderName identifies the config instance name.
type ProviderName string

// Config defines the interface for configuration management.
// Implementations should support nested configuration access via dot-notation keys.
type Config interface {
	// Name returns the name of the provider implementation (e.g., "SimpleConfig", "KoanfConfig").
	// This may be used to provide additional context for the config instance.
	Name() ProviderName

	// GetValue retrieves a configuration value by key.
	// Keys can use dot-notation for nested access (e.g., "server.port").
	// Returns the value and an error if the key is empty or value is not found.
	//
	// Example:
	//   port, err := cfg.GetValue(ctx, "server.port")
	//   if err != nil {
	//       // handle error
	//   }
	//   portInt := port.(int)
	GetValue(ctx context.Context, key string) (any, error)

	// GetConfig retrieves a nested configuration as a new Config instance.
	// The value at the key must be convertible to a configuration map.
	GetConfig(ctx context.Context, key string) (Config, error)
}

// MutableConfig defines the ability to set values.
// Implementations should handle nested key creation automatically.
type MutableConfig interface {
	Config

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
	Config

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

// AllGetter is an optional interface for configs that can return all values as a map.
// This is useful for debugging, serialization, or when you need the entire config structure.
type AllGetter interface {
	// All returns all configuration as a map.
	// Returns nil if the config is nil or empty.
	//
	// Example:
	//   all := cfg.All(ctx)
	//   for key, value := range all {
	//       fmt.Printf("%s: %v\n", key, value)
	//   }
	All(ctx context.Context) map[string]any
}

// AllKeysProvider is an optional interface for configs that can list all keys.
// This is useful for discovering available configuration keys.
type AllKeysProvider interface {
	// Keys returns all keys in the configuration with the given prefix.
	// If prefix is empty, all keys are returned.
	// Returns nil if the config is nil.
	//
	// Example:
	//   allKeys := cfg.Keys("")           // Get all keys
	//   serverKeys := cfg.Keys("server")  // Get keys starting with "server"
	Keys(prefix string) []string
}

// Deleter is an optional interface for configs that support key deletion.
type Deleter interface {
	// Delete deletes a key from the configuration.
	// Returns an error if the key is empty or not found.
	//
	// Example:
	//   err := cfg.Delete("server.debug")
	Delete(key string) error
}
