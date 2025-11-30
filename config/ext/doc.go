// Package ext provides extended configuration interfaces and utilities
// that build upon the base config package.
//
// # Overview
//
// This package defines additional interfaces for mutable configurations
// and struct unmarshalling capabilities, along with helper functions
// that work with any config.Config implementation. Functions extract
// config from context using config.ContextConfig.
//
// # Context-Based Functions
//
// The primary functions extract config from context:
//
//	// Store config in context
//	ctx = config.ContextWithConfig(ctx, cfg)
//
//	// Unmarshal extracts config from context
//	var server ServerConfig
//	err := ext.Unmarshal(ctx, "server", &server)
//
//	// SetValue extracts config from context
//	err := ext.SetValue(ctx, "server.port", 9090)
//
// Use the *WithConfig variants when you need to pass config explicitly:
//
//	err := ext.UnmarshalWithConfig(ctx, cfg, "server", &server)
//	err := ext.SetValueWithConfig(ctx, cfg, "server.port", 9090)
//
// # ConfigWrapper
//
// The ConfigWrapper wraps any config.Config and provides consistent access
// to MarshableConfig and MutableConfig capabilities:
//
//	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
//	wrapper := ext.NewConfigWrapper(cfg)
//
//	// Use Unmarshal consistently (uses mapstructure if native not available)
//	var server ServerConfig
//	err := wrapper.Unmarshal(ctx, "server", &server)
//
//	// SetValue works if wrapped config supports it
//	if wrapper.IsMutable() {
//	    wrapper.SetValue(ctx, "server.port", 9090)
//	}
//
// # Interfaces
//
// The package defines two main extension interfaces:
//
//   - MutableConfig: Defines SetValue for modifying configuration
//   - MarshableConfig: Defines Unmarshal for struct unmarshalling
//
// # Unmarshal Function
//
// The Unmarshal function extracts config from context and delegates to
// MarshableConfig if implemented:
//
//	type ServerConfig struct {
//	    Host string `config:"host"`
//	    Port int    `config:"port"`
//	}
//	ctx = config.ContextWithConfig(ctx, cfg)
//	var server ServerConfig
//	err := ext.Unmarshal(ctx, "server", &server)
//
// If the config implements MarshableConfig, its native Unmarshal method is used.
// Otherwise, returns an error. Use ConfigWrapper for mapstructure fallback.
//
// # Unmarshal Options
//
// The Unmarshal function accepts various options to customize behavior:
//
//	// Use JSON struct tags instead of "config"
//	err := ext.Unmarshal(ctx, "server", &server, ext.WithJSONTag())
//
//	// Enable strict mode (error on unused keys and unset fields)
//	err := ext.Unmarshal(ctx, "server", &server, ext.WithStrictMode())
//
//	// Custom decode hooks for type conversions
//	err := ext.Unmarshal(ctx, "server", &server, ext.WithDecodeHook(myHook))
//
// # Default Tag
//
// By default, the "config" struct tag is used for field mapping:
//
//	type Config struct {
//	    ServerHost string `config:"server_host"`
//	    ServerPort int    `config:"server_port"`
//	}
//
// Use WithTagName, WithJSONTag, WithYAMLTag, or WithMapstructureTag to change this.
//
// # Type Conversions
//
// The package supports automatic type conversions including:
//   - String to time.Duration (e.g., "30s" -> 30*time.Second)
//   - String to slice (comma-separated, e.g., "a,b,c" -> []string{"a","b","c"})
//   - Weak type conversions when enabled (e.g., "8080" -> 8080)
//
// # Error Handling
//
// The package uses the errext package for structured error handling:
//
//	err := ext.Unmarshal(ctx, cfg, "missing", &target)
//	if errors.Is(err, ext.ErrExtKeyNotFound) {
//	    // Handle missing key
//	}
package ext
