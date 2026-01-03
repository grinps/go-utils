package confighttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/grinps/go-utils/config"
)

// mockConfig implements config.Config for testing
type mockConfig struct {
	name      config.ProviderName
	data      map[string]any
	getErr    error
	setErr    error
	deleteErr error
}

func (m *mockConfig) Name() config.ProviderName {
	if m.name == "" {
		return "MockConfig"
	}
	return m.name
}

func (m *mockConfig) GetValue(ctx context.Context, key string) (any, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if key == "" {
		return m.data, nil
	}
	parts := strings.Split(key, ".")
	var current any = m.data
	for _, part := range parts {
		if currentMap, ok := current.(map[string]any); ok {
			if val, found := currentMap[part]; found {
				current = val
			} else {
				return nil, errors.New("key not found")
			}
		} else {
			return nil, errors.New("not a map")
		}
	}
	return current, nil
}

func (m *mockConfig) GetConfig(ctx context.Context, key string) (config.Config, error) {
	val, err := m.GetValue(ctx, key)
	if err != nil {
		return nil, err
	}
	if mapVal, ok := val.(map[string]any); ok {
		return &mockConfig{data: mapVal}, nil
	}
	return nil, errors.New("not a map")
}

// mockMutableConfig adds SetValue support
type mockMutableConfig struct {
	mockConfig
}

func (m *mockMutableConfig) SetValue(ctx context.Context, key string, value any) error {
	if m.setErr != nil {
		return m.setErr
	}
	if m.data == nil {
		m.data = make(map[string]any)
	}
	parts := strings.Split(key, ".")
	current := m.data
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return nil
		}
		if next, ok := current[part].(map[string]any); ok {
			current = next
		} else {
			newMap := make(map[string]any)
			current[part] = newMap
			current = newMap
		}
	}
	return nil
}

// mockFullConfig implements all optional interfaces
type mockFullConfig struct {
	mockMutableConfig
	keys    []string
	keysErr error
	allData map[string]any
}

func (m *mockFullConfig) Keys(prefix string) []string {
	if prefix == "" {
		return m.keys
	}
	var filtered []string
	for _, k := range m.keys {
		if strings.HasPrefix(k, prefix) {
			filtered = append(filtered, k)
		}
	}
	return filtered
}

func (m *mockFullConfig) Delete(key string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	// Simple delete for testing
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		if _, found := m.data[key]; !found {
			return errors.New("key not found")
		}
		delete(m.data, key)
		return nil
	}
	return errors.New("nested delete not implemented in mock")
}

func (m *mockFullConfig) All(ctx context.Context) map[string]any {
	if m.allData != nil {
		return m.allData
	}
	return m.data
}

// Helper functions for tests
func newTestHandler(data map[string]any, opts ...Option) *Handler {
	cfg := &mockConfig{data: data}
	return NewHandler(cfg, opts...)
}

func newMutableTestHandler(data map[string]any, opts ...Option) *Handler {
	cfg := &mockMutableConfig{mockConfig: mockConfig{data: data}}
	return NewHandler(cfg, opts...)
}

func newFullTestHandler(data map[string]any, keys []string, opts ...Option) *Handler {
	cfg := &mockFullConfig{
		mockMutableConfig: mockMutableConfig{mockConfig: mockConfig{data: data}},
		keys:              keys,
	}
	return NewHandler(cfg, opts...)
}

func doRequest(handler http.Handler, method, path string, body any) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != nil {
		bodyBytes, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(bodyBytes)
	}
	req := httptest.NewRequest(method, path, bodyReader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func parseResponse[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	var result T
	if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v, body: %s", err, rec.Body.String())
	}
	return result
}

// Tests

