package confighttp

import (
	"github.com/grinps/go-utils/errext"
)

const (
	// ErrorTypeConfigHTTP is the error type for config/http package errors.
	ErrorTypeConfigHTTP = "github.com/grinps/go-utils/config/http"
)

// Error codes for the config/http package.
var (
	// ErrHTTPNilConfig indicates a nil config was provided to the handler.
	ErrHTTPNilConfig = errext.NewErrorCodeOfType(1, ErrorTypeConfigHTTP)

	// ErrHTTPKeyNotFound indicates the requested configuration key was not found.
	ErrHTTPKeyNotFound = errext.NewErrorCodeOfType(2, ErrorTypeConfigHTTP)

	// ErrHTTPInvalidKey indicates an invalid key format was provided.
	ErrHTTPInvalidKey = errext.NewErrorCodeOfType(3, ErrorTypeConfigHTTP)

	// ErrHTTPReadOnly indicates a write operation was attempted in read-only mode.
	ErrHTTPReadOnly = errext.NewErrorCodeOfType(4, ErrorTypeConfigHTTP)

	// ErrHTTPNotMutable indicates the config doesn't support mutations.
	ErrHTTPNotMutable = errext.NewErrorCodeOfType(5, ErrorTypeConfigHTTP)

	// ErrHTTPInvalidBody indicates the request body is invalid.
	ErrHTTPInvalidBody = errext.NewErrorCodeOfType(6, ErrorTypeConfigHTTP)

	// ErrHTTPMethodNotAllowed indicates the HTTP method is not allowed.
	ErrHTTPMethodNotAllowed = errext.NewErrorCodeOfType(7, ErrorTypeConfigHTTP)

	// ErrHTTPKeyFiltered indicates access to the key is filtered/denied.
	ErrHTTPKeyFiltered = errext.NewErrorCodeOfType(8, ErrorTypeConfigHTTP)

	// ErrHTTPDeleteNotSupported indicates delete is not supported by the config.
	ErrHTTPDeleteNotSupported = errext.NewErrorCodeOfType(9, ErrorTypeConfigHTTP)

	// ErrHTTPReloadNotConfigured indicates reload handler is not configured.
	ErrHTTPReloadNotConfigured = errext.NewErrorCodeOfType(10, ErrorTypeConfigHTTP)

	// ErrHTTPReloadFailed indicates the reload operation failed.
	ErrHTTPReloadFailed = errext.NewErrorCodeOfType(11, ErrorTypeConfigHTTP)

	// ErrHTTPInternalError indicates an internal server error occurred.
	ErrHTTPInternalError = errext.NewErrorCodeOfType(12, ErrorTypeConfigHTTP)

	// ErrHTTPSetValueFailed indicates setting a value failed.
	ErrHTTPSetValueFailed = errext.NewErrorCodeOfType(13, ErrorTypeConfigHTTP)

	// ErrHTTPDeleteFailed indicates deleting a key failed.
	ErrHTTPDeleteFailed = errext.NewErrorCodeOfType(14, ErrorTypeConfigHTTP)
)

// ErrorCode represents an error code string for API responses.
type ErrorCode string

// Error code strings for API responses.
const (
	ErrorCodeNilConfig          ErrorCode = "CONFIG_NIL"
	ErrorCodeKeyNotFound        ErrorCode = "CONFIG_KEY_NOT_FOUND"
	ErrorCodeInvalidKey         ErrorCode = "CONFIG_INVALID_KEY"
	ErrorCodeReadOnly           ErrorCode = "CONFIG_READ_ONLY"
	ErrorCodeNotMutable         ErrorCode = "CONFIG_NOT_MUTABLE"
	ErrorCodeInvalidBody        ErrorCode = "CONFIG_INVALID_BODY"
	ErrorCodeMethodNotAllowed   ErrorCode = "METHOD_NOT_ALLOWED"
	ErrorCodeKeyFiltered        ErrorCode = "CONFIG_KEY_FILTERED"
	ErrorCodeDeleteNotSupported ErrorCode = "CONFIG_DELETE_NOT_SUPPORTED"
	ErrorCodeReloadNotConfig    ErrorCode = "CONFIG_RELOAD_NOT_CONFIGURED"
	ErrorCodeReloadFailed       ErrorCode = "CONFIG_RELOAD_FAILED"
	ErrorCodeInternalError      ErrorCode = "INTERNAL_ERROR"
	ErrorCodeSetValueFailed     ErrorCode = "CONFIG_SET_VALUE_FAILED"
	ErrorCodeDeleteFailed       ErrorCode = "CONFIG_DELETE_FAILED"
)
