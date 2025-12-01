# Go Base Utils

A comprehensive collection of Go utility packages providing foundational building blocks for Go applications.

## Overview

The `base-utils` library provides a set of well-tested, production-ready utilities for common Go programming patterns including:

- **Platform Abstraction** - OS-level operations abstraction for testability
- **Dependency Injection System** - Service registry and dependency management
- **Channel Selection** - Advanced channel selection and monitoring utilities
- **I/O Utilities** - Source resolution and I/O abstractions
- **Logging** - Environment-configurable logging utilities
- **Type Utilities** - Comparison, equality, and string handling interfaces
- **Generic Registry** - Type-safe registry with comparable keys
- **Extended Error Handling** - Structured error generation, categorization, and templating
- **Configuration Management** - Flexible, context-aware configuration management and extensions like Struct unmarshalling, mutable configs, and context-based config discovery

## Installation

```bash
go get github.com/grinps/go-utils/base-utils
```

## Modules

### 1. Platform Package

**Import:** `github.com/grinps/go-utils/base-utils/platform`

Provides an abstraction layer over OS-level operations to enable better testing and alternative implementations without changing code.

#### Features
- **Signal Handling** - OS signal operations (SIGINT, SIGTERM, etc.)
- **Environment Variables** - Getting/setting environment variables
- **File Operations** - File system operations (read, write, stat, etc.)
- **Process Operations** - Process-related operations (PID, hostname, etc.)
- **Clock/Time** - Time operations for testing time-dependent code

#### Quick Example

```go
import "github.com/grinps/go-utils/base-utils/platform"

// Production code
p := platform.NewOSPlatform()
value := p.Env().Getenv("MY_VAR")
data, err := p.File().ReadFile("/path/to/file")

// Test code
mock := platform.NewMockPlatform()
mock.Env().Setenv("MY_VAR", "test_value")
mock.File().WriteFile("/test/file", []byte("content"), 0644)
```

<details>
<summary><b>📖 Detailed Platform Documentation (Click to expand)</b></summary>

#### Interface Reference

##### Platform Interface

The main interface that provides access to all subsystems:

```go
type Platform interface {
    Signal() SignalHandler
    Env() EnvHandler
    File() FileHandler
    Process() ProcessHandler
    Clock() Clock
}
```

##### SignalHandler Interface

Handle OS signals:

```go
type SignalHandler interface {
    Notify(c chan<- os.Signal, sig ...os.Signal)
    Stop(c chan<- os.Signal)
    Ignore(sig ...os.Signal)
    Reset(sig ...os.Signal)
}
```

**Example:**
```go
sigChan := make(chan os.Signal, 1)
p.Signal().Notify(sigChan, os.Interrupt, syscall.SIGTERM)
sig := <-sigChan
```

##### EnvHandler Interface

Manage environment variables:

```go
type EnvHandler interface {
    Getenv(key string) string
    Setenv(key, value string) error
    Unsetenv(key string) error
    LookupEnv(key string) (string, bool)
    Environ() []string
    Clearenv()
    ExpandEnv(s string) string
}
```

##### FileHandler Interface

Perform file system operations:

```go
type FileHandler interface {
    Open(name string) (File, error)
    Create(name string) (File, error)
    OpenFile(name string, flag int, perm fs.FileMode) (File, error)
    Remove(name string) error
    RemoveAll(path string) error
    Rename(oldpath, newpath string) error
    Stat(name string) (fs.FileInfo, error)
    Lstat(name string) (fs.FileInfo, error)
    ReadFile(name string) ([]byte, error)
    WriteFile(name string, data []byte, perm fs.FileMode) error
    Mkdir(name string, perm fs.FileMode) error
    MkdirAll(path string, perm fs.FileMode) error
    ReadDir(name string) ([]fs.DirEntry, error)
    Getwd() (dir string, err error)
    Chdir(dir string) error
    TempDir() string
    UserHomeDir() (string, error)
}
```

##### ProcessHandler Interface

Access process information:

```go
type ProcessHandler interface {
    Getpid() int
    Getppid() int
    Getuid() int
    Geteuid() int
    Getgid() int
    Getegid() int
    Exit(code int)
    Hostname() (name string, err error)
    FindProcess(pid int) (Process, error)
    StartProcess(name string, argv []string, attr *os.ProcAttr) (Process, error)
}
```

