// Package config provides a flexible, context-aware configuration management library
// for Go applications with type-safe value retrieval and pointer-based assignment.
//
// # Overview
//
// The config package supports nested configuration maps accessed via dot-notation keys,
// type-safe retrieval with compile-time guarantees, and structured error handling via
// the errext package.
//
// # Basic Usage
//
// Initialize a configuration and inject it into context:
//
//	ctx := context.Background()
//	data := map[string]any{
//	    "server": map[string]any{
//	        "port": 8080,
//	        "host": "localhost",
//	    },
//	}
//	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
//	ctx = config.ContextWithConfig(ctx, cfg)
//
// # Retrieving Values
//
// The package provides a single primary function GetValueE that uses pointer-based
// assignment for type safety:
//
//	// Direct config method
//	var port int
//	err := cfg.GetValue(ctx, "server.port", &port)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Package-level function (recommended)
//	var host string
//	err = config.GetValueE(ctx, "server.host", &host)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Default Value Pattern
//
// Variables retain their initial values when keys are not found, enabling a
// default value pattern:
//
//	timeout := 30 // default value
//	err := config.GetValueE(ctx, "server.timeout", &timeout)
//	if err != nil {
//	    // timeout still contains 30 (default)
//	    log.Printf("Using default timeout: %d", timeout)
//	}
//
// # Nested Configurations
//
// Access nested values using dot notation or retrieve sub-configurations:
//
//	// Dot notation
//	var dbHost string
//	cfg.GetValue(ctx, "database.host", &dbHost)
//
//	// Sub-configuration
//	dbCfg, err := cfg.GetConfig(ctx, "database")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	var host string
//	dbCfg.GetValue(ctx, "host", &host)
//
// # Error Handling
//
// The package uses the errext package for structured error handling with error codes:
//
//	var port int
//	err := cfg.GetValue(ctx, "missing.key", &port)
//	if err != nil {
//	    // Check specific error types
//	    if errors.Is(err, config.ErrConfigMissingValue) {
//	        // Handle missing value
//	    }
//	}
//
// # Error Codes
//
// The following error codes are defined:
//   - ErrConfigCodeUnknown: Unknown error
//   - ErrConfigMissingValue: Value not found
//   - ErrConfigEmptyKey: Empty key provided
//   - ErrConfigInvalidKey: Invalid key format
//   - ErrConfigNilConfig: Nil config encountered
//   - ErrConfigInvalidValueType: Type mismatch
//   - ErrConfigInvalidValue: Invalid value or conversion error
//   - ErrConfigNilReturnValue: Nil return value pointer provided
//
// # Type Safety
//
// The package ensures type safety at runtime using reflection. Type mismatches
// result in clear error messages:
//
//	var wrongType string
//	err := cfg.GetValue(ctx, "server.port", &wrongType) // port is int
//	// Returns: ErrConfigInvalidValueType with details
//
// # Custom Delimiters
//
// Change the key delimiter from the default "." to any string:
//
//	cfg := config.NewSimpleConfig(ctx,
//	    config.WithConfigurationMap(data),
//	    config.WithDelimiter("/"))
//	var port int
//	cfg.GetValue(ctx, "server/port", &port)
//
// # Context Integration
//
// The package is designed to work seamlessly with context.Context:
//
//	// Store config in context
//	ctx = config.ContextWithConfig(ctx, cfg)
//
//	// Retrieve config from context
//	cfg := config.ContextConfig(ctx, true) // true = use default if not found
//
//	// Use package-level functions (automatically use context config)
//	var value string
//	config.GetValueE(ctx, "key", &value)
//
// # Testing
//
// The SimpleConfig implementation is ideal for testing:
//
//	func TestMyFunction(t *testing.T) {
//	    ctx := context.Background()
//	    testData := map[string]any{
//	        "api": map[string]any{
//	            "endpoint": "https://test.example.com",
//	            "timeout": 5,
//	        },
//	    }
//	    cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(testData))
//	    ctx = config.ContextWithConfig(ctx, cfg)
//
//	    // Test your function with the test config
//	    result := MyFunction(ctx)
//	    // assertions...
//	}
//
// # Performance
//
// The package uses reflection for type checking, which has a small performance cost.
// For high-performance scenarios, consider caching retrieved values or using the
// Config.GetValue method directly on a stored config instance.
//
// # Thread Safety
//
// SimpleConfig is not thread-safe for concurrent reads and writes. If you need to
// modify configuration at runtime from multiple goroutines, add appropriate
// synchronization or use an immutable configuration pattern.
package config
