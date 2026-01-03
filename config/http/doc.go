// Package confighttp provides HTTP handlers for configuration management.
//
// This package enables exposing configuration values through a RESTful HTTP
// interface, allowing runtime configuration inspection and modification.
// It integrates seamlessly with the github.com/grinps/go-utils/config package
// and supports all Config implementations (SimpleConfig, KoanfConfig, etc.).
//
// # Overview
//
// The package provides a ConfigHandler that implements http.Handler and can be
// mounted on any standard Go HTTP server or router. It follows REST conventions
// with hierarchical resource URLs that map naturally to configuration key paths.
//
// # Quick Start
//
//	import (
//	    "net/http"
//	    "github.com/grinps/go-utils/config"
//	    confighttp "github.com/grinps/go-utils/config/http"
//	)
//
//	func main() {
//	    cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
//	    handler := confighttp.NewHandler(cfg)
//	    http.ListenAndServe(":8080", handler)
//	}
//
// # API Endpoints
//
// The handler exposes the following REST endpoints:
//
//	GET    /                     Get entire configuration
//	GET    /{key...}             Get value at key path (e.g., /server/port)
//	HEAD   /{key...}             Check if key exists
//	PUT    /{key...}             Set value at key path (requires MutableConfig)
//	DELETE /{key...}             Delete key (requires MutableConfig with Delete)
//	GET    /_meta/keys           List all configuration keys
//	GET    /_meta/keys/{prefix}  List keys with prefix
//	GET    /_meta/info           Get handler information
//	POST   /_admin/reload        Reload configuration (if reload handler configured)
//
// # URL to Key Mapping
//
// URL paths are converted to dot-notation configuration keys:
//
//	/server/port     → server.port
//	/database/host   → database.host
//	/app/settings/debug → app.settings.debug
//
// # Response Format
//
// All responses are JSON formatted:
//
//	// Success response for GET /server/port
//	{
//	    "key": "server.port",
//	    "value": 8080
//	}
//
//	// Success response for GET /server
//	{
//	    "key": "server",
//	    "value": {
//	        "port": 8080,
//	        "host": "localhost"
//	    }
//	}
//
//	// Error response
//	{
//	    "error": {
//	        "code": "CONFIG_MISSING_VALUE",
//	        "message": "key not found",
//	        "details": {"key": "server.missing"}
//	    }
//	}
//
// # Mounting on Existing Servers
//
// The handler can be mounted at any path on an existing server:
//
//	mux := http.NewServeMux()
//	mux.HandleFunc("/api/users", usersHandler)
//
//	// Mount config handler at /api/config
//	configHandler := confighttp.NewHandler(cfg,
//	    confighttp.WithBasePath("/api/config"),
//	)
//	mux.Handle("/api/config/", configHandler)
//
// # Security
//
// For production use, consider these security options:
//
//	handler := confighttp.NewHandler(cfg,
//	    // Read-only mode (disable PUT/DELETE)
//	    confighttp.WithReadOnly(true),
//
//	    // Filter sensitive keys
//	    confighttp.WithKeyFilter(func(key string) bool {
//	        return !strings.HasPrefix(key, "secrets.")
//	    }),
//
//	    // Add authentication middleware
//	    confighttp.WithMiddleware(authMiddleware),
//	)
//
// # Thread Safety
//
// The ConfigHandler is safe for concurrent use. However, thread safety of
// configuration modifications depends on the underlying Config implementation.
// Most implementations are safe for concurrent reads but not concurrent writes.
//
// # Interface Support
//
// The handler automatically detects and uses extended interfaces from the config package:
//
//   - config.Config: Required, provides GetValue and GetConfig
//   - config.MutableConfig: Optional, enables PUT operations via SetValue
//   - config.AllKeysProvider: Optional, enables /_meta/keys endpoint
//   - config.Deleter: Optional, enables DELETE operations
//   - config.AllGetter: Optional, enables full config retrieval at root
//
// SimpleConfig implements all optional interfaces. Use type assertions to check support:
//
//	if keysProvider, ok := cfg.(config.AllKeysProvider); ok {
//	    keys := keysProvider.Keys("server")
//	}
//
//	if deleter, ok := cfg.(config.Deleter); ok {
//	    err := deleter.Delete("server.debug")
//	}
package confighttp