##### Clock Interface

Handle time operations:

```go
type Clock interface {
    Now() time.Time
    Sleep(d time.Duration)
    After(d time.Duration) <-chan time.Time
    Tick(d time.Duration) <-chan time.Time
    NewTimer(d time.Duration) Timer
    NewTicker(d time.Duration) Ticker
}
```

#### Mock Platform Usage

The mock platform provides controllable implementations for testing:

**MockClock - Control Time in Tests:**
```go
mock := platform.NewMockPlatform()
clock := mock.Clock().(*platform.MockClock)

// Set specific time
testTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
clock.Set(testTime)

// Advance time
clock.Advance(1 * time.Hour)
now := clock.Now() // Returns testTime + 1 hour
```

**MockFileHandler - In-Memory File System:**
```go
file := mock.File()
file.WriteFile("/test/file.txt", []byte("content"), 0644)
data, _ := file.ReadFile("/test/file.txt") // Returns "content"
```

**MockProcessHandler - Controllable Process Info:**
```go
proc := mock.Process().(*platform.MockProcessHandler)
proc.SetPid(12345)
proc.SetHostname("test-host")
```

#### Testing Patterns

**Pattern 1: Constructor Injection**
```go
type Service struct {
    platform platform.Platform
}

func NewService(p platform.Platform) *Service {
    return &Service{platform: p}
}

// Test
func TestService(t *testing.T) {
    mock := platform.NewMockPlatform()
    service := NewService(mock)
    // Test service...
}
```

**Pattern 2: Time-Dependent Testing**
```go
func TestTimeDependent(t *testing.T) {
    mock := platform.NewMockPlatform()
    clock := mock.Clock().(*platform.MockClock)
    
    clock.Set(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
    result := DoSomethingTimeDependent(mock)
    
    clock.Advance(24 * time.Hour)
    result2 := DoSomethingTimeDependent(mock)
    // Assert results...
}
```

</details>

---

### 2. System Package

**Import:** `github.com/grinps/go-utils/base-utils/system`

Provides a dependency injection system with service registry and retrieval capabilities.

#### Features
- Service registration and retrieval
- Type-safe service management
- Context-aware operations
- Configurable get/registration options
- Support for service lifecycle management

#### Quick Example

```go
import (
    "context"
    "github.com/grinps/go-utils/base-utils/system"
)

// Create a system
sys := system.NewSystem()

// Register a service
err := sys.RegisterService(ctx, "database", dbInstance)

// Retrieve a service
db, err := sys.GetService(ctx, "database", "DatabaseType")
```

#### Key Interfaces

- `System` - Core system interface
- `RegistrySystem` - Service registration and retrieval
- `GetOption` - Options for service retrieval
- `RegistrationOption` - Options for service registration

---

### 3. GoSub Package

**Import:** `github.com/grinps/go-utils/base-utils/gosub`

Advanced channel selection and monitoring utilities for managing multiple goroutines and channels.

#### Features
- Multi-channel selection with callbacks
- Context-based selection
- Signal-based selection
- Timer-based selection
- Proxy channel management
- Event-driven architecture

#### Quick Example

```go
import "github.com/grinps/go-utils/base-utils/gosub"

// Create a selection collection
collection := gosub.NewSelectCollection()

// Register selectors
collection.Register(gosub.NewContextSelection(ctx, onContextDone))
collection.Register(gosub.NewSignalSelection(sigChan, onSignal))
collection.Register(gosub.NewTimerSelection(timer, onTimer))

// Initialize and start selecting
collection.Initialize()
collection.Select()
```

#### Key Types

- `SelectCollection` - Manages multiple selectors
- `SelectionConfig` - Configuration for individual selectors
- `OnSelect` - Callback function for selection events
- `SelectorIdentifier` - Unique identifier for selectors

---

### 4. Registry Package

**Import:** `github.com/grinps/go-utils/base-utils/registry`

Generic, type-safe registry for storing and retrieving values with comparable keys.

#### Features
- Generic type support
- Thread-safe operations
- Comparable key types
- Custom key interfaces
- Automatic type conversion

#### Quick Example