func TestNewHandler(t *testing.T) {
	t.Run("nil config returns nil", func(t *testing.T) {
		h := NewHandler(nil)
		if h != nil {
			t.Error("expected nil handler for nil config")
		}
	})

	t.Run("valid config returns handler", func(t *testing.T) {
		cfg := &mockConfig{data: map[string]any{"key": "value"}}
		h := NewHandler(cfg)
		if h == nil {
			t.Error("expected non-nil handler")
		}
		if h.config != cfg {
			t.Error("config not set correctly")
		}
	})

	t.Run("options are applied", func(t *testing.T) {
		cfg := &mockConfig{data: map[string]any{}}
		h := NewHandler(cfg,
			WithBasePath("/api"),
			WithReadOnly(true),
			WithDelimiter("/"),
			WithMetaEndpoints(false),
			WithAdminEndpoints(false),
		)
		if h.basePath != "/api" {
			t.Errorf("expected basePath /api, got %s", h.basePath)
		}
		if !h.readOnly {
			t.Error("expected readOnly true")
		}
		if h.delimiter != "/" {
			t.Errorf("expected delimiter /, got %s", h.delimiter)
		}
		if h.enableMeta {
			t.Error("expected enableMeta false")
		}
		if h.enableAdmin {
			t.Error("expected enableAdmin false")
		}
	})

	t.Run("interface detection", func(t *testing.T) {
		// Test with basic config
		basicCfg := &mockConfig{data: map[string]any{}}
		h1 := NewHandler(basicCfg)
		if h1.mutable != nil {
			t.Error("expected mutable nil for basic config")
		}

		// Test with mutable config
		mutableCfg := &mockMutableConfig{mockConfig: mockConfig{data: map[string]any{}}}
		h2 := NewHandler(mutableCfg)
		if h2.mutable == nil {
			t.Error("expected mutable non-nil for mutable config")
		}

		// Test with full config
		fullCfg := &mockFullConfig{mockMutableConfig: mockMutableConfig{mockConfig: mockConfig{data: map[string]any{}}}}
		h3 := NewHandler(fullCfg)
		if h3.allKeys == nil {
			t.Error("expected allKeys non-nil for full config")
		}
		if h3.deleter == nil {
			t.Error("expected deleter non-nil for full config")
		}
		if h3.allGetter == nil {
			t.Error("expected allGetter non-nil for full config")
		}
	})
}

func TestHandler_ServeHTTP_NilHandler(t *testing.T) {
	var h *Handler
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", rec.Code)
	}
}

