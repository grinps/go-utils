# Config Ext Package

The `ext` package provides extended configuration utilities that build upon the base `config` package.

## Features

- **ConfigWrapper**: Wraps any `config.Config` to provide consistent `MarshableConfig` and `MutableConfig` capabilities
- **Mapstructure Fallback**: Automatic fallback to mapstructure for configs that don't natively support unmarshalling
- **Flexible Options**: Customizable unmarshalling behavior via functional options
- **Type Conversions**: Automatic conversion of strings to durations, slices, and more
- **Telemetry Support**: Implements `config.TelemetryAware` for telemetry integration
- **High Test Coverage**: >96% test coverage

## Installation

```bash
go get github.com/grinps/go-utils/config/ext
```

## Usage

### Basic Usage with ConfigWrapper

The `ConfigWrapper` wraps any `config.Config` and provides consistent access to `MarshableConfig` and `MutableConfig` capabilities with mapstructure fallback:

```go
cfg := config.NewSimpleConfig(ctx, config.WithConfigurationMap(data))

// Wrap the config
wrapper := ext.NewConfigWrapper(cfg)

// Use Unmarshal consistently (uses mapstructure if MarshableConfig not implemented)
var server ServerConfig
if err := wrapper.Unmarshal(ctx, "server", &server); err != nil {
    log.Fatal(err)
}

// Check capabilities
if wrapper.IsMutable() {
    wrapper.SetValue(ctx, "server.port", 9090)
}

// Access underlying config
original := wrapper.Unwrap()
```

### Using Different Struct Tags

```go
// Use JSON tags
type Config struct {
    Host string `json:"host"`
    Port int    `json:"port"`
}

var cfg Config
err := wrapper.Unmarshal(ctx, "server", &cfg, ext.WithJSONTag())

// Or use mapstructure tags (viper-style)
err := wrapper.Unmarshal(ctx, "server", &cfg, ext.WithMapstructureTag())

// Or use YAML tags
err := wrapper.Unmarshal(ctx, "server", &cfg, ext.WithYAMLTag())

// Or specify any tag name
err := wrapper.Unmarshal(ctx, "server", &cfg, ext.WithTagName("custom"))
```

### Strict Mode

```go
// Error if config has unused keys or struct has unset fields
err := wrapper.Unmarshal(ctx, "server", &server, ext.WithStrictMode())
```

### Custom Decode Hooks

```go
import "github.com/go-viper/mapstructure/v2"

// Custom hook to parse custom types
hook := func(from reflect.Type, to reflect.Type, data any) (any, error) {
    if to == reflect.TypeOf(MyCustomType{}) {
        // Custom conversion logic
        return parseMyCustomType(data)
    }
    return data, nil
}

err := wrapper.Unmarshal(ctx, "server", &server, ext.WithDecodeHook(hook))
```

## Interfaces

### MutableConfig

Defines the ability to set configuration values:

```go
type MutableConfig interface {
    SetValue(ctx context.Context, key string, newValue any) error
}
```

### MarshableConfig

Provides struct unmarshalling capabilities:

```go
type MarshableConfig interface {
    Unmarshal(ctx context.Context, key string, target any, options ...UnmarshalOption) error
}
```

### ConfigWrapper

Wraps any `config.Config` and implements all standard interfaces with passthrough support:

```go
type ConfigWrapper struct {
    // ...
}

// Implements config.Config
func (w *ConfigWrapper) Name() config.ProviderName
func (w *ConfigWrapper) GetValue(ctx, key) (any, error)
func (w *ConfigWrapper) GetConfig(ctx, key) (config.Config, error)

// Implements config.MarshableConfig (uses mapstructure fallback if needed)
func (w *ConfigWrapper) Unmarshal(ctx, key, target, options...) error

// Implements config.MutableConfig (returns error if wrapped config doesn't support it)
func (w *ConfigWrapper) SetValue(ctx, key, newValue) error

// Implements config.AllGetter (passthrough if supported)
func (w *ConfigWrapper) All(ctx) map[string]any

// Implements config.AllKeysProvider (passthrough if supported)
func (w *ConfigWrapper) Keys(prefix string) []string

// Implements config.Deleter (passthrough if supported)
func (w *ConfigWrapper) Delete(key string) error

// Utility methods
func (w *ConfigWrapper) Unwrap() config.Config
func (w *ConfigWrapper) IsMutable() bool
func (w *ConfigWrapper) IsMarshable() bool
func (w *ConfigWrapper) HasAllGetter() bool
func (w *ConfigWrapper) HasAllKeys() bool
func (w *ConfigWrapper) HasDeleter() bool

// Implements config.TelemetryAware
func (w *ConfigWrapper) ShouldInstrument(ctx, key, op) bool
func (w *ConfigWrapper) GenerateTelemetryAttributes(ctx, op, attrs) []any
```

### Passthrough Behavior

ConfigWrapper automatically detects and passes through calls to optional interfaces:

```go
wrapper := ext.NewConfigWrapper(cfg)

// Keys() passes through if wrapped config implements AllKeysProvider
if wrapper.HasAllKeys() {
    keys := wrapper.Keys("server")
}

// Delete() passes through if wrapped config implements Deleter
if wrapper.HasDeleter() {
    err := wrapper.Delete("server.debug")
}

// All() passes through if wrapped config implements AllGetter
if wrapper.HasAllGetter() {
    all := wrapper.All(ctx)
}
```

## Options Reference

| Option | Description | Default |
|--------|-------------|---------|
| `WithTagName(name)` | Struct tag for field mapping | `"config"` |
| `WithJSONTag()` | Use `json` struct tag | - |
| `WithYAMLTag()` | Use `yaml` struct tag | - |
| `WithMapstructureTag()` | Use `mapstructure` struct tag | - |
| `WithWeaklyTypedInput(bool)` | Enable weak type conversions | `true` |
| `WithSquash(bool)` | Enable embedded struct flattening | `true` |
| `WithErrorUnused(bool)` | Error on unused config keys | `false` |
| `WithErrorUnset(bool)` | Error on unset struct fields | `false` |
| `WithStrictMode()` | Enable both unused and unset errors | - |
| `WithDecodeHook(hook)` | Add custom decode hook | - |
| `WithMetadata(*Metadata)` | Store decode metadata | - |

## Default Type Conversions

The following conversions are enabled by default:

- **String to Duration**: `"30s"` → `30 * time.Second`
- **String to Slice**: `"a,b,c"` → `[]string{"a", "b", "c"}`
- **Weak typing** (when enabled): `"8080"` → `8080`

## Testing

```bash
cd config/ext
go test -v ./...
```

## License

Part of the go-utils project.