```go
import "github.com/grinps/go-utils/base-utils/registry"

// Create a registry
reg := registry.NewComparableRegistry[string, *MyService]()

// Register a value
reg.Register("myService", serviceInstance)

// Retrieve a value
service := reg.Get("myService")

// Unregister a value
old := reg.Unregister("myService")
```

#### Key Interfaces

- `Register[Key, Value]` - Generic registry interface
- `CustomKey[T]` - Custom key implementation
- `CustomKeyAsFunction[T]` - Function-based custom key

---

### 5. IOUtils Package

**Import:** `github.com/grinps/go-utils/base-utils/ioutils`

I/O utilities for source resolution and capability detection.

#### Features
- Source capability detection
- Source type identification
- Resolver pattern for I/O sources
- Context-aware operations

#### Quick Example

```go
import "github.com/grinps/go-utils/base-utils/ioutils"

// Check if source supports a capability
if source.Supports(ctx, ioutils.CapabilityRead) {
    // Perform read operation
}

// Resolve a source
resolver := ioutils.NewResolver()
source, err := resolver.Resolve(ctx, "file:///path/to/file")
```

#### Key Types

- `Source` - Base source interface
- `SourceCapability` - Capability enumeration
- `SourceType` - Source type enumeration
- `Resolver` - Source resolution interface

---

### 6. Logger Package

**Import:** `github.com/grinps/go-utils/logger`

A lightweight extension for Go 1.21+ `log/slog` that adds hierarchical method tracing (Markers) and context-aware logger propagation.

#### Features
- **Hierarchical Markers**: Track execution paths (e.g., `Service.Repo.Query(id=1)`) automatically.
- **Context Integration**: Store and retrieve `slog.Logger` from `context.Context`.
- **Zero-Allocation Tracing**: Optimized, lazy-evaluated markers and conditional trace logging.
- **Standard Library Compatible**: Fully compatible with `log/slog`.

#### Quick Example

```go
import (
    "context"
    "github.com/grinps/go-utils/logger"
)

func ProcessOrder(ctx context.Context, orderID string) {
    // Auto-logs "Entering ProcessOrder" at Trace level (-8)
    // Updates ctx with marker "Parent.ProcessOrder(orderID)"
    ctx = logger.Entering(ctx, "ProcessOrder", "orderID", orderID)
    defer logger.Exiting(ctx)

    // Get the logger with the current marker attached
    log := logger.LoggerFromContext(ctx)
    log.Info("Validating order...") 
}
```

#### Key Functions
- `Entering(ctx, method, args...)`: Starts a span, updates context marker & logger.
- `Exiting(ctx)`: Logs exit message (if enabled).
- `LoggerFromContext(ctx)`: Retrieves the scoped `*slog.Logger`.
- `NewMarker(name)`: Creates a standalone marker.

---

### 7. Core Utilities

**Import:** `github.com/grinps/go-utils/base-utils`

Core utility interfaces and types used across the library.

#### Comparison Utilities

```go
// Equality interface
type Equality interface {
    Equals(targetObject Equality) bool
}

// Comparable interface
type Comparable interface {
    Compare(targetObject Comparable) CompareResult
}

// CompareResult values
const (
    Less          CompareResult = -1
    Equals        CompareResult = 0
    Greater       CompareResult = 1
    NotApplicable CompareResult = 10
)
```

#### String Utilities

```go
// StringCollector - Efficient string building
type StringCollector interface {
    fmt.Stringer
    io.Writer
    io.ByteWriter
    io.StringWriter
}

// StringSecure - Secure string handling
type StringSecure interface {
    DestroyE() (bool, error)
    fmt.Stringer
    fmt.GoStringer
}
```

---

### 8. Errext Package

**Import:** `github.com/grinps/go-utils/errext`

Extended error handling with error codes, types, and structured attributes.

#### Features
- **Error Codes & Types** - Assign integer codes and string types to errors
- **Structured Attributes** - Add slog-style key-value pairs for error context
- **Stack Traces** - Optional stack trace capture for debugging
- **Stdlib Compatibility** - Fully implements `error`, `errors.Is`, `errors.As`, `errors.Unwrap`
- **Panic Recovery** - Utilities to safely handle panics

#### Quick Example

```go
import "github.com/grinps/go-utils/errext"

// Define a unique error code
var ErrInvalidInput = errext.NewErrorCode(1001)

func Process(val string) error {
    if val == "" {
        // Create error instance
        return ErrInvalidInput.New("input cannot be empty")
    }
    return nil
}

// Enable stack traces (optional)
func init() {
    errext.EnableStackTrace = true
}
```

