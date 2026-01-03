package ext

import (
	"github.com/grinps/go-utils/errext"
)

const (
	ErrorTypeExt = "github.com/grinps/go-utils/config/ext"
)

// Error codes for the ext package.
var (
	// ErrExtUnmarshalFailed indicates unmarshalling to struct failed.
	ErrExtUnmarshalFailed = errext.NewErrorCodeOfType(100, ErrorTypeExt)

	// ErrExtInvalidTarget indicates the target is not a valid pointer to struct.
	ErrExtInvalidTarget = errext.NewErrorCodeOfType(101, ErrorTypeExt)

	// ErrExtKeyNotFound indicates the configuration key was not found.
	ErrExtKeyNotFound = errext.NewErrorCodeOfType(102, ErrorTypeExt)

	// ErrExtNilConfig indicates a nil config was provided.
	ErrExtNilConfig = errext.NewErrorCodeOfType(103, ErrorTypeExt)

	// ErrExtSetValueFailed indicates setting a value failed.
	ErrExtSetValueFailed = errext.NewErrorCodeOfType(104, ErrorTypeExt)

	// ErrExtDeleteNotSupported indicates the wrapped config doesn't support delete.
	ErrExtDeleteNotSupported = errext.NewErrorCodeOfType(105, ErrorTypeExt)
)
