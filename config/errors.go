package config

import (
	"github.com/grinps/go-utils/errext"
)

// Error codes for the config package.
// These use the errext package for structured error handling with attributes.
var (
	// ErrConfigCodeUnknown represents an unknown configuration error.
	ErrConfigCodeUnknown = errext.NewErrorCode(0)

	// ErrConfigMissingValue indicates a configuration value was not found.
	ErrConfigMissingValue = errext.NewErrorCode(1)

	// ErrConfigEmptyKey indicates an empty key was provided.
	ErrConfigEmptyKey = errext.NewErrorCode(2)

	// ErrConfigInvalidKey indicates the key format is invalid.
	ErrConfigInvalidKey = errext.NewErrorCode(3)

	// ErrConfigInvalidConfig indicates the configuration structure is invalid.
	ErrConfigInvalidConfig = errext.NewErrorCode(4)

	// ErrConfigNilConfig indicates a nil config was encountered.
	ErrConfigNilConfig = errext.NewErrorCode(5)

	// ErrConfigMessageParsingFailed indicates message parsing failed.
	ErrConfigMessageParsingFailed = errext.NewErrorCode(6)

	// ErrConfigInvalidValueType indicates the value type doesn't match the expected type.
	ErrConfigInvalidValueType = errext.NewErrorCode(7)

	// ErrConfigKeyParsingFailed indicates key parsing failed.
	ErrConfigKeyParsingFailed = errext.NewErrorCode(8)

	// ErrConfigInvalidValue indicates the value is invalid or cannot be converted.
	ErrConfigInvalidValue = errext.NewErrorCode(9)

	// ErrConfigNilReturnValue indicates a nil return value pointer was provided.
	ErrConfigNilReturnValue = errext.NewErrorCode(10)
)