---

### 9. Config Package

**Import:** `github.com/grinps/go-utils/config`

A flexible, context-aware configuration management library supporting nested maps with dual retrieval APIs.

#### Features
- **Context Aware**: All configuration operations accept `context.Context`.
- **Dual Retrieval APIs**: Direct `GetValue(ctx, key) (any, error)` and type-safe `GetValueE[T](ctx, key, *T) error`.
- **Type-Safe Retrieval**: Generic `GetValueE[T]` function with compile-time type safety and pointer-based assignment.
- **Default Value Pattern**: GetValueE preserves existing values when keys are not found.
- **Dot-Notation Keys**: Access nested values using dot notation (e.g., `server.port`).
- **Nested Configurations**: Retrieve sub-configurations via `GetConfig(ctx, key)`.
- **Mutable Configurations**: `MutableConfig` interface for setting values with `SetValue(ctx, key, value)`.
- **Marshable Configurations**: `MarshableConfig` interface for unmarshalling into structs.
- **Simple In-Memory Implementation**: `SimpleConfig` implements all interfaces for easy testing.
- **Structured Error Handling**: Uses `errext` package for rich error information.
- **High Test Coverage**: >93% test coverage.

#### Core Interfaces

```go
// Config - Basic configuration retrieval
type Config interface {
    GetValue(ctx context.Context, key string) (any, error)
    GetConfig(ctx context.Context, key string) (Config, error)
}

// MutableConfig - Supports setting values
type MutableConfig interface {
    Config
    SetValue(ctx context.Context, key string, value any) error
}

// MarshableConfig - Supports unmarshalling into structs
type MarshableConfig interface {
    Config
    Unmarshal(ctx context.Context, key string, target any, options ...any) error
}
```

#### Quick Example

```go
import (
    "context"
    "github.com/grinps/go-utils/config"
    "log"
)

func main() {
    ctx := context.Background()
    data := map[string]any{
        "server": map[string]any{
            "host": "localhost",
            "port": 8080,
        },
    }
    cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
    
    // Method 1: Direct GetValue - returns (any, error)
    val, err := cfg.GetValue(ctx, "server.port")
    if err != nil {
        log.Fatal(err)
    }
    port := val.(int) // Type assertion required
    
    // Method 2: Type-safe GetValueE (recommended)
    var host string
    err = config.GetValueE(ctx, "server.host", &host)
    if err != nil {
        log.Fatal(err)
    }
    
    // Method 3: Default value pattern
    timeout := 30 // default
    config.GetValueE(ctx, "server.timeout", &timeout)
    // timeout remains 30 if key doesn't exist
    
    // Store config in context
    ctx = config.ContextWithConfig(ctx, cfg)
    
    // Retrieve from context
    cfg = config.ContextConfig(ctx, true)
}
```

#### Setting Values

```go
// SimpleConfig implements MutableConfig
cfg := config.NewSimpleConfig(ctx)

// Set nested values (creates intermediate maps automatically)
err := cfg.SetValue(ctx, "server.port", 9090)

// Or use package-level function with context
ctx = config.ContextWithConfig(ctx, cfg)
err = config.SetValue(ctx, "server.host", "0.0.0.0")
```

#### Unmarshalling into Structs

```go
type ServerConfig struct {
    Host string `config:"host"`
    Port int    `config:"port"`
}

// SimpleConfig implements MarshableConfig
var server ServerConfig
err := cfg.Unmarshal(ctx, "server", &server)

// Or use package-level function with context
err = config.Unmarshal(ctx, "server", &server)
```

#### Key Functions
- `NewSimpleConfig(ctx, opts...)` - Creates in-memory config
- `GetValueE[T](ctx, key, *T)` - Type-safe retrieval from context
- `SetValue(ctx, key, value)` - Sets value in context config
- `Unmarshal[T](ctx, key, *T)` - Unmarshals into struct from context
- `ContextWithConfig(ctx, cfg)` - Stores config in context
- `ContextConfig(ctx, useDefault)` - Retrieves config from context

---

### 10. Config Ext Package

**Import:** `github.com/grinps/go-utils/config/ext`