func TestHandler_GET_Value(t *testing.T) {
	data := map[string]any{
		"server": map[string]any{
			"port": 8080,
			"host": "localhost",
		},
		"app": map[string]any{
			"name": "test-app",
		},
	}

	t.Run("get nested value", func(t *testing.T) {
		h := newTestHandler(data)
		rec := doRequest(h, http.MethodGet, "/server/port", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		resp := parseResponse[ValueResponse](t, rec)
		if resp.Key != "server.port" {
			t.Errorf("expected key server.port, got %s", resp.Key)
		}
		if resp.Value.(float64) != 8080 {
			t.Errorf("expected value 8080, got %v", resp.Value)
		}
	})

	t.Run("get object value", func(t *testing.T) {
		h := newTestHandler(data)
		rec := doRequest(h, http.MethodGet, "/server", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		resp := parseResponse[ValueResponse](t, rec)
		if resp.Key != "server" {
			t.Errorf("expected key server, got %s", resp.Key)
		}
	})

	t.Run("get root with AllGetter", func(t *testing.T) {
		h := newFullTestHandler(data, nil)
		rec := doRequest(h, http.MethodGet, "/", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("get root without AllGetter", func(t *testing.T) {
		h := newTestHandler(data)
		rec := doRequest(h, http.MethodGet, "/", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("key not found", func(t *testing.T) {
		h := newTestHandler(data)
		rec := doRequest(h, http.MethodGet, "/nonexistent", nil)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}

		resp := parseResponse[ErrorResponse](t, rec)
		if resp.Error.Code != ErrorCodeKeyNotFound {
			t.Errorf("expected error code %s, got %s", ErrorCodeKeyNotFound, resp.Error.Code)
		}
	})

	t.Run("key filtered", func(t *testing.T) {
		h := newTestHandler(data, WithKeyFilter(func(key string) bool {
			return !strings.HasPrefix(key, "server")
		}))
		rec := doRequest(h, http.MethodGet, "/server/port", nil)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}

		resp := parseResponse[ErrorResponse](t, rec)
		if resp.Error.Code != ErrorCodeKeyFiltered {
			t.Errorf("expected error code %s, got %s", ErrorCodeKeyFiltered, resp.Error.Code)
		}
	})

	t.Run("with base path", func(t *testing.T) {
		h := newTestHandler(data, WithBasePath("/api/config"))
		rec := doRequest(h, http.MethodGet, "/api/config/server/port", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})
}

func TestHandler_HEAD(t *testing.T) {
	data := map[string]any{
		"key": "value",
	}

	t.Run("key exists", func(t *testing.T) {
		h := newTestHandler(data)
		rec := doRequest(h, http.MethodHead, "/key", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if rec.Body.Len() != 0 {
			t.Error("expected empty body for HEAD")
		}
	})

	t.Run("key not found", func(t *testing.T) {
		h := newTestHandler(data)
		rec := doRequest(h, http.MethodHead, "/nonexistent", nil)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("root exists", func(t *testing.T) {
		h := newTestHandler(data)
		rec := doRequest(h, http.MethodHead, "/", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("key filtered", func(t *testing.T) {
		h := newTestHandler(data, WithKeyFilter(func(key string) bool {
			return false
		}))
		rec := doRequest(h, http.MethodHead, "/key", nil)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})
}

func TestHandler_PUT(t *testing.T) {
	t.Run("set value success", func(t *testing.T) {
		h := newMutableTestHandler(map[string]any{})
		rec := doRequest(h, http.MethodPut, "/server/port", SetValueRequest{Value: 9090})

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", rec.Code, rec.Body.String())
		}

		resp := parseResponse[ValueResponse](t, rec)
		if resp.Key != "server.port" {
			t.Errorf("expected key server.port, got %s", resp.Key)
		}
	})

	t.Run("put without key", func(t *testing.T) {
		h := newMutableTestHandler(map[string]any{})
		rec := doRequest(h, http.MethodPut, "/", SetValueRequest{Value: "test"})

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("read-only mode", func(t *testing.T) {
		h := newMutableTestHandler(map[string]any{}, WithReadOnly(true))
		rec := doRequest(h, http.MethodPut, "/key", SetValueRequest{Value: "test"})

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}

		resp := parseResponse[ErrorResponse](t, rec)
		if resp.Error.Code != ErrorCodeReadOnly {
			t.Errorf("expected error code %s, got %s", ErrorCodeReadOnly, resp.Error.Code)
		}
	})

	t.Run("not mutable", func(t *testing.T) {
		h := newTestHandler(map[string]any{})
		rec := doRequest(h, http.MethodPut, "/key", SetValueRequest{Value: "test"})

		if rec.Code != http.StatusNotImplemented {
			t.Errorf("expected status 501, got %d", rec.Code)
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		h := newMutableTestHandler(map[string]any{})
		req := httptest.NewRequest(http.MethodPut, "/key", strings.NewReader("invalid json"))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("key filtered", func(t *testing.T) {
		h := newMutableTestHandler(map[string]any{}, WithKeyFilter(func(key string) bool {
			return false
		}))
		rec := doRequest(h, http.MethodPut, "/key", SetValueRequest{Value: "test"})

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("set value error", func(t *testing.T) {
		cfg := &mockMutableConfig{mockConfig: mockConfig{data: map[string]any{}}}
		cfg.setErr = errors.New("set failed")
		h := NewHandler(cfg)
		rec := doRequest(h, http.MethodPut, "/key", SetValueRequest{Value: "test"})

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestHandler_DELETE(t *testing.T) {
	t.Run("delete success", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{"key": "value"}, nil)
		rec := doRequest(h, http.MethodDelete, "/key", nil)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d, body: %s", rec.Code, rec.Body.String())
		}
	})

	t.Run("delete without key", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{}, nil)
		rec := doRequest(h, http.MethodDelete, "/", nil)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}
	})

	t.Run("read-only mode", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{"key": "value"}, nil, WithReadOnly(true))
		rec := doRequest(h, http.MethodDelete, "/key", nil)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("delete not supported", func(t *testing.T) {
		h := newMutableTestHandler(map[string]any{"key": "value"})
		rec := doRequest(h, http.MethodDelete, "/key", nil)

		if rec.Code != http.StatusNotImplemented {
			t.Errorf("expected status 501, got %d", rec.Code)
		}
	})

	t.Run("key not found", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{}, nil)
		rec := doRequest(h, http.MethodDelete, "/nonexistent", nil)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("key filtered", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{"key": "value"}, nil, WithKeyFilter(func(key string) bool {
			return false
		}))
		rec := doRequest(h, http.MethodDelete, "/key", nil)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", rec.Code)
		}
	})

	t.Run("delete error", func(t *testing.T) {
		cfg := &mockFullConfig{mockMutableConfig: mockMutableConfig{mockConfig: mockConfig{data: map[string]any{"key": "value"}}}}
		cfg.deleteErr = errors.New("some error")
		h := NewHandler(cfg)
		rec := doRequest(h, http.MethodDelete, "/key", nil)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})
}

func TestHandler_Meta_Info(t *testing.T) {
	t.Run("get info", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{}, nil, WithBasePath("/api"))
		rec := doRequest(h, http.MethodGet, "/_meta/info", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		resp := parseResponse[InfoResponse](t, rec)
		if resp.Provider != "MockConfig" {
			t.Errorf("expected provider MockConfig, got %s", resp.Provider)
		}
		if !resp.Mutable {
			t.Error("expected Mutable true")
		}
		if !resp.HasKeys {
			t.Error("expected HasKeys true")
		}
		if !resp.HasDelete {
			t.Error("expected HasDelete true")
		}
		if resp.BasePath != "/api" {
			t.Errorf("expected BasePath /api, got %s", resp.BasePath)
		}
	})

	t.Run("get info via _meta", func(t *testing.T) {
		h := newTestHandler(map[string]any{})
		rec := doRequest(h, http.MethodGet, "/_meta", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		h := newTestHandler(map[string]any{})
		rec := doRequest(h, http.MethodPost, "/_meta/info", nil)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", rec.Code)
		}
	})

	t.Run("unknown meta endpoint", func(t *testing.T) {
		h := newTestHandler(map[string]any{})
		rec := doRequest(h, http.MethodGet, "/_meta/unknown", nil)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("meta disabled", func(t *testing.T) {
		h := newTestHandler(map[string]any{}, WithMetaEndpoints(false))
		rec := doRequest(h, http.MethodGet, "/_meta/info", nil)

		// When meta is disabled, /_meta/info is treated as a config key
		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})
}

func TestHandler_Meta_Keys(t *testing.T) {
	keys := []string{"server.port", "server.host", "app.name"}

	t.Run("get all keys", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{}, keys)
		rec := doRequest(h, http.MethodGet, "/_meta/keys", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		resp := parseResponse[KeysResponse](t, rec)
		if resp.Count != 3 {
			t.Errorf("expected count 3, got %d", resp.Count)
		}
	})

	t.Run("get keys with prefix", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{}, keys)
		rec := doRequest(h, http.MethodGet, "/_meta/keys/server", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		resp := parseResponse[KeysResponse](t, rec)
		if resp.Count != 2 {
			t.Errorf("expected count 2, got %d", resp.Count)
		}
		if resp.Prefix != "server" {
			t.Errorf("expected prefix server, got %s", resp.Prefix)
		}
	})

	t.Run("keys not supported", func(t *testing.T) {
		h := newTestHandler(map[string]any{})
		rec := doRequest(h, http.MethodGet, "/_meta/keys", nil)

		if rec.Code != http.StatusNotImplemented {
			t.Errorf("expected status 501, got %d", rec.Code)
		}
	})

	t.Run("keys filtered", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{}, keys, WithKeyFilter(func(key string) bool {
			return strings.HasPrefix(key, "app")
		}))
		rec := doRequest(h, http.MethodGet, "/_meta/keys", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		resp := parseResponse[KeysResponse](t, rec)
		if resp.Count != 1 {
			t.Errorf("expected count 1 (only app.name), got %d", resp.Count)
		}
	})
}

func TestHandler_Admin_Reload(t *testing.T) {
	t.Run("reload success", func(t *testing.T) {
		reloadCalled := false
		h := newTestHandler(map[string]any{}, WithReloadHandler(func(ctx context.Context) error {
			reloadCalled = true
			return nil
		}))
		rec := doRequest(h, http.MethodPost, "/_admin/reload", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if !reloadCalled {
			t.Error("expected reload to be called")
		}

		resp := parseResponse[SuccessResponse](t, rec)
		if !resp.Success {
			t.Error("expected success true")
		}
	})

	t.Run("reload not configured", func(t *testing.T) {
		h := newTestHandler(map[string]any{})
		rec := doRequest(h, http.MethodPost, "/_admin/reload", nil)

		if rec.Code != http.StatusNotImplemented {
			t.Errorf("expected status 501, got %d", rec.Code)
		}
	})

	t.Run("reload error", func(t *testing.T) {
		h := newTestHandler(map[string]any{}, WithReloadHandler(func(ctx context.Context) error {
			return errors.New("reload failed")
		}))
		rec := doRequest(h, http.MethodPost, "/_admin/reload", nil)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", rec.Code)
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		h := newTestHandler(map[string]any{}, WithReloadHandler(func(ctx context.Context) error {
			return nil
		}))
		rec := doRequest(h, http.MethodGet, "/_admin/reload", nil)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", rec.Code)
		}
	})

	t.Run("unknown admin endpoint", func(t *testing.T) {
		h := newTestHandler(map[string]any{})
		rec := doRequest(h, http.MethodGet, "/_admin/unknown", nil)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", rec.Code)
		}
	})

	t.Run("admin disabled", func(t *testing.T) {
		h := newTestHandler(map[string]any{}, WithAdminEndpoints(false), WithReloadHandler(func(ctx context.Context) error {
			return nil
		}))
		rec := doRequest(h, http.MethodPost, "/_admin/reload", nil)

		// When admin is disabled, /_admin/reload is treated as a config key
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405 (PUT not allowed on basic config), got %d", rec.Code)
		}
	})
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	t.Run("basic config - only GET/HEAD allowed", func(t *testing.T) {
		h := newTestHandler(map[string]any{"key": "value"})
		rec := doRequest(h, http.MethodPost, "/key", nil)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", rec.Code)
		}

		// Check Allow header
		allow := rec.Header().Get("Allow")
		if !strings.Contains(allow, "GET") || !strings.Contains(allow, "HEAD") {
			t.Errorf("expected Allow header to contain GET, HEAD, got %s", allow)
		}
	})

	t.Run("mutable config - GET/HEAD/PUT allowed", func(t *testing.T) {
		h := newMutableTestHandler(map[string]any{"key": "value"})
		rec := doRequest(h, http.MethodPost, "/key", nil)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", rec.Code)
		}

		allow := rec.Header().Get("Allow")
		if !strings.Contains(allow, "PUT") {
			t.Errorf("expected Allow header to contain PUT, got %s", allow)
		}
	})

	t.Run("full config - all methods allowed", func(t *testing.T) {
		h := newFullTestHandler(map[string]any{"key": "value"}, nil)
		rec := doRequest(h, http.MethodPost, "/key", nil)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", rec.Code)
		}

		allow := rec.Header().Get("Allow")
		if !strings.Contains(allow, "DELETE") {
			t.Errorf("expected Allow header to contain DELETE, got %s", allow)
		}
	})
}

func TestHandler_Middleware(t *testing.T) {
	t.Run("middleware is applied", func(t *testing.T) {
		middlewareCalled := false
		mw := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				middlewareCalled = true
				next.ServeHTTP(w, r)
			})
		}

		h := newTestHandler(map[string]any{"key": "value"}, WithMiddleware(mw))
		rec := doRequest(h, http.MethodGet, "/key", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}
		if !middlewareCalled {
			t.Error("expected middleware to be called")
		}
	})

	t.Run("multiple middleware in order", func(t *testing.T) {
		var order []int
		mw1 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, 1)
				next.ServeHTTP(w, r)
			})
		}
		mw2 := func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				order = append(order, 2)
				next.ServeHTTP(w, r)
			})
		}

		h := newTestHandler(map[string]any{"key": "value"}, WithMiddleware(mw1, mw2))
		doRequest(h, http.MethodGet, "/key", nil)

		if len(order) != 2 || order[0] != 1 || order[1] != 2 {
			t.Errorf("expected middleware order [1, 2], got %v", order)
		}
	})
}

