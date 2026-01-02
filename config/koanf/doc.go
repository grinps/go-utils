// Package koanf provides a wrapper around knadh/koanf that implements the
// config.Config, config.MutableConfig, and config.MarshableConfig interfaces.
//
// This package bridges the powerful koanf configuration library with the
// standardized config interfaces, enabling seamless integration with the
// broader config ecosystem while leveraging koanf's extensive provider and
// parser support.
//
// # Features
//
//   - Implements config.Config, config.MutableConfig, and config.MarshableConfig interfaces
//   - Implements config.TelemetryAware for telemetry integration
//   - Support for multiple configuration sources (files, env vars, command-line flags, etc.)
//   - Nested configuration access via dot-notation keys
//   - Type-safe unmarshalling to structs with multiple tag support (koanf, json, yaml, mapstructure)
//   - Mutable configuration with SetValue
//   - Provider-based configuration loading with parsers
//   - Configuration merging from multiple sources
//   - Structured error handling using errext package
//   - High test coverage (>94%)
//
// # Basic Usage
//
// Creating a new configuration:
//
//	import (
//	    "context"
//	    "github.com/grinps/go-utils/config/koanf"
//	)
//
//	func main() {
//	    ctx := context.Background()
//
//	    // Create empty config
//	    cfg, err := koanf.NewKoanfConfig(ctx)
//
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//
//	    // Set values
//	    cfg.(*koanf.KoanfConfig).SetValue(ctx, "server.port", 8080)
//	    cfg.(*koanf.KoanfConfig).SetValue(ctx, "server.host", "localhost")
//	}
//
// # Loading from Providers
//
// Koanf supports loading configuration from various sources:
//
//	import (
//	    "github.com/knadh/koanf/parsers/json"
//	    "github.com/knadh/koanf/parsers/yaml"
//	    "github.com/knadh/koanf/providers/file"
//	    "github.com/knadh/koanf/providers/env"
//	)
//
//	// Load from JSON file
//	cfg, err := koanf.NewKoanfConfig(ctx,
//	    koanf.WithProvider(file.Provider("config.json"), json.Parser()),
//	)
//
//	// Or load after creation
//	cfg, err := koanf.NewKoanfConfig(ctx)
//	err := cfg.(*koanf.KoanfConfig).Load(ctx, file.Provider("config.yaml"), yaml.Parser())
//
//	// Load from environment variables
//	err = cfg.(*koanf.KoanfConfig).Load(ctx, env.Provider("APP_", ".", nil), nil)
//
// # Retrieving Values
//
// Values can be retrieved using the Config interface:
//
//	// Get a simple value
//	port, err := cfg.GetValue(ctx, "server.port")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	portInt := port.(int)
//
//	// Get a nested config
//	serverCfg, err := cfg.GetConfig(ctx, "server")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	host, err := serverCfg.GetValue(ctx, "host")
//
// # Unmarshalling to Structs
//
// The package supports type-safe unmarshalling with multiple tag formats:
//
//	type ServerConfig struct {
//	    Host string `koanf:"host"`
//	    Port int    `koanf:"port"`
//	}
//
//	type DatabaseConfig struct {
//	    Host     string `koanf:"host"`
//	    Port     int    `koanf:"port"`
//	    Username string `koanf:"username"`
//	}
//
//	type AppConfig struct {
//	    Server   ServerConfig   `koanf:"server"`
//	    Database DatabaseConfig `koanf:"database"`
//	}
//
//	// Unmarshal entire config
//	var app AppConfig
//	err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "", &app)
//
//	// Unmarshal sub-config
//	var server ServerConfig
//	err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "server", &server)
//
//	// Use different tag formats
//	type JSONConfig struct {
//	    Host string `json:"host"`
//	    Port int    `json:"port"`
//	}
//	var jsonCfg JSONConfig
//	err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "server", &jsonCfg, koanf.WithJSONTag())
//
// # Merging Configurations
//
// Multiple configurations can be merged together:
//
//	// Load base config
//	baseCfg := koanf.NewKoanfConfig(ctx,
//	    koanf.WithProvider(file.Provider("base.json"), json.Parser()),
//	)
//
//	// Load override config
//	overrideCfg := koanf.NewKoanfConfig(ctx,
//	    koanf.WithProvider(file.Provider("override.json"), json.Parser()),
//	)
//
//	// Merge override into base
//	err := baseCfg.(*koanf.KoanfConfig).Merge(ctx, overrideCfg.(*koanf.KoanfConfig))
//
// # Custom Delimiters
//
// By default, keys use dot notation (e.g., "server.port"). You can customize the delimiter:
//
//	cfg := koanf.NewKoanfConfig(ctx, koanf.WithDelimiter("/"))
//
//	// Now use / instead of .
//	val, err := cfg.GetValue(ctx, "server/port")
//
// # Flat Paths
//
// For structs with flat path tags:
//
//	type FlatConfig struct {
//	    ServerPort int    `koanf:"server.port"`
//	    ServerHost string `koanf:"server.host"`
//	}
//
//	var flat FlatConfig
//	err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "", &flat, koanf.WithFlatPaths(true))
//
// # Error Handling
//
// The package uses the errext package for structured error handling:
//
//	import "github.com/grinps/go-utils/errext"
//
//	val, err := cfg.GetValue(ctx, "nonexistent.key")
//	if err != nil {
//	    if errext.IsErrorCode(err, koanf.ErrKoanfMissingValue) {
//	        // Handle missing value
//	    }
//	}
//
// # Available Providers
//
// Koanf supports many providers (install separately):
//   - file: Load from files
//   - env: Load from environment variables
//   - confmap: Load from Go maps
//   - structs: Load from Go structs
//   - rawbytes: Load from raw bytes
//   - s3: Load from AWS S3
//   - vault: Load from HashiCorp Vault
//   - consul: Load from Consul
//   - etcd: Load from etcd
//   - And many more...
//
// # Available Parsers
//
// Koanf supports many parsers (install separately):
//   - json: JSON parser
//   - yaml: YAML parser
//   - toml: TOML parser
//   - hcl: HCL parser
//   - dotenv: .env file parser
//   - And many more...
//
// # Thread Safety
//
// KoanfConfig is safe for concurrent reads but not for concurrent writes.
// If you need to modify configuration from multiple goroutines, use external
// synchronization.
//
// # Integration with config Package
//
// KoanfConfig implements all standard config interfaces:
//
//	var cfg config.Config = koanf.NewKoanfConfig(ctx)
//
// This allows seamless integration with code that expects these interfaces.
//
// # Telemetry Support
//
// KoanfConfig implements config.TelemetryAware interface:
//   - ShouldInstrument() always returns true (telemetry enabled)
//   - GenerateTelemetryAttributes() returns attributes as-is
package koanf
