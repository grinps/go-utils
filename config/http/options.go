package confighttp

import (
	"context"
	"net/http"
)

// Option is a functional option for configuring a Handler.
type Option func(*Handler)

// WithBasePath sets the base path for the handler.
// The base path is stripped from incoming request URLs before processing.
// Default is empty (handler serves from root).
//
// Example:
//
//	handler := confighttp.NewHandler(cfg, confighttp.WithBasePath("/api/config"))
//	// Request to /api/config/server/port will look up "server.port"
func WithBasePath(path string) Option {
	return func(h *Handler) {
		h.basePath = path
	}
}

// WithReadOnly sets the handler to read-only mode.
// When enabled, PUT and DELETE operations return 405 Method Not Allowed.
// Default is false.
//
// Example:
//
//	handler := confighttp.NewHandler(cfg, confighttp.WithReadOnly(true))
func WithReadOnly(readOnly bool) Option {
	return func(h *Handler) {
		h.readOnly = readOnly
	}
}

// WithKeyFilter sets a filter function for configuration keys.
// The filter is called for each key access. If it returns false, the key
// is treated as not found. This is useful for hiding sensitive configuration.
// Default is nil (all keys accessible).
//
// Example:
//
//	handler := confighttp.NewHandler(cfg, confighttp.WithKeyFilter(func(key string) bool {
//	    return !strings.HasPrefix(key, "secrets.")
//	}))
func WithKeyFilter(filter func(key string) bool) Option {
	return func(h *Handler) {
		h.keyFilter = filter
	}
}

// WithReloadHandler sets a function to be called when POST /_admin/reload is invoked.
// This allows triggering configuration reloads from external sources.
// Default is nil (reload endpoint returns 501 Not Implemented).
//
// Example:
//
//	handler := confighttp.NewHandler(cfg, confighttp.WithReloadHandler(func(ctx context.Context) error {
//	    return cfg.Load(ctx, file.Provider("config.json"), json.Parser())
//	}))
func WithReloadHandler(fn func(ctx context.Context) error) Option {
	return func(h *Handler) {
		h.onReload = fn
	}
}

// WithMiddleware adds middleware to the handler chain.
// Middleware is applied in the order provided.
// This is useful for adding authentication, logging, or other cross-cutting concerns.
//
// Example:
//
//	handler := confighttp.NewHandler(cfg, confighttp.WithMiddleware(
//	    loggingMiddleware,
//	    authMiddleware,
//	))
func WithMiddleware(middleware ...func(http.Handler) http.Handler) Option {
	return func(h *Handler) {
		h.middleware = append(h.middleware, middleware...)
	}
}

// WithDelimiter sets the key delimiter used for path-to-key conversion.
// Default is "." (e.g., /server/port becomes server.port).
//
// Example:
//
//	handler := confighttp.NewHandler(cfg, confighttp.WithDelimiter("/"))
//	// Request to /server/port will look up "server/port" (literal key)
func WithDelimiter(delimiter string) Option {
	return func(h *Handler) {
		h.delimiter = delimiter
	}
}

// WithMetaEndpoints enables or disables the /_meta endpoints.
// Default is true.
//
// Example:
//
//	handler := confighttp.NewHandler(cfg, confighttp.WithMetaEndpoints(false))
func WithMetaEndpoints(enabled bool) Option {
	return func(h *Handler) {
		h.enableMeta = enabled
	}
}

// WithAdminEndpoints enables or disables the /_admin endpoints.
// Default is true.
//
// Example:
//
//	handler := confighttp.NewHandler(cfg, confighttp.WithAdminEndpoints(false))
func WithAdminEndpoints(enabled bool) Option {
	return func(h *Handler) {
		h.enableAdmin = enabled
	}
}