func TestHandler_FilterMap(t *testing.T) {
	data := map[string]any{
		"public": map[string]any{
			"name": "app",
		},
		"secrets": map[string]any{
			"password": "secret123",
		},
	}

	t.Run("filter nested map", func(t *testing.T) {
		h := newFullTestHandler(data, nil, WithKeyFilter(func(key string) bool {
			return !strings.Contains(key, "secrets")
		}))
		rec := doRequest(h, http.MethodGet, "/", nil)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", rec.Code)
		}

		resp := parseResponse[ValueResponse](t, rec)
		valueMap := resp.Value.(map[string]any)
		if _, found := valueMap["secrets"]; found {
			t.Error("expected secrets to be filtered out")
		}
		if _, found := valueMap["public"]; !found {
			t.Error("expected public to be present")
		}
	})
}

func TestHandler_HelperMethods(t *testing.T) {
	t.Run("IsMutable", func(t *testing.T) {
		h1 := newTestHandler(map[string]any{})
		if h1.IsMutable() {
			t.Error("expected IsMutable false for basic config")
		}

		h2 := newMutableTestHandler(map[string]any{})
		if !h2.IsMutable() {
			t.Error("expected IsMutable true for mutable config")
		}

		var h3 *Handler
		if h3.IsMutable() {
			t.Error("expected IsMutable false for nil handler")
		}
	})

	t.Run("IsMarshable", func(t *testing.T) {
		h := newTestHandler(map[string]any{})
		// Our mock doesn't implement MarshableConfig
		if h.IsMarshable() {
			t.Error("expected IsMarshable false")
		}

		var h2 *Handler
		if h2.IsMarshable() {
			t.Error("expected IsMarshable false for nil handler")
		}
	})

	t.Run("Config", func(t *testing.T) {
		cfg := &mockConfig{data: map[string]any{}}
		h := NewHandler(cfg)
		if h.Config() != cfg {
			t.Error("expected Config() to return original config")
		}

		var h2 *Handler
		if h2.Config() != nil {
			t.Error("expected Config() nil for nil handler")
		}
	})
}

