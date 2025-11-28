# errext - Extended Error Handling for Go

`errext` extends standard Go error functionality to support structured error generation, categorization, and contextual attributes. It is designed for applications that require strict error codes and machine-readable error types while maintaining idiomatic Go error compatibility.

## Features

*   **Error Codes & Types:** Assign integer codes and string types to errors for reliable categorization and comparison.
*   **Structured Attributes:** Add slog-style key-value pairs to errors for structured context and observability.
*   **Stack Traces:** Optional stack trace capture for debugging, printed via `%+v`.
*   **Stdlib Compatibility:** Fully implements `error` interface, supports `errors.Is`, `errors.As`, and `errors.Unwrap`.
*   **Uniqueness Registry:** Prevents duplicate error codes for the same error type.
*   **Panic Recovery:** Utilities to safely handle panics and convert them to structured errors.

## Installation

```bash
go get github.com/grinps/go-utils/errext
```

## Usage

### Defining Error Codes

Create reusable `ErrorCode` definitions at the package level. This acts as a factory for your errors.

```go
package mypkg

import "github.com/grinps/go-utils/errext"

// Define a unique error code
var ErrInvalidInput = errext.NewErrorCode(1001)

// Define with Type
var ErrDatabaseConnection = errext.NewErrorCodeOfType(2001, "DatabaseError")
```

### Creating Errors

Use the `ErrorCode` to generate error instances.

```go
func ProcessInput(val string) error {
    if val == "" {
        return ErrInvalidInput.New("input cannot be empty")
    }
    return nil
}

// Wrapping existing errors
func ConnectDB() error {
    err := db.Connect()
    if err != nil {
        return ErrDatabaseConnection.NewWithError("connection failed", err)
    }
    return nil
}
```

### Stack Traces

Stack trace capture is **disabled by default** to minimize performance overhead. Enable it globally in your application startup:

```go
func init() {
    errext.EnableStackTrace = true
}
```

When enabled, you can print the stack trace using `%+v`:

```go
err := ErrInvalidInput.New("bad data")
fmt.Printf("%+v\n", err)
```

### Error Checking

Use `errors.Is` and `errors.As` as usual.

```go
if errors.Is(err, ErrInvalidInput) {
    // Handle invalid input
}

// Extracting metadata
var ec errext.ErrorCode
if errors.As(err, &ec) {
    // Access specific ErrorCode logic
}
```

### Structured Attributes

Add context to errors using key-value pairs.

```go
// Default attributes on ErrorCode
var ErrDatabaseOp = errext.NewErrorCodeWithOptions(
    errext.WithErrorCode(5001),
    errext.WithAttributes("component", "database", "severity", "high"),
)

// Runtime attributes when creating errors
err := ErrDatabaseOp.New("query failed", "table", "users", "query_time_ms", 1500)
// Error string: "query failed [component=database severity=high table=users query_time_ms=1500]"

// Attributes are formatted in slog style: [key1=val1 key2=val2]
```

## Performance

*   **Memory:** `ErrorCode` instances should be created once (globals). `Error` instances are lightweight wrappers.
*   **Stack Traces:** Capturing stack traces involves memory allocation. Keep `EnableStackTrace` false in high-throughput paths if not needed, or toggle it via configuration.

## Contributing

Run tests:
```bash
go test ./...
```