Extended configuration utilities that build upon the base `config` package.

#### Features
- **ConfigWrapper**: Wraps any `config.Config` to provide `MarshableConfig` and `MutableConfig` capabilities with mapstructure fallback.
- **Mapstructure Fallback**: Automatic fallback to mapstructure for configs that don't natively support unmarshalling.
- **Flexible Options**: Customizable unmarshalling via functional options (tag names, strict mode, decode hooks).
- **Type Conversions**: Automatic string-to-duration, string-to-slice, and weak type conversions.
- **High Test Coverage**: >96% test coverage.

#### Quick Example

```go
import (
    "context"
    "github.com/grinps/go-utils/config"
    "github.com/grinps/go-utils/config/ext"
)

type ServerConfig struct {
    Host string `config:"host"`
    Port int    `config:"port"`
}

func main() {
    ctx := context.Background()
    data := map[string]any{
        "server": map[string]any{"host": "localhost", "port": 8080},
    }
    cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))
    
    // Wrap config to get consistent unmarshalling
    wrapper := ext.NewConfigWrapper(cfg)
    
    // Unmarshal using mapstructure fallback
    var server ServerConfig
    if err := wrapper.Unmarshal(ctx, "server", &server); err != nil {
        log.Fatal(err)
    }
    
    // SetValue if config supports it
    if wrapper.IsMutable() {
        wrapper.SetValue(ctx, "server.port", 9090)
    }
}
```

#### Key Functions
- `NewConfigWrapper(cfg)` - Wraps config with mapstructure fallback
- `wrapper.Unmarshal(ctx, key, target, opts...)` - Unmarshals config into struct
- `wrapper.SetValue(ctx, key, value)` - Sets value if config is mutable
- `wrapper.IsMutable()` - Checks if config supports mutation
- `wrapper.IsMarshable()` - Checks if config has native unmarshal support

---

## Testing

All packages include comprehensive test coverage:

```bash
# Test all packages
go test ./...

# Test with coverage
go test -cover ./...

# Test specific package
go test ./platform/...
```

### Test Coverage

- **Platform Package**: 93.2%
- **Config Package**: 93.3%
- **Config Ext Package**: 96.4%
- **System Package**: Comprehensive unit tests
- **GoSub Package**: Selection and event tests
- **Registry Package**: Generic type tests
- **IOUtils Package**: Resolver tests
- **Logger Package**: Configuration tests

---

## Design Principles

### 1. **Interface-Driven Design**
All major components are defined as interfaces, allowing for easy mocking and alternative implementations.

### 2. **Context-Aware**
Operations accept `context.Context` for cancellation, timeouts, and request-scoped values.

### 3. **Type Safety**
Extensive use of Go generics where applicable for compile-time type safety.

### 4. **Testability**
Mock implementations provided for all major interfaces, enabling comprehensive unit testing.

### 5. **Zero External Dependencies**
Core packages minimize external dependencies, relying primarily on the Go standard library.



## Contributing

Contributions are welcome! Please ensure:

1. All tests pass: `go test ./...`
2. Code is formatted: `go fmt ./...`
3. Linting passes: `golangci-lint run`
4. Test coverage remains above 80%
5. Documentation is updated

---

## Requirements

- Go 1.25 or higher (for generics support)
- No external dependencies for core packages

---

## License

This library is part of the go-utils project.

---

## Support

