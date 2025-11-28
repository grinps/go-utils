// Package errext extends core error functionality in Go to support a standard pattern of error generation and handling.
//
// This package provides the following extension to error.
//  1. Cause - The original error that was thrown (extractable using Unwrap function).
//  2. Code value - A code (integer) can be assigned to an error to uniquely identify the error. This allows
//     creation of comparable errors that may not have a matching string. For example two different calls to same function can return
//     two different error object, but we can use error code to compare and validate that the error is same.
//  3. Error type - A way to classify errors for comparison. For example different errors can be marked as input value error type
//     which can be used to write handlers that can provide a common response for invalid input.
//  4. Error message template - allows specifying pre-defined error template with specific parameters. The parameter values
//     can be provided while creating errors using [errext.ErrorCode].
//  5. Stack Capture - Captures stack trace when errors are created. Controlled via [errext.EnableStackTrace] flag (default false).
//     Use %+v with fmt to print stack traces.
//  6. Error Matching - Supports [errors.As] for both the error itself and extracting the underlying [errext.ErrorCode].
//
// All the above capabilities are available through [errext.ErrorCode] which define an error template that can be used to create
// errors.
package errext