func TestHandler_CustomDelimiter(t *testing.T) {
	// With custom delimiter "/", path /server/port becomes key "server/port"
	// The mockConfig.GetValue will try to split on "." which means it looks for
	// "server/port" as a literal key in the root map
	data := map[string]any{
		"server/port": 8080,
	}

	h := newTestHandler(data, WithDelimiter("/"))
	rec := doRequest(h, http.MethodGet, "/server/port", nil)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse[ValueResponse](t, rec)
	if resp.Key != "server/port" {
		t.Errorf("expected key server/port, got %s", resp.Key)
	}
	if resp.Value.(float64) != 8080 {
		t.Errorf("expected value 8080, got %v", resp.Value)
	}
}

func TestResponses_JoinStrings(t *testing.T) {
	tests := []struct {
		strs     []string
		sep      string
		expected string
	}{
		{[]string{}, ", ", ""},
		{[]string{"a"}, ", ", "a"},
		{[]string{"a", "b"}, ", ", "a, b"},
		{[]string{"a", "b", "c"}, "-", "a-b-c"},
	}

	for _, tt := range tests {
		result := joinStrings(tt.strs, tt.sep)
		if result != tt.expected {
			t.Errorf("joinStrings(%v, %q) = %q, expected %q", tt.strs, tt.sep, result, tt.expected)
		}
	}
}

