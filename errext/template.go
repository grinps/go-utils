package errext

import (
	"fmt"
	"strings"
)

// WithAttributes adds default attributes to the ErrorCode.
// These attributes will be included in all errors created from this code.
// Attributes should be provided as alternating key-value pairs in slog style.
//
// Example:
//
//	ec := errext.NewErrorCodeWithOptions(
//	    errext.WithErrorCode(100),
//	    errext.WithAttributes("component", "database", "severity", "high"),
//	)
//	err := ec.New("connection failed", "retry_count", 3)
//	// Output: "connection failed [component=database severity=high retry_count=3]"
//
// If an odd number of arguments is provided, the last key will have a "null" value.
func WithAttributes(args ...interface{}) ErrorCodeOptions {
	return func(errorCode *ErrorCodeImpl) *ErrorCodeImpl {
		if errorCode != nil {
			errorCode.defaultArgs = append(errorCode.defaultArgs, args...)
		}
		return errorCode
	}
}

// formatAttributes formats the key-value pairs into a string using slog-style formatting.
// Arguments are expected to be alternating keys and values.
// Output format: [key1=val1 key2=val2 ...]
// If an odd number of arguments is provided, the last key will have value "null".
func formatAttributes(args []interface{}) string {
	if len(args) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("[")
	for i := 0; i < len(args); i += 2 {
		if i > 0 {
			sb.WriteString(" ")
		}
		key := fmt.Sprint(args[i])
		val := "null"
		if i+1 < len(args) {
			val = fmt.Sprint(args[i+1])
		}
		sb.WriteString(key)
		sb.WriteString("=")
		sb.WriteString(val)
	}
	sb.WriteString("]")
	return sb.String()
}
