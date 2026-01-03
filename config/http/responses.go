package confighttp

import (
	"encoding/json"
	"net/http"
)

// ValueResponse represents a successful value retrieval response.
type ValueResponse struct {
	Key   string `json:"key"`
	Value any    `json:"value"`
}

// KeysResponse represents a response containing configuration keys.
type KeysResponse struct {
	Keys   []string `json:"keys"`
	Count  int      `json:"count"`
	Prefix string   `json:"prefix,omitempty"`
}

// InfoResponse represents handler information response.
type InfoResponse struct {
	Provider    string `json:"provider"`
	Mutable     bool   `json:"mutable"`
	Marshable   bool   `json:"marshable"`
	HasKeys     bool   `json:"has_keys"`
	HasDelete   bool   `json:"has_delete"`
	HasReload   bool   `json:"has_reload"`
	ReadOnly    bool   `json:"read_only"`
	BasePath    string `json:"base_path,omitempty"`
	KeyFiltered bool   `json:"key_filtered"`
}

// SetValueRequest represents a request to set a configuration value.
type SetValueRequest struct {
	Value any `json:"value"`
}

// ErrorDetail contains details about an error.
type ErrorDetail struct {
	Code    ErrorCode      `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// SuccessResponse represents a generic success response.
type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// writeError writes an error response with the given status code and error details.
func writeError(w http.ResponseWriter, status int, code ErrorCode, message string, details map[string]any) {
	writeJSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// writeValue writes a value response.
func writeValue(w http.ResponseWriter, key string, value any) {
	writeJSON(w, http.StatusOK, ValueResponse{
		Key:   key,
		Value: value,
	})
}

// writeKeys writes a keys response.
func writeKeys(w http.ResponseWriter, keys []string, prefix string) {
	writeJSON(w, http.StatusOK, KeysResponse{
		Keys:   keys,
		Count:  len(keys),
		Prefix: prefix,
	})
}

// writeSuccess writes a success response.
func writeSuccess(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, SuccessResponse{
		Success: true,
		Message: message,
	})
}

// writeNotFound writes a 404 not found error.
func writeNotFound(w http.ResponseWriter, key string) {
	writeError(w, http.StatusNotFound, ErrorCodeKeyNotFound, "key not found", map[string]any{"key": key})
}

// writeMethodNotAllowed writes a 405 method not allowed error.
func writeMethodNotAllowed(w http.ResponseWriter, method string, allowed []string) {
	w.Header().Set("Allow", joinStrings(allowed, ", "))
	writeError(w, http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed, "method not allowed", map[string]any{
		"method":  method,
		"allowed": allowed,
	})
}

// writeBadRequest writes a 400 bad request error.
func writeBadRequest(w http.ResponseWriter, code ErrorCode, message string, details map[string]any) {
	writeError(w, http.StatusBadRequest, code, message, details)
}

// writeInternalError writes a 500 internal server error.
func writeInternalError(w http.ResponseWriter, message string) {
	writeError(w, http.StatusInternalServerError, ErrorCodeInternalError, message, nil)
}

// writeForbidden writes a 403 forbidden error.
func writeForbidden(w http.ResponseWriter, code ErrorCode, message string, details map[string]any) {
	writeError(w, http.StatusForbidden, code, message, details)
}

// writeNotImplemented writes a 501 not implemented error.
func writeNotImplemented(w http.ResponseWriter, code ErrorCode, message string) {
	writeError(w, http.StatusNotImplemented, code, message, nil)
}

// joinStrings joins strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
