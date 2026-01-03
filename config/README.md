# Config Package

The `config` package provides a flexible, context-aware configuration management library for Go applications. It supports nested configuration maps, type-safe retrieval with pointer-based assignment, and extensible error handling via the `errext` package.

## Features

- **Context Aware**: All configuration operations accept `context.Context`.
- **Dual Retrieval APIs**: Direct `GetValue(ctx, key) (any, error)` and type-safe `GetValueE[T](ctx, key, *T) error`.
- **Type-Safe Retrieval**: Generic `GetValueE[T]` function with compile-time type safety and pointer-based assignment.
- **Default Value Pattern**: GetValueE preserves existing values when keys are not found.
- **Dot-Notation Keys**: Access nested values using dot notation (e.g., `server.port`).
- **Simple In-Memory Implementation**: Includes `SimpleConfig` for easy testing and mocking.
- **Structured Error Handling**: Uses `errext` package for rich error information.
- **Built-in Telemetry**: Integrated tracing and metrics via the `telemetry` package.
- **High Test Coverage**: ~95% test coverage with comprehensive edge case handling.

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

The package provides two retrieval approaches:

```go
// Method 1: Direct GetValue - returns (any, error)
val, err := cfg.GetValue(ctx, "server.port")
if err != nil {
    log.Fatal(err)
}
port := val.(int) // Type assertion required
fmt.Println(port) // 8080

// Method 2: Type-safe GetValueE (recommended)
var host string
err = config.GetValueE(ctx, "server.host", &host)
if err != nil {
    log.Fatal(err)
}
fmt.Println(host) // "localhost"

// Method 3: With default value pattern (GetValueE only)
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

val, err := serverConfig.GetValue(ctx, "host")
if err != nil {
    log.Fatal(err)
}
host := val.(string)
fmt.Println(host) // "localhost"
```

### Custom Delimiters

```go
// Use a custom delimiter instead of "."
cfg := config.NewSimpleConfig(ctx, 
    config.WithConfigurationMap(data),
    config.WithDelimiter("/"))

val, err := cfg.GetValue(ctx, "server/port")
port := val.(int)
```

## API Reference

### Config Interface

```go
type Config interface {
    // Name returns the provider name (e.g., "SimpleConfig", "KoanfConfig").
    Name() ProviderName

    // GetValue retrieves a configuration value by key.
    // Returns the value and an error if the key is not found.
    GetValue(ctx context.Context, key string) (any, error)

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

// SetAsDefault sets a custom default Config
func SetAsDefault(cfg Config)
```

### Explicit Config Functions

```go
// GetValueWithConfig retrieves a value from a specific config (not from context)
func GetValueWithConfig[T any](ctx context.Context, cfg Config, key string, returnValue *T) error

// GetConfig retrieves a nested config from context
func GetConfig(ctx context.Context, key string) (Config, error)

// GetConfigWithConfig retrieves a nested config from a specific config
func GetConfigWithConfig(ctx context.Context, cfg Config, key string) (Config, error)
```

### Optional Interfaces

The package defines optional interfaces that Config implementations may support:

```go
// AllGetter returns all configuration as a map
type AllGetter interface {
    All(ctx context.Context) map[string]any
}

// AllKeysProvider lists all configuration keys
type AllKeysProvider interface {
    Keys(prefix string) []string
}

// Deleter supports key deletion
type Deleter interface {
    Delete(key string) error
}
```

SimpleConfig implements all optional interfaces:

```go
// Get all configuration
if allGetter, ok := cfg.(config.AllGetter); ok {
    all := allGetter.All(ctx)
    fmt.Printf("Config: %v\n", all)
}

// List all keys
if keysProvider, ok := cfg.(config.AllKeysProvider); ok {
    keys := keysProvider.Keys("")           // All keys
    serverKeys := keysProvider.Keys("server") // Keys with prefix
}

// Delete a key
if deleter, ok := cfg.(config.Deleter); ok {
    err := deleter.Delete("server.debug")
}
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

### Example 4: Custom Default Configuration

```go
// Set a custom default configuration at application startup
data := map[string]any{
    "app": map[string]any{
        "name": "my-app",
        "env":  "production",
    },
}
cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
config.SetAsDefault(cfg)

// Now any code that uses ContextConfig with defaultIfNotAvailable=true
// will get this custom config when no config is in context
```

### Example 5: Using Explicit Config Functions

```go
// When you have multiple configs and need to use a specific one
primaryCfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(primaryData))
backupCfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(backupData))

// Use GetValueWithConfig for explicit config retrieval
var port int
if err := config.GetValueWithConfig(ctx, primaryCfg, "server.port", &port); err != nil {
    // Fallback to backup config
    _ = config.GetValueWithConfig(ctx, backupCfg, "server.port", &port)
}

// Get nested config sections
serverCfg, err := config.GetConfigWithConfig(ctx, primaryCfg, "server")
if err != nil {
    log.Fatal(err)
}
val, _ := serverCfg.GetValue(ctx, "host")
```

## Telemetry

The config package includes built-in telemetry support using the `github.com/grinps/go-utils/telemetry` package. Telemetry automatically captures spans and metrics for all configuration operations.

### Metrics

The following metrics are automatically recorded:

| Metric | Type | Description |
|--------|------|-------------|
| `config.get_value.count` | Counter | Number of GetValue operations |
| `config.get_value.duration_ms` | Histogram | Duration of GetValue operations |
| `config.set_value.count` | Counter | Number of SetValue operations |
| `config.set_value.duration_ms` | Histogram | Duration of SetValue operations |
| `config.get_config.count` | Counter | Number of GetConfig operations |
| `config.get_config.duration_ms` | Histogram | Duration of GetConfig operations |
| `config.unmarshal.count` | Counter | Number of Unmarshal operations |
| `config.unmarshal.duration_ms` | Histogram | Duration of Unmarshal operations |
| `config.errors.count` | Counter | Number of errors across all operations |

### Attributes

All telemetry includes these attributes:

- `config.key_prefix` - First segment of the key (for cardinality control)
- `config.impl_type` - Implementation name (e.g., "SimpleConfig")
- `config.success` - Whether the operation succeeded
- `config.error_code` - Error code (when applicable)
- `config.target_type` - Target struct type (for Unmarshal operations)

### Controlling Telemetry

```go
// Disable telemetry globally
config.SetTelemetryEnabled(false)

// Check if telemetry is enabled
if config.IsTelemetryEnabled() {
    // telemetry is active
}

// Re-enable telemetry
config.SetTelemetryEnabled(true)
```

### TelemetryAware Interface

Custom Config implementations can implement `TelemetryAware` for fine-grained control:

```go
type TelemetryAware interface {
    // ShouldInstrument allows opting out of telemetry for specific operations
    ShouldInstrument(ctx context.Context, key string, op string) bool

    // GenerateTelemetryAttributes allows adding custom attributes
    GenerateTelemetryAttributes(ctx context.Context, op string, attrs []any) []any
}
```

Example implementation:

```go
type MyConfig struct {
    // ...
}

func (c *MyConfig) ShouldInstrument(ctx context.Context, key, op string) bool {
    // Skip telemetry for sensitive keys
    return !strings.HasPrefix(key, "secrets.")
}

func (c *MyConfig) GenerateTelemetryAttributes(ctx context.Context, op string, attrs []any) []any {
    // Add custom attributes
    return append(attrs, "config.source", "etcd", "config.version", "v2")
}
```

## Testing

The package includes comprehensive tests with ~95% coverage:

```bash
cd config
go test -v ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## License

Part of the go-utils project.

