package koanf

import (
	"github.com/knadh/koanf/v2"
)

// UnmarshalOption configures the unmarshalling behavior for KoanfConfig.
type UnmarshalOption func(*koanf.UnmarshalConf)

// WithTag sets the struct tag name used for field mapping.
// Default is "koanf".
//
// Example:
//
//	type Config struct {
//	    Port int `json:"port"` // Use json tag instead
//	}
//	err := cfg.Unmarshal(ctx, "server", &config, WithTag("json"))
func WithTag(tagName string) UnmarshalOption {
	return func(c *koanf.UnmarshalConf) {
		c.Tag = tagName
	}
}

// WithFlatPaths enables flat path access for nested structures.
// When enabled, nested keys are accessed with flat paths instead of nested maps.
//
// Example:
//
//	// With flat paths: {"server.port": 8080}
//	// Without: {"server": {"port": 8080}}
func WithFlatPaths(enabled bool) UnmarshalOption {
	return func(c *koanf.UnmarshalConf) {
		c.FlatPaths = enabled
	}
}

// WithKoanfTag is a convenience option that sets the tag name to "koanf".
// This is the default tag name.
func WithKoanfTag() UnmarshalOption {
	return WithTag("koanf")
}

// WithJSONTag is a convenience option that sets the tag name to "json".
// This is useful when working with structs that use JSON tags.
func WithJSONTag() UnmarshalOption {
	return WithTag("json")
}

// WithYAMLTag is a convenience option that sets the tag name to "yaml".
// This is useful when working with structs that use YAML tags.
func WithYAMLTag() UnmarshalOption {
	return WithTag("yaml")
}

// WithMapstructureTag is a convenience option that sets the tag name to "mapstructure".
// This is useful when working with structs that use viper-style tags.
func WithMapstructureTag() UnmarshalOption {
	return WithTag("mapstructure")
}

// KoanfOption configures a KoanfConfig instance during creation.
type KoanfOption func(*KoanfConfig)

// WithDelimiter sets a custom delimiter for key parsing (default is ".").
//
// Example:
//
//	cfg := koanf.NewKoanfConfig(ctx, koanf.WithDelimiter("/"))
//	val, _ := cfg.GetValue(ctx, "server/port") // Uses / instead of .
func WithDelimiter(delimiter string) KoanfOption {
	return func(k *KoanfConfig) {
		k.delimiter = delimiter
		// Note: The koanf instance delimiter is set during creation
	}
}

// WithKoanfInstance allows using an existing koanf.Koanf instance.
// This is useful when you need to configure koanf with specific settings
// before wrapping it with KoanfConfig.
//
// Example:
//
//	k := koanf.New(".")
//	cfg := koanf.NewKoanfConfig(ctx, koanf.WithKoanfInstance(k))
func WithKoanfInstance(instance *koanf.Koanf) KoanfOption {
	return func(k *KoanfConfig) {
		k.k = instance
	}
}

// WithProvider loads configuration from a provider during initialization.
// This is a convenience option that calls Load after creating the config.
//
// Example:
//
//	import (
//	    "github.com/knadh/koanf/parsers/json"
//	    "github.com/knadh/koanf/providers/file"
//	)
//
//	cfg := koanf.NewKoanfConfig(ctx,
//	    koanf.WithProvider(file.Provider("config.json"), json.Parser()),
//	)
type ProviderOption struct {
	Provider koanf.Provider
	Parser   koanf.Parser
}

// WithProvider creates a provider option for loading configuration during initialization.
func WithProvider(provider koanf.Provider, parser koanf.Parser) ProviderOption {
	return ProviderOption{
		Provider: provider,
		Parser:   parser,
	}
}