For issues, questions, or contributions, please refer to the main [go-utils repository](https://github.com/grinps/go-utils).

---

## Module Documentation

Each package has comprehensive Go documentation available on pkg.go.dev:

| Package | Description | Documentation |
|---------|-------------|---------------|
| **platform** | OS abstraction layer (see expandable section above) | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/platform.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/platform) |
| **system** | Service registry and dependency injection | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/system.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/system) |
| **gosub** | Channel selection utilities | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/gosub.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/gosub) |
| **registry** | Generic registry implementation | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/registry.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/registry) |
| **ioutils** | I/O utilities and resolvers | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/ioutils.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/ioutils) |
| **logs** | Logging utilities | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/logs.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/logs) |
| **base-utils** | Core utilities | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils) |
| **errext** | Extended error handling | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/errext.svg)](https://pkg.go.dev/github.com/grinps/go-utils/errext) |
| **config** | Configuration management | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/config.svg)](https://pkg.go.dev/github.com/grinps/go-utils/config) |
| **config/ext** | Config extensions (ConfigWrapper) | [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/config/ext.svg)](https://pkg.go.dev/github.com/grinps/go-utils/config/ext) |

---

## Quick Reference

| Package | Purpose | Key Interface |
|---------|---------|---------------|
| `platform` | OS abstraction | `Platform` |
| `system` | Dependency injection | `RegistrySystem` |
| `gosub` | Channel selection | `SelectCollection` |
| `registry` | Generic registry | `Register[K,V]` |
| `ioutils` | I/O utilities | `Source` |
| `logs` | Logging | `Log()`, `Warn()` |
| `base_utils` | Core utilities | `Equality`, `Comparable` |
| `errext` | Error handling | `ErrorCode`, `Error` |
| `config` | Configuration | `Config`, `MutableConfig`, `MarshableConfig` |
| `config/ext` | Config extensions |  `ConfigWrapper` |

---

## Changelog

### Platform Package

#### Current (November 2025)
- ✅ **Complete platform abstraction layer** - OS-level operations abstraction
- ✅ **Signal Handling** - Full OS signal operations support (Notify, Stop, Ignore, Reset)
- ✅ **Environment Variables** - Complete env var management (Get, Set, Unset, Lookup, Expand)
- ✅ **File Operations** - Comprehensive file system abstraction (Read, Write, Stat, Mkdir, etc.)
- ✅ **Process Operations** - Process info and control (PID, UID, GID, Hostname, FindProcess)
- ✅ **Clock/Time** - Time operations with mock support (Now, Sleep, Timer, Ticker)
- ✅ **Mock Implementations** - Full mock platform for testing with controllable behavior
- ✅ **Test Coverage** - Achieved 93.2% test coverage with comprehensive test suite
- ✅ **Documentation** - Complete API documentation, examples, and testing patterns

**Go Documentation:** [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/platform.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/platform)

---

### System Package

#### [v0.1.0](https://github.com/grinps/go-utils/releases/tag/base-utils/system/v0.1.0) (December 2023)
- ✅ **Initial Release** - System Register & Get functionality
- ✅ **Service Registry** - Type-safe service registration and retrieval
- ✅ **Context Support** - Context-aware operations throughout
- ✅ **Get Options** - Configurable service retrieval with option pattern
- ✅ **Registration Options** - Flexible service registration with extensibility
- ✅ **RegistrySystem Interface** - Core interface for dependency injection

**Go Documentation:** [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/system.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/system)

---

### GoSub Package

#### [v0.1.0](https://github.com/grinps/go-utils/releases/tag/base-utils/gosub/v0.1.0) (December 2022)
- ✅ **Initial Release** - Common capability to support Go routine sync patterns
- ✅ **Multi-Channel Selection** - SelectCollection for managing multiple channels
- ✅ **Context Selection** - Context-based monitoring and cancellation
- ✅ **Signal Selection** - OS signal monitoring integration
- ✅ **Timer Selection** - Timer-based event handling
- ✅ **Proxy Channels** - Channel proxying and forwarding support
- ✅ **Event Callbacks** - Flexible OnSelect callback pattern

**Go Documentation:** [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/gosub.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/gosub)

---

### Registry Package

#### [v0.6.0](https://github.com/grinps/go-utils/releases/tag/base-utils/registry/v0.6.0) (December 2023)
- 🐛 **Bugfix** - Missed registry updates

#### [v0.5.0](https://github.com/grinps/go-utils/releases/tag/base-utils/registry/v0.5.0) (January 2023)
- ✅ **Any Key Support** - Support for any Key type instead of comparable constraint
- ✅ **Register Interface** - Defined Register as interface for extensibility
- ✅ **CustomKey Interface** - Support for extensible and comparable keys
- ✅ **NewRegisterWithAnyKey** - Helper for creating registry with any Key type
- 🔧 **Refactoring** - Removed reference to Key from registrationRecord

#### [v0.4.0](https://github.com/grinps/go-utils/releases/tag/base-utils/registry/v0.4.0) (January 2023)
- ✅ **CustomKey Support** - Added support for complex key comparison

#### [v0.3.0](https://github.com/grinps/go-utils/releases/tag/base-utils/registry/v0.3.0) (December 2022)
- 🔧 **Dependency Update** - Upgraded base-utils/logs package

#### [v0.2.0](https://github.com/grinps/go-utils/releases/tag/base-utils/registry/v0.2.0) (October 2022)
- ✅ **Generic Value** - Added support for Value as generic parameter

#### [v0.1.0](https://github.com/grinps/go-utils/releases/tag/base-utils/registry/v0.1.0) (October 2022)
- ✅ **Initial Release** - Generic registry implementation with comparable keys

**Go Documentation:** [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/registry.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/registry)

### IOUtils Package

#### Current
- ✅ **Source Interface** - Base source abstraction for I/O operations
- ✅ **Capability Detection** - Source capability checking via Supports method
- ✅ **Source Types** - Type enumeration for different source types
- ✅ **Resolver Pattern** - Source resolution support
- ✅ **Context Support** - Context-aware operations

**Go Documentation:** [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils/ioutils.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils/ioutils)

---

### Core Utilities (base-utils)

#### [v0.2.0](https://github.com/grinps/go-utils/releases/tag/base-utils/v0.2.0) (December 2022)
- 📄 **License** - Added LICENSE file

#### [v0.1.0](https://github.com/grinps/go-utils/releases/tag/base-utils/v0.1.0) (January 2022)
- ✅ **Initial Release** - Migrated logger function to base util with test cases
- ✅ **Equality Interface** - Object equality comparison
- ✅ **Comparable Interface** - Ordered comparison with CompareResult
- ✅ **StringCollector** - Efficient string building interface
- ✅ **StringSecure** - Secure string handling abstractions

**Go Documentation:** [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/base-utils.svg)](https://pkg.go.dev/github.com/grinps/go-utils/base-utils)

---

### Errext Package

#### [v0.8.0](https://github.com/grinps/go-utils/releases/tag/errext/v0.8.0)
- ✅ **Initial Release** - Structured error handling
- ✅ **Error Codes & Types** - Integer codes and string categorization
- ✅ **Structured Attributes** - slog-style key-value pairs for error context
- ✅ **Stack Traces** - Optional stack capture
- ✅ **Stdlib Compatibility** - `errors.Is` and `errors.As` support

**Go Documentation:** [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/errext.svg)](https://pkg.go.dev/github.com/grinps/go-utils/errext)

---

### Config Package

#### [v0.2.0](https://github.com/grinps/go-utils/releases/tag/config/v0.2.0) (November 2025)
- ✅ **Initial Release** - Flexible, context-aware configuration management
- ✅ **Context Aware** - All configuration operations accept `context.Context`.
- ✅ **Type-Safe Retrieval** - Generic `GetValueE[T]` functions.
- ✅ **Context-Based Functions** - `Unmarshal` and `SetValue` extract config from context automatically
- ✅ **MutableConfig Interface** - Defines `SetValue` for modifying configuration
- ✅ **MarshableConfig Interface** - Defines `Unmarshal` for struct unmarshalling
- ✅ **Dot-Notation Keys** - Access nested values using dot notation (e.g., `server.port`).
- ✅ **Simple In-Memory Implementation** - Includes `SimpleConfig` for easy testing and mocking.
- ✅ **Structured Error Handling** - Uses `errext` package for rich error information.

**Go Documentation:** [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/config.svg)](https://pkg.go.dev/github.com/grinps/go-utils/config)

---

### Config Ext Package

#### [v0.2.0](https://github.com/grinps/go-utils/releases/tag/config/ext/v0.2.0) (November 2025)
- ✅ **ConfigWrapper** - Wraps any `config.Config` with `MarshableConfig` and `MutableConfig` capabilities
- ✅ **Flexible Unmarshal Options** - Tag names (`json`, `yaml`, `mapstructure`), strict mode, decode hooks
- ✅ **Mapstructure Fallback** - `ConfigWrapper` provides mapstructure-based unmarshalling for any config
- ✅ **Type Conversions** - String-to-duration, string-to-slice, weak type conversions
- ✅ **High Test Coverage** - >96% test coverage

**Go Documentation:** [![Go Reference](https://pkg.go.dev/badge/github.com/grinps/go-utils/config/ext.svg)](https://pkg.go.dev/github.com/grinps/go-utils/config/ext)

---

**Version:** 1.0.0  
**Last Updated:** November 2025