func TestHandler_GetRoot_NoAllGetter_NoEmptyKeySupport(t *testing.T) {
	// Test when GetValue("") fails and there's no AllGetter
	cfg := &mockConfig{data: map[string]any{"key": "value"}}
	cfg.getErr = errors.New("empty key not supported")
	h := NewHandler(cfg)
	rec := doRequest(h, http.MethodGet, "/", nil)

	// Should return the fallback message
	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
}

func TestOptions_WithKeyFilter(t *testing.T) {
	filter := func(key string) bool { return key == "allowed" }
	h := newTestHandler(map[string]any{"allowed": "yes", "denied": "no"}, WithKeyFilter(filter))

	if h.keyFilter == nil {
		t.Error("expected keyFilter to be set")
	}
	if !h.isKeyAllowed("allowed") {
		t.Error("expected 'allowed' to be allowed")
	}
	if h.isKeyAllowed("denied") {
		t.Error("expected 'denied' to be denied")
	}
}

func TestOptions_WithReloadHandler(t *testing.T) {
	called := false
	reloadFn := func(ctx context.Context) error {
		called = true
		return nil
	}

	h := newTestHandler(map[string]any{}, WithReloadHandler(reloadFn))
	if h.onReload == nil {
		t.Error("expected onReload to be set")
	}

	_ = h.onReload(context.Background())
	if !called {
		t.Error("expected reload function to be called")
	}
}

