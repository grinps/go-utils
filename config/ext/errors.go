package ext

import (
	"github.com/grinps/go-utils/errext"
)

// Error codes for the ext package.
var (
	// ErrExtUnmarshalFailed indicates unmarshalling to struct failed.
	ErrExtUnmarshalFailed = errext.NewErrorCode(100)

	// ErrExtInvalidTarget indicates the target is not a valid pointer to struct.
	ErrExtInvalidTarget = errext.NewErrorCode(101)

	// ErrExtKeyNotFound indicates the configuration key was not found.
	ErrExtKeyNotFound = errext.NewErrorCode(102)

	// ErrExtNilConfig indicates a nil config was provided.
	ErrExtNilConfig = errext.NewErrorCode(103)

	// ErrExtSetValueFailed indicates setting a value failed.
	ErrExtSetValueFailed = errext.NewErrorCode(104)

	// ErrExtValueConversionFailed indicates value type conversion failed.
	ErrExtValueConversionFailed = errext.NewErrorCode(105)
)
