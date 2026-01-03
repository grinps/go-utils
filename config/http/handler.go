package confighttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/grinps/go-utils/config"
)

// Handler provides HTTP handlers for configuration management.
// It implements http.Handler and can be mounted on any standard Go HTTP server.
//
// Example:
//
//	cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
//	handler := confighttp.NewHandler(cfg)
//	http.ListenAndServe(":8080", handler)
type Handler struct {
	config config.Config

	// Options
	basePath    string
	readOnly    bool
	delimiter   string
	keyFilter   func(key string) bool
	onReload    func(ctx context.Context) error
	middleware  []func(http.Handler) http.Handler
	enableMeta  bool
	enableAdmin bool

	// Cached interface checks
	mutable   config.MutableConfig
	marshable config.MarshableConfig
	allKeys   config.AllKeysProvider
	deleter   config.Deleter
	allGetter config.AllGetter
}

// Ensure Handler implements http.Handler
var _ http.Handler = (*Handler)(nil)

// NewHandler creates a new Handler for the given config.
// Returns nil if cfg is nil.
//
// Example:
//
//	handler := confighttp.NewHandler(cfg,
//	    confighttp.WithBasePath("/api/config"),
//	    confighttp.WithReadOnly(true),
//	)
func NewHandler(cfg config.Config, opts ...Option) *Handler {
	if cfg == nil {
		return nil
	}

	h := &Handler{
		config:      cfg,
		delimiter:   ".",
		enableMeta:  true,
		enableAdmin: true,
	}

	// Cache interface checks
	if mc, ok := cfg.(config.MutableConfig); ok {
		h.mutable = mc
	}
	if mc, ok := cfg.(config.MarshableConfig); ok {
		h.marshable = mc
	}
	if ak, ok := cfg.(config.AllKeysProvider); ok {
		h.allKeys = ak
	}
	if d, ok := cfg.(config.Deleter); ok {
		h.deleter = d
	}
	if ag, ok := cfg.(config.AllGetter); ok {
		h.allGetter = ag
	}

	// Apply options
	for _, opt := range opts {
		opt(h)
	}

	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h == nil || h.config == nil {
		writeInternalError(w, "handler not initialized")
		return
	}

	// Apply middleware chain
	var handler http.Handler = http.HandlerFunc(h.route)
	for i := len(h.middleware) - 1; i >= 0; i-- {
		handler = h.middleware[i](handler)
	}

	handler.ServeHTTP(w, r)
}

// route dispatches requests to appropriate handlers based on path.
func (h *Handler) route(w http.ResponseWriter, r *http.Request) {
	path := h.extractPath(r)

	// Handle meta endpoints
	if h.enableMeta && strings.HasPrefix(path, "_meta/") {
		h.handleMeta(w, r, strings.TrimPrefix(path, "_meta/"))
		return
	}
	if h.enableMeta && path == "_meta" {
		h.handleMeta(w, r, "")
		return
	}

	// Handle admin endpoints
	if h.enableAdmin && strings.HasPrefix(path, "_admin/") {
		h.handleAdmin(w, r, strings.TrimPrefix(path, "_admin/"))
		return
	}
	if h.enableAdmin && path == "_admin" {
		h.handleAdmin(w, r, "")
		return
	}

	// Handle config endpoints
	h.handleConfig(w, r, path)
}

// extractPath extracts the config key path from the request URL.
func (h *Handler) extractPath(r *http.Request) string {
	path := r.URL.Path

	// Strip base path if configured
	if h.basePath != "" {
		path = strings.TrimPrefix(path, h.basePath)
	}

	// Clean path
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	return path
}

// pathToKey converts a URL path to a configuration key.
func (h *Handler) pathToKey(path string) string {
	if path == "" {
		return ""
	}
	// Replace / with delimiter
	return strings.ReplaceAll(path, "/", h.delimiter)
}

// isKeyAllowed checks if access to a key is allowed by the filter.
func (h *Handler) isKeyAllowed(key string) bool {
	if h.keyFilter == nil {
		return true
	}
	return h.keyFilter(key)
}

// handleConfig handles configuration value requests.
func (h *Handler) handleConfig(w http.ResponseWriter, r *http.Request, path string) {
	key := h.pathToKey(path)

	switch r.Method {
	case http.MethodGet:
		h.handleGet(w, r, key)
	case http.MethodHead:
		h.handleHead(w, r, key)
	case http.MethodPut:
		h.handlePut(w, r, key)
	case http.MethodDelete:
		h.handleDelete(w, r, key)
	default:
		allowed := []string{http.MethodGet, http.MethodHead}
		if h.mutable != nil && !h.readOnly {
			allowed = append(allowed, http.MethodPut)
		}
		if h.deleter != nil && !h.readOnly {
			allowed = append(allowed, http.MethodDelete)
		}
		writeMethodNotAllowed(w, r.Method, allowed)
	}
}

