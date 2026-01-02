// Package config provides a flexible, context-aware configuration management library
// for Go applications with type-safe value retrieval and pointer-based assignment.
//
// # Overview
//
// The config package supports nested configuration maps accessed via dot-notation keys,
// type-safe retrieval with compile-time guarantees, structured error handling via
// the errext package, and built-in telemetry support.
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
// The package provides two approaches for retrieving values:
//
//	// Direct config method - returns (any, error)
//	val, err := cfg.GetValue(ctx, "server.port")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	port := val.(int) // Type assertion required
//
//	// Package-level function with type safety (recommended)
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
//	val, err := cfg.GetValue(ctx, "database.host")
//	dbHost := val.(string)
//
//	// Sub-configuration
//	dbCfg, err := cfg.GetConfig(ctx, "database")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	val, err = dbCfg.GetValue(ctx, "host")
//	host := val.(string)
//
// # Error Handling
//
// The package uses the errext package for structured error handling with error codes:
//
//	val, err := cfg.GetValue(ctx, "missing.key")
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
// The GetValue method returns any and requires type assertion. For type-safe
// retrieval with compile-time guarantees, use GetValueE:
//
//	// Type assertion required
//	val, _ := cfg.GetValue(ctx, "server.port")
//	port := val.(int)
//
//	// Type-safe with GetValueE
//	var port int
//	err := config.GetValueE(ctx, "server.port", &port)
//	// Returns: ErrConfigInvalidValueType if types don't match
//
// # Custom Delimiters
//
// Change the key delimiter from the default "." to any string:
//
//	cfg := config.NewSimpleConfig(ctx,
//	    config.WithConfigurationMap(data),
//	    config.WithDelimiter("/"))
//	val, err := cfg.GetValue(ctx, "server/port")
//	port := val.(int)
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
// # Setting Default Configuration
//
// Set a custom default configuration that is used when no config is in context:
//
//	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
//	config.SetAsDefault(cfg)
//
//	// Now ContextConfig(ctx, true) returns this config when no config in context
//	defaultCfg := config.Default()
//
// # Explicit Config Functions
//
// For cases where you want to use a specific config without context:
//
//	// GetValueWithConfig - type-safe retrieval with explicit config
//	var port int
//	err := config.GetValueWithConfig(ctx, cfg, "server.port", &port)
//
//	// GetConfigWithConfig - get nested config with explicit config
//	serverCfg, err := config.GetConfigWithConfig(ctx, cfg, "server")
//
//	// GetConfig - get nested config from context
//	serverCfg, err := config.GetConfig(ctx, "server")
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
//
// # Telemetry
//
// The package includes built-in telemetry support for tracing and metrics. All
// configuration operations automatically capture spans and record metrics when
// telemetry is enabled.
//
// Control telemetry globally:
//
//	config.SetTelemetryEnabled(false) // Disable telemetry
//	config.SetTelemetryEnabled(true)  // Re-enable telemetry
//	if config.IsTelemetryEnabled() {
//	    // telemetry is active
//	}
//
// Custom Config implementations can implement the TelemetryAware interface for
// fine-grained control over instrumentation and custom attributes.
//
// # Config Interface
//
// All Config implementations must provide a Name() method that returns a ProviderName.
// This identifier is used for telemetry attributes and debugging:
//
//	type Config interface {
//	    Name() ProviderName
//	    GetValue(ctx context.Context, key string) (any, error)
//	    GetConfig(ctx context.Context, key string) (Config, error)
//	}
//
// The ProviderName type is a string alias that identifies the config implementation
// (e.g., "SimpleConfig", "KoanfConfig").
package config
