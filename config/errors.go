package config

import (
	"github.com/grinps/go-utils/errext"
)

const (
	ErrorTypeConfig = "github.com/grinps/go-utils/config"
)

// Error codes for the config package.
// These use the errext package for structured error handling with attributes.
var (
	// ErrConfigCodeUnknown represents an unknown configuration error.
	ErrConfigCodeUnknown = errext.NewErrorCodeOfType(0, ErrorTypeConfig)

	// ErrConfigMissingValue indicates a configuration value was not found.
	ErrConfigMissingValue = errext.NewErrorCodeOfType(1, ErrorTypeConfig)

	// ErrConfigEmptyKey indicates an empty key was provided.
	ErrConfigEmptyKey = errext.NewErrorCodeOfType(2, ErrorTypeConfig)

	// ErrConfigInvalidKey indicates the key format is invalid.
	ErrConfigInvalidKey = errext.NewErrorCodeOfType(3, ErrorTypeConfig)

	// ErrConfigInvalidConfig indicates the configuration structure is invalid.
	ErrConfigInvalidConfig = errext.NewErrorCodeOfType(4, ErrorTypeConfig)

	// ErrConfigNilConfig indicates a nil config was encountered.
	ErrConfigNilConfig = errext.NewErrorCodeOfType(5, ErrorTypeConfig)

	// ErrConfigMessageParsingFailed indicates message parsing failed.
	ErrConfigMessageParsingFailed = errext.NewErrorCodeOfType(6, ErrorTypeConfig)

	// ErrConfigInvalidValueType indicates the value type doesn't match the expected type.
	ErrConfigInvalidValueType = errext.NewErrorCodeOfType(7, ErrorTypeConfig)

	// ErrConfigKeyParsingFailed indicates key parsing failed.
	ErrConfigKeyParsingFailed = errext.NewErrorCodeOfType(8, ErrorTypeConfig)

	// ErrConfigInvalidValue indicates the value is invalid or cannot be converted.
	ErrConfigInvalidValue = errext.NewErrorCodeOfType(9, ErrorTypeConfig)

	// ErrConfigNilReturnValue indicates a nil return value pointer was provided.
	ErrConfigNilReturnValue = errext.NewErrorCodeOfType(10, ErrorTypeConfig)

	// ErrConfigUnmarshalFailed indicates unmarshalling to struct failed.
	ErrConfigUnmarshalFailed = errext.NewErrorCodeOfType(11, ErrorTypeConfig)

	// ErrConfigInvalidTarget indicates the target is not a valid pointer to struct.
	ErrConfigInvalidTarget = errext.NewErrorCodeOfType(12, ErrorTypeConfig)

	// ErrConfigSetValueFailed indicates setting a value failed.
	ErrConfigSetValueFailed = errext.NewErrorCodeOfType(13, ErrorTypeConfig)
)
