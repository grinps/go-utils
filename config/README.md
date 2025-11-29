# Config Package

The `config` package provides a flexible, context-aware configuration management library for Go applications. It supports nested configuration maps, type-safe retrieval with pointer-based assignment, and extensible error handling via the `errext` package.

## Features

- **Context Aware**: All configuration operations accept `context.Context`.
- **Type-Safe Retrieval**: Generic `GetValueE[T]` function with compile-time type safety.
- **Pointer-Based Assignment**: Values are assigned via pointers, enabling default value patterns.
- **Dot-Notation Keys**: Access nested values using dot notation (e.g., `server.port`).
- **Simple In-Memory Implementation**: Includes `SimpleConfig` for easy testing and mocking.
- **Structured Error Handling**: Uses `errext` package for rich error information.
- **High Test Coverage**: >93% test coverage with comprehensive edge case handling.

## Installation

```bash
go get github.com/grinps/go-utils/config
```

## Usage

### Basic Initialization

```go
import (
    "context"
    "github.com/grinps/go-utils/config"
)

func main() {
    ctx := context.Background()
    data := map[string]any{
        "server": map[string]any{
            "port": 8080,
            "host": "localhost",
        },
        "database": map[string]any{
            "host": "db.example.com",
            "port": 5432,
        },
    }
    
    // Initialize with a map
    cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
    
    // Inject into context (recommended for deeper propagation)
    ctx = config.ContextWithConfig(ctx, cfg)
}
```

### Retrieving Values

The API uses pointer-based assignment for type safety and default value support:

```go
// Method 1: Direct config method (pointer-based)
var port int
err := cfg.GetValue(ctx, "server.port", &port)
if err != nil {
    log.Fatal(err)
}
fmt.Println(port) // 8080

// Method 2: Package-level function (recommended)
var host string
err = config.GetValueE(ctx, "server.host", &host)
if err != nil {
    log.Fatal(err)
}
fmt.Println(host) // "localhost"

// Method 3: With default value pattern
timeout := 30 // default value
err = config.GetValueE(ctx, "server.timeout", &timeout)
// If key is missing, timeout retains its default value (30)
// If key exists, timeout is updated with the configured value
if err != nil {
    // Handle error or ignore if default is acceptable
    log.Printf("Using default timeout: %d", timeout)
}
```

### Working with Nested Configurations

```go
// Get a sub-configuration
serverConfig, err := cfg.GetConfig(ctx, "server")
if err != nil {
    log.Fatal(err)
}

var host string
err = serverConfig.GetValue(ctx, "host", &host)
fmt.Println(host) // "localhost"
```

### Custom Delimiters

```go
// Use a custom delimiter instead of "."
cfg := config.NewSimpleConfig(ctx, 
    config.WithConfigurationMap(data),
    config.WithDelimiter("/"))

var port int
err := cfg.GetValue(ctx, "server/port", &port)
```

## API Reference

### Config Interface

```go
type Config interface {
    // GetValue retrieves a configuration value by key and stores it in the provided pointer.
    // The returnValue parameter must be a non-nil pointer.
    GetValue(ctx context.Context, key string, returnValue any) error

    // GetConfig retrieves a nested configuration as a new Config instance.
    GetConfig(ctx context.Context, key string) (Config, error)
}
```

### Package Functions

```go
// GetValueE retrieves a configuration value and stores it in returnValue.
// If returnValue contains a default value, it is preserved when the key is not found.
// Returns an error if the key is empty, config is nil, or type conversion fails.
func GetValueE[T any](ctx context.Context, key string, returnValue *T) error
```

### Context Functions

```go
// ContextWithConfig attaches a Config to a context
func ContextWithConfig(ctx context.Context, config Config) context.Context

// ContextConfig retrieves a Config from context
func ContextConfig(ctx context.Context, defaultIfNotAvailable bool) Config

// Default returns the package-level default Config
func Default() Config
```

## Error Handling

The package uses the `errext` package for structured error handling:

```go
var port int
err := cfg.GetValue(ctx, "missing.key", &port)
if err != nil {
    // Error includes structured information
    fmt.Println(err) // [ErrorCode: 1] missing value {key=missing.key}
}
```

### Error Codes

- `ErrConfigCodeUnknown` - Unknown error
- `ErrConfigMissingValue` - Value not found
- `ErrConfigEmptyKey` - Empty key provided
- `ErrConfigInvalidKey` - Invalid key format
- `ErrConfigNilConfig` - Nil config encountered
- `ErrConfigInvalidValueType` - Type mismatch
- `ErrConfigInvalidValue` - Invalid value or conversion error
- `ErrConfigNilReturnValue` - Nil return value pointer provided

## Examples

### Example 1: Application Configuration

```go
type AppConfig struct {
    ServerPort int
    ServerHost string
    DBHost     string
    DBPort     int
}

func LoadConfig(ctx context.Context) (*AppConfig, error) {
    cfg := config.ContextConfig(ctx, true)
    
    var appCfg AppConfig
    if err := cfg.GetValue(ctx, "server.port", &appCfg.ServerPort); err != nil {
        return nil, err
    }
    if err := cfg.GetValue(ctx, "server.host", &appCfg.ServerHost); err != nil {
        return nil, err
    }
    if err := cfg.GetValue(ctx, "database.host", &appCfg.DBHost); err != nil {
        return nil, err
    }
    if err := cfg.GetValue(ctx, "database.port", &appCfg.DBPort); err != nil {
        return nil, err
    }
    
    return &appCfg, nil
}
```

### Example 2: Default Values

```go
// Set defaults
logLevel := "info"
maxConnections := 100
timeout := 30 * time.Second

// Override with config if available (errors are ignored, defaults preserved)
_ = config.GetValueE(ctx, "log.level", &logLevel)
_ = config.GetValueE(ctx, "db.max_connections", &maxConnections)
_ = config.GetValueE(ctx, "server.timeout", &timeout)

// Values now contain either configured values or defaults
fmt.Printf("Log Level: %s\n", logLevel)
fmt.Printf("Max Connections: %d\n", maxConnections)
fmt.Printf("Timeout: %v\n", timeout)
```

### Example 3: Error Handling Patterns

```go
// Pattern 1: Fail on missing required config
var apiKey string
if err := config.GetValueE(ctx, "api.key", &apiKey); err != nil {
    return fmt.Errorf("required config 'api.key' missing: %w", err)
}

// Pattern 2: Use default on missing optional config
retryCount := 3 // default
if err := config.GetValueE(ctx, "api.retry_count", &retryCount); err != nil {
    log.Printf("Using default retry count: %d", retryCount)
}

// Pattern 3: Check specific error types
var dbHost string
if err := config.GetValueE(ctx, "database.host", &dbHost); err != nil {
    if errors.Is(err, config.ErrConfigMissingValue) {
        dbHost = "localhost" // fallback
    } else {
        return err // other errors are fatal
    }
}
```

## Testing

The package includes comprehensive tests with >93% coverage:

```bash
cd config
go test -v ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## License

Part of the go-utils project.