// handleGet handles GET requests for configuration values.
func (h *Handler) handleGet(w http.ResponseWriter, r *http.Request, key string) {
	ctx := r.Context()

	// Handle root request
	if key == "" {
		h.handleGetAll(w, r)
		return
	}

	// Check key filter
	if !h.isKeyAllowed(key) {
		writeForbidden(w, ErrorCodeKeyFiltered, "access to key is denied", map[string]any{"key": key})
		return
	}

	// Get value
	value, err := h.config.GetValue(ctx, key)
	if err != nil {
		writeNotFound(w, key)
		return
	}

	writeValue(w, key, value)
}

// handleGetAll handles GET requests for the entire configuration.
func (h *Handler) handleGetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// If we have AllGetter, use it
	if h.allGetter != nil {
		all := h.allGetter.All(ctx)
		if h.keyFilter != nil {
			all = h.filterMap(all, "")
		}
		writeValue(w, "", all)
		return
	}

	// Otherwise, try to get root as a value (may fail for some implementations)
	value, err := h.config.GetValue(ctx, "")
	if err == nil {
		if m, ok := value.(map[string]any); ok && h.keyFilter != nil {
			value = h.filterMap(m, "")
		}
		writeValue(w, "", value)
		return
	}

	// Fall back to returning empty object with message
	writeJSON(w, http.StatusOK, map[string]any{
		"message": "use specific key paths to access configuration",
	})
}

// filterMap recursively filters a map based on key filter.
func (h *Handler) filterMap(m map[string]any, prefix string) map[string]any {
	if h.keyFilter == nil {
		return m
	}

	result := make(map[string]any)
	for k, v := range m {
		fullKey := k
		if prefix != "" {
			fullKey = prefix + h.delimiter + k
		}

		if !h.isKeyAllowed(fullKey) {
			continue
		}

		if nested, ok := v.(map[string]any); ok {
			filtered := h.filterMap(nested, fullKey)
			if len(filtered) > 0 {
				result[k] = filtered
			}
		} else {
			result[k] = v
		}
	}
	return result
}

// handleHead handles HEAD requests to check key existence.
func (h *Handler) handleHead(w http.ResponseWriter, r *http.Request, key string) {
	ctx := r.Context()

	if key == "" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check key filter
	if !h.isKeyAllowed(key) {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Check if key exists
	_, err := h.config.GetValue(ctx, key)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// handlePut handles PUT requests to set configuration values.
func (h *Handler) handlePut(w http.ResponseWriter, r *http.Request, key string) {
	ctx := r.Context()

	// Check if key is provided
	if key == "" {
		writeBadRequest(w, ErrorCodeInvalidKey, "key is required for PUT", nil)
		return
	}

	// Check read-only mode
	if h.readOnly {
		writeForbidden(w, ErrorCodeReadOnly, "handler is in read-only mode", nil)
		return
	}

	// Check if config is mutable
	if h.mutable == nil {
		writeNotImplemented(w, ErrorCodeNotMutable, "configuration does not support mutations")
		return
	}

	// Check key filter
	if !h.isKeyAllowed(key) {
		writeForbidden(w, ErrorCodeKeyFiltered, "access to key is denied", map[string]any{"key": key})
		return
	}

	// Parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeBadRequest(w, ErrorCodeInvalidBody, "failed to read request body", map[string]any{"error": err.Error()})
		return
	}

	var req SetValueRequest
	if err := json.Unmarshal(body, &req); err != nil {
		writeBadRequest(w, ErrorCodeInvalidBody, "invalid JSON body", map[string]any{"error": err.Error()})
		return
	}

	// Set value
	if err := h.mutable.SetValue(ctx, key, req.Value); err != nil {
		writeError(w, http.StatusInternalServerError, ErrorCodeSetValueFailed, "failed to set value", map[string]any{
			"key":   key,
			"error": err.Error(),
		})
		return
	}

	// Check if key existed before (for 200 vs 201)
	writeJSON(w, http.StatusOK, ValueResponse{
		Key:   key,
		Value: req.Value,
	})
}

// handleDelete handles DELETE requests to remove configuration keys.
func (h *Handler) handleDelete(w http.ResponseWriter, r *http.Request, key string) {
	// Check if key is provided
	if key == "" {
		writeBadRequest(w, ErrorCodeInvalidKey, "key is required for DELETE", nil)
		return
	}

	// Check read-only mode
	if h.readOnly {
		writeForbidden(w, ErrorCodeReadOnly, "handler is in read-only mode", nil)
		return
	}

	// Check if config supports delete
	if h.deleter == nil {
		writeNotImplemented(w, ErrorCodeDeleteNotSupported, "configuration does not support delete")
		return
	}

	// Check key filter
	if !h.isKeyAllowed(key) {
		writeForbidden(w, ErrorCodeKeyFiltered, "access to key is denied", map[string]any{"key": key})
		return
	}

	// Delete key
	if err := h.deleter.Delete(key); err != nil {
		// Check if it's a not found error
		if strings.Contains(err.Error(), "not found") {
			writeNotFound(w, key)
			return
		}
		writeError(w, http.StatusInternalServerError, ErrorCodeDeleteFailed, "failed to delete key", map[string]any{
			"key":   key,
			"error": err.Error(),
		})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleMeta handles /_meta endpoint requests.
func (h *Handler) handleMeta(w http.ResponseWriter, r *http.Request, subpath string) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, r.Method, []string{http.MethodGet})
		return
	}

	switch {
	case subpath == "" || subpath == "info":
		h.handleMetaInfo(w, r)
	case subpath == "keys":
		h.handleMetaKeys(w, r, "")
	case strings.HasPrefix(subpath, "keys/"):
		prefix := strings.TrimPrefix(subpath, "keys/")
		h.handleMetaKeys(w, r, prefix)
	default:
		writeNotFound(w, "_meta/"+subpath)
	}
}

