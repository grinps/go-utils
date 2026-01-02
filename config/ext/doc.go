// Package ext provides extended configuration utilities that build upon
// the base config package.
//
// # Overview
//
// This package provides ConfigWrapper which wraps any config.Config and adds
// consistent MarshableConfig and MutableConfig capabilities with mapstructure
// fallback support. ConfigWrapper also implements config.TelemetryAware for
// telemetry integration.
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
// # Unmarshal Options
//
// Customize unmarshalling behavior with functional options:
//
//	// Use JSON tags instead of "config"
//	err := wrapper.Unmarshal(ctx, "server", &server, ext.WithJSONTag())
//
//	// Enable strict mode
//	err := wrapper.Unmarshal(ctx, "server", &server, ext.WithStrictMode())
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
//	err := wrapper.Unmarshal(ctx, "missing", &target)
//	if errors.Is(err, ext.ErrExtKeyNotFound) {
//	    // Handle missing key
//	}
//
// # Telemetry Support
//
// ConfigWrapper implements config.TelemetryAware interface:
//   - ShouldInstrument() always returns true (telemetry enabled)
//   - GenerateTelemetryAttributes() returns attributes as-is
package ext
