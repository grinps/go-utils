# Koanf Config Package

The `config/koanf` package provides a wrapper around [knadh/koanf](https://github.com/knadh/koanf) that implements the `config.Config`, `config.MutableConfig`, and `config.MarshableConfig` interfaces.

This package bridges the powerful koanf configuration library with the standardized config interfaces, enabling seamless integration with the broader config ecosystem while leveraging koanf's extensive provider and parser support.

## Features

- **Standard Interface Implementation**: Implements `config.Config`, `config.MutableConfig`, and `config.MarshableConfig`
- **Telemetry Support**: Implements `config.TelemetryAware` for telemetry integration
- **Multiple Configuration Sources**: Files, environment variables, command-line flags, S3, Consul, Vault, and more
- **Nested Configuration**: Access nested values using dot-notation keys (customizable delimiter)
- **Type-Safe Unmarshalling**: Unmarshal to structs with support for multiple tag formats (koanf, json, yaml, mapstructure)
- **Mutable Configuration**: Set and modify values at runtime
- **Provider-Based Loading**: Load from various sources using koanf's provider system
- **Configuration Merging**: Merge multiple configurations with override support
- **Structured Error Handling**: Uses `errext` package for rich error information
- **High Test Coverage**: >94% test coverage with comprehensive edge case handling

## Installation

```bash
# Install the koanf wrapper
go get github.com/grinps/go-utils/config/koanf

# Install providers and parsers as needed
go get github.com/knadh/koanf/providers/file
go get github.com/knadh/koanf/parsers/json
go get github.com/knadh/koanf/parsers/yaml
```

## Quick Start

### Basic Usage

```go
import (
    "context"
    "github.com/grinps/go-utils/config/koanf"
)

func main() {
    ctx := context.Background()
    
    // Create empty config
    cfg := koanf.NewKoanfConfig(ctx)
    
    // Set values
    kcfg := cfg.(*koanf.KoanfConfig)
    kcfg.SetValue(ctx, "server.port", 8080)
    kcfg.SetValue(ctx, "server.host", "localhost")
    
    // Get values
    port, err := cfg.GetValue(ctx, "server.port")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("Port:", port)
}
```

### Loading from Files

```go
import (
    "github.com/knadh/koanf/parsers/json"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/file"
)

// Load from JSON file during creation
cfg := koanf.NewKoanfConfig(ctx,
    koanf.WithProvider(file.Provider("config.json"), json.Parser()),
)

// Or load after creation
cfg := koanf.NewKoanfConfig(ctx)
kcfg := cfg.(*koanf.KoanfConfig)
err := kcfg.Load(ctx, file.Provider("config.yaml"), yaml.Parser())
```

### Loading from Environment Variables

```go
import "github.com/knadh/koanf/providers/env"

cfg := koanf.NewKoanfConfig(ctx)
kcfg := cfg.(*koanf.KoanfConfig)

// Load environment variables with prefix "APP_"
// e.g., APP_SERVER_PORT=8080 becomes server.port=8080
err := kcfg.Load(ctx, env.Provider("APP_", ".", nil), nil)
```

### Unmarshalling to Structs

```go
type ServerConfig struct {
    Host string `koanf:"host"`
    Port int    `koanf:"port"`
}

type DatabaseConfig struct {
    Host     string `koanf:"host"`
    Port     int    `koanf:"port"`
    Username string `koanf:"username"`
}

type AppConfig struct {
    Server   ServerConfig   `koanf:"server"`
    Database DatabaseConfig `koanf:"database"`
}

// Unmarshal entire config
var app AppConfig
err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "", &app)

// Unmarshal sub-config
var server ServerConfig
err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "server", &server)
```

### Using Different Tag Formats

```go
// JSON tags
type JSONConfig struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}
var jsonCfg JSONConfig
err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "server", &jsonCfg, koanf.WithJSONTag())

// YAML tags
type YAMLConfig struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}
var yamlCfg YAMLConfig
err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "server", &yamlCfg, koanf.WithYAMLTag())

// Mapstructure tags (viper-compatible)
type ViperConfig struct {
    Host string `mapstructure:"host"`
    Port int    `mapstructure:"port"`
}
var viperCfg ViperConfig
err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "server", &viperCfg, koanf.WithMapstructureTag())
```

## Advanced Usage

### Merging Configurations

```go
// Load base configuration
baseCfg := koanf.NewKoanfConfig(ctx,
    koanf.WithProvider(file.Provider("base.json"), json.Parser()),
)

// Load environment-specific overrides
overrideCfg := koanf.NewKoanfConfig(ctx,
    koanf.WithProvider(file.Provider("prod.json"), json.Parser()),
)

// Merge override into base (override values take precedence)
err := baseCfg.(*koanf.KoanfConfig).Merge(ctx, overrideCfg.(*koanf.KoanfConfig))
```

### Custom Delimiters

```go
// Use "/" instead of "." for key paths
cfg := koanf.NewKoanfConfig(ctx, koanf.WithDelimiter("/"))

cfg.(*koanf.KoanfConfig).SetValue(ctx, "server/port", 8080)
val, err := cfg.GetValue(ctx, "server/port")
```

### Flat Path Unmarshalling

```go
type FlatConfig struct {
    ServerPort int    `koanf:"server.port"`
    ServerHost string `koanf:"server.host"`
    DBHost     string `koanf:"database.host"`
}

var flat FlatConfig
err := cfg.(*koanf.KoanfConfig).Unmarshal(ctx, "", &flat, koanf.WithFlatPaths(true))
```

### Using Existing Koanf Instance

```go
import "github.com/knadh/koanf/v2"

// Create and configure koanf instance directly
k := koanf.New(".")
// ... configure k ...

// Wrap it with KoanfConfig
cfg := koanf.FromKoanf(k)
```

## Configuration Options

### Creation Options

- `WithDelimiter(delimiter string)`: Set custom key delimiter (default: ".")
- `WithKoanfInstance(k *koanf.Koanf)`: Use existing koanf instance
- `WithProvider(provider, parser)`: Load from provider during creation

### Unmarshal Options

- `WithTag(tagName string)`: Set struct tag name for field mapping
- `WithKoanfTag()`: Use "koanf" tag (default)
- `WithJSONTag()`: Use "json" tag
- `WithYAMLTag()`: Use "yaml" tag
- `WithMapstructureTag()`: Use "mapstructure" tag (viper-compatible)
- `WithFlatPaths(enabled bool)`: Enable flat path access for nested structures

## Available Providers

Koanf supports many providers (install separately with `go get github.com/knadh/koanf/providers/$provider`):

- **file**: Load from files
- **env**: Load from environment variables
- **confmap**: Load from Go maps
- **structs**: Load from Go structs
- **rawbytes**: Load from raw bytes
- **s3**: Load from AWS S3
- **vault**: Load from HashiCorp Vault
- **consul**: Load from Consul
- **etcd**: Load from etcd
- **parameterstore**: Load from AWS Parameter Store
- **appconfig**: Load from AWS AppConfig
- **basicflag**: Load from Go flag.FlagSet
- **posflag**: Load from spf13/pflag

## Available Parsers

Koanf supports many parsers (install separately with `go get github.com/knadh/koanf/parsers/$parser`):

- **json**: JSON parser
- **yaml**: YAML parser
- **toml**: TOML parser (v1 and v2)
- **hcl**: HCL parser
- **hjson**: HJSON parser
- **dotenv**: .env file parser
- **nestedtext**: NestedText parser

## Error Handling

The package uses the `errext` package for structured error handling:

```go
import "github.com/grinps/go-utils/errext"

val, err := cfg.GetValue(ctx, "nonexistent.key")
if err != nil {
    if errext.IsErrorCode(err, koanf.ErrKoanfMissingValue) {
        // Handle missing value
        log.Println("Key not found, using default")
    } else {
        log.Fatal(err)
    }
}
```

### Error Codes

- `ErrKoanfNilConfig`: Nil koanf config encountered
- `ErrKoanfEmptyKey`: Empty key provided
- `ErrKoanfMissingValue`: Configuration value not found
- `ErrKoanfInvalidValue`: Invalid or unconvertible value
- `ErrKoanfSetValueFailed`: Setting value failed
- `ErrKoanfUnmarshalFailed`: Unmarshalling to struct failed
- `ErrKoanfInvalidTarget`: Invalid target for unmarshalling
- `ErrKoanfLoadFailed`: Loading from provider failed
- `ErrKoanfMergeFailed`: Merging configurations failed
- `ErrKoanfInvalidProvider`: Invalid provider provided

## Interface Compatibility

KoanfConfig implements all standard config interfaces:

```go
import "github.com/grinps/go-utils/config"

// As Config
var cfg config.Config = koanf.NewKoanfConfig(ctx)

```

## Thread Safety

KoanfConfig is safe for concurrent reads but not for concurrent writes. If you need to modify configuration from multiple goroutines, use external synchronization (e.g., `sync.RWMutex`).

## Examples

### Complete Application Configuration

```go
package main

import (
    "context"
    "log"
    
    "github.com/grinps/go-utils/config/koanf"
    "github.com/knadh/koanf/parsers/json"
    "github.com/knadh/koanf/parsers/yaml"
    "github.com/knadh/koanf/providers/env"
    "github.com/knadh/koanf/providers/file"
)

type Config struct {
    Server struct {
        Host string `koanf:"host"`
        Port int    `koanf:"port"`
    } `koanf:"server"`
    
    Database struct {
        Host     string `koanf:"host"`
        Port     int    `koanf:"port"`
        Username string `koanf:"username"`
        Password string `koanf:"password"`
    } `koanf:"database"`
}

func main() {
    ctx := context.Background()
    
    // Create config and load from multiple sources
    cfg := koanf.NewKoanfConfig(ctx)
    kcfg := cfg.(*koanf.KoanfConfig)
    
    // 1. Load base config from YAML
    if err := kcfg.Load(ctx, file.Provider("config.yaml"), yaml.Parser()); err != nil {
        log.Fatal(err)
    }
    
    // 2. Load environment-specific overrides from JSON
    if err := kcfg.Load(ctx, file.Provider("config.prod.json"), json.Parser()); err != nil {
        log.Printf("No prod config: %v", err)
    }
    
    // 3. Override with environment variables (highest priority)
    if err := kcfg.Load(ctx, env.Provider("APP_", ".", nil), nil); err != nil {
        log.Fatal(err)
    }
    
    // Unmarshal to struct
    var appConfig Config
    if err := kcfg.Unmarshal(ctx, "", &appConfig); err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Server: %s:%d", appConfig.Server.Host, appConfig.Server.Port)
    log.Printf("Database: %s:%d", appConfig.Database.Host, appConfig.Database.Port)
}
```

## Testing

The package includes comprehensive tests with >96% coverage:

```bash
cd config/koanf
go test -v -cover
```

## License

This package is part of the grinps/go-utils library and follows the same license.

## See Also

- [config package](../README.md) - Parent config package
- [knadh/koanf](https://github.com/knadh/koanf) - Underlying koanf library
- [config/ext package](../ext/README.md) - Config extensions with mapstructure support
