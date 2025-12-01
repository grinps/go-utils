package koanf

import (
	"github.com/grinps/go-utils/errext"
)

const (
	ErrorTypeKoanf = "github.com/grinps/go-utils/config/koanf"
)

// Error codes for the koanf package.
// These use the errext package for structured error handling with attributes.
var (
	// ErrKoanfUnknown represents an unknown koanf error.
	ErrKoanfUnknown = errext.NewErrorCodeOfType(0, ErrorTypeKoanf)

	// ErrKoanfNilConfig indicates a nil koanf config was encountered.
	ErrKoanfNilConfig = errext.NewErrorCodeOfType(1, ErrorTypeKoanf)

	// ErrKoanfEmptyKey indicates an empty key was provided.
	ErrKoanfEmptyKey = errext.NewErrorCodeOfType(2, ErrorTypeKoanf)

	// ErrKoanfMissingValue indicates a configuration value was not found.
	ErrKoanfMissingValue = errext.NewErrorCodeOfType(3, ErrorTypeKoanf)

	// ErrKoanfInvalidValue indicates the value is invalid or cannot be converted.
	ErrKoanfInvalidValue = errext.NewErrorCodeOfType(4, ErrorTypeKoanf)

	// ErrKoanfSetValueFailed indicates setting a value failed.
	ErrKoanfSetValueFailed = errext.NewErrorCodeOfType(5, ErrorTypeKoanf)

	// ErrKoanfUnmarshalFailed indicates unmarshalling to struct failed.
	ErrKoanfUnmarshalFailed = errext.NewErrorCodeOfType(6, ErrorTypeKoanf)

	// ErrKoanfInvalidTarget indicates the target is not a valid pointer to struct.
	ErrKoanfInvalidTarget = errext.NewErrorCodeOfType(7, ErrorTypeKoanf)

	// ErrKoanfLoadFailed indicates loading configuration from provider failed.
	ErrKoanfLoadFailed = errext.NewErrorCodeOfType(8, ErrorTypeKoanf)

	// ErrKoanfMergeFailed indicates merging configuration failed.
	ErrKoanfMergeFailed = errext.NewErrorCodeOfType(9, ErrorTypeKoanf)

	// ErrKoanfInvalidProvider indicates an invalid provider was provided.
	ErrKoanfInvalidProvider = errext.NewErrorCodeOfType(10, ErrorTypeKoanf)
)