func TestWriteHelpers(t *testing.T) {
	t.Run("writeJSON with nil data", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeJSON(rec, http.StatusNoContent, nil)

		if rec.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", rec.Code)
		}
		if rec.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json")
		}
	})

	t.Run("writeError", func(t *testing.T) {
		rec := httptest.NewRecorder()
		writeError(rec, http.StatusBadRequest, ErrorCodeInvalidKey, "test message", map[string]any{"foo": "bar"})

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", rec.Code)
		}

		resp := parseResponse[ErrorResponse](t, rec)
		if resp.Error.Code != ErrorCodeInvalidKey {
			t.Errorf("expected code %s, got %s", ErrorCodeInvalidKey, resp.Error.Code)
		}
		if resp.Error.Message != "test message" {
			t.Errorf("expected message 'test message', got %s", resp.Error.Message)
		}
	})
}

func TestHandler_ReadBodyError(t *testing.T) {
	h := newMutableTestHandler(map[string]any{})

	// Create a request with a body that fails to read
	req := httptest.NewRequest(http.MethodPut, "/key", &errorReader{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rec.Code)
	}
}

type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestHandler_EmptyKeysResult(t *testing.T) {
	// Test when Keys returns nil
	cfg := &mockFullConfig{
		mockMutableConfig: mockMutableConfig{mockConfig: mockConfig{data: map[string]any{}}},
		keys:              nil,
	}
	h := NewHandler(cfg)
	rec := doRequest(h, http.MethodGet, "/_meta/keys", nil)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}

	resp := parseResponse[KeysResponse](t, rec)
	if resp.Keys == nil {
		t.Error("expected keys to be empty slice, not nil")
	}
	if resp.Count != 0 {
		t.Errorf("expected count 0, got %d", resp.Count)
	}
}