// handleMetaInfo handles GET /_meta/info requests.
func (h *Handler) handleMetaInfo(w http.ResponseWriter, r *http.Request) {
	info := InfoResponse{
		Provider:    string(h.config.Name()),
		Mutable:     h.mutable != nil,
		Marshable:   h.marshable != nil,
		HasKeys:     h.allKeys != nil,
		HasDelete:   h.deleter != nil,
		HasReload:   h.onReload != nil,
		ReadOnly:    h.readOnly,
		BasePath:    h.basePath,
		KeyFiltered: h.keyFilter != nil,
	}
	writeJSON(w, http.StatusOK, info)
}

// handleMetaKeys handles GET /_meta/keys requests.
func (h *Handler) handleMetaKeys(w http.ResponseWriter, r *http.Request, prefix string) {
	if h.allKeys == nil {
		writeNotImplemented(w, ErrorCodeKeyNotFound, "configuration does not support listing keys")
		return
	}

	// Convert URL prefix to key prefix
	keyPrefix := h.pathToKey(prefix)

	keys := h.allKeys.Keys(keyPrefix)

	// Filter keys if filter is configured
	if h.keyFilter != nil {
		filtered := make([]string, 0, len(keys))
		for _, key := range keys {
			if h.isKeyAllowed(key) {
				filtered = append(filtered, key)
			}
		}
		keys = filtered
	}

	if keys == nil {
		keys = []string{}
	}

	writeKeys(w, keys, keyPrefix)
}

// handleAdmin handles /_admin endpoint requests.
func (h *Handler) handleAdmin(w http.ResponseWriter, r *http.Request, subpath string) {
	switch {
	case subpath == "reload":
		h.handleAdminReload(w, r)
	default:
		writeNotFound(w, "_admin/"+subpath)
	}
}

// handleAdminReload handles POST /_admin/reload requests.
func (h *Handler) handleAdminReload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, r.Method, []string{http.MethodPost})
		return
	}

	if h.onReload == nil {
		writeNotImplemented(w, ErrorCodeReloadNotConfig, "reload handler not configured")
		return
	}

	if err := h.onReload(r.Context()); err != nil {
		writeError(w, http.StatusInternalServerError, ErrorCodeReloadFailed, "reload failed", map[string]any{
			"error": err.Error(),
		})
		return
	}

	writeSuccess(w, http.StatusOK, "configuration reloaded")
}

// IsMutable returns true if the underlying config supports SetValue.
func (h *Handler) IsMutable() bool {
	return h != nil && h.mutable != nil
}

// IsMarshable returns true if the underlying config supports Unmarshal.
func (h *Handler) IsMarshable() bool {
	return h != nil && h.marshable != nil
}

// Config returns the underlying config.Config.
func (h *Handler) Config() config.Config {
	if h == nil {
		return nil
	}
	return h.config
}
