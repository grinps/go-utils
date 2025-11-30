package ext

import (
	"github.com/go-viper/mapstructure/v2"
)

// UnmarshalOption configures the unmarshalling behavior.
type UnmarshalOption func(*unmarshalConfig)

// unmarshalConfig holds the configuration for unmarshalling operations.
type unmarshalConfig struct {
	// tagName is the struct tag to use for field mapping (default: "config")
	tagName string

	// weaklyTypedInput enables weak type conversions (e.g., string "8080" to int 8080)
	weaklyTypedInput bool

	// squash enables embedded struct flattening
	squash bool

	// errorUnused returns an error if config has keys not in the target struct
	errorUnused bool

	// errorUnset returns an error if struct fields don't have matching config keys
	errorUnset bool

	// decodeHooks are custom decode hooks for type conversions
	decodeHooks []mapstructure.DecodeHookFunc

	// metadata stores metadata about the decoding process
	metadata *mapstructure.Metadata
}

// defaultUnmarshalConfig returns the default unmarshalling configuration.
func defaultUnmarshalConfig() *unmarshalConfig {
	return &unmarshalConfig{
		tagName:          "config",
		weaklyTypedInput: true,
		squash:           true,
		errorUnused:      false,
		errorUnset:       false,
		decodeHooks:      nil,
		metadata:         nil,
	}
}

// WithTagName sets the struct tag name used for field mapping.
// Default is "config".
//
// Example:
//
//	type Config struct {
//	    Port int `json:"port"` // Use json tag instead
//	}
//	err := ext.Unmarshal(ctx, cfg, "server", &config, ext.WithTagName("json"))
func WithTagName(tagName string) UnmarshalOption {
	return func(c *unmarshalConfig) {
		c.tagName = tagName
	}
}

// WithWeaklyTypedInput enables or disables weak type conversions.
// When enabled, strings can be converted to numbers, bools, etc.
// Default is true.
//
// Example:
//
//	// Config has "port": "8080" (string)
//	// Target struct has Port int
//	// With weak typing, "8080" -> 8080 succeeds
func WithWeaklyTypedInput(enabled bool) UnmarshalOption {
	return func(c *unmarshalConfig) {
		c.weaklyTypedInput = enabled
	}
}

// WithSquash enables or disables embedded struct flattening.
// When enabled, embedded structs are flattened into the parent.
// Default is true.
//
// Example:
//
//	type BaseConfig struct {
//	    Host string `config:"host"`
//	}
//	type ServerConfig struct {
//	    BaseConfig `config:",squash"`
//	    Port int   `config:"port"`
//	}
func WithSquash(enabled bool) UnmarshalOption {
	return func(c *unmarshalConfig) {
		c.squash = enabled
	}
}

// WithErrorUnused returns an error if the config contains keys
// that don't have corresponding fields in the target struct.
// Default is false.
func WithErrorUnused(enabled bool) UnmarshalOption {
	return func(c *unmarshalConfig) {
		c.errorUnused = enabled
	}
}

// WithErrorUnset returns an error if the target struct has fields
// without corresponding keys in the config.
// Default is false.
func WithErrorUnset(enabled bool) UnmarshalOption {
	return func(c *unmarshalConfig) {
		c.errorUnset = enabled
	}
}

// WithDecodeHook adds a custom decode hook for type conversions.
// Multiple hooks can be added and are executed in order.
//
// Example:
//
//	// Custom hook to parse duration strings
//	hook := func(from reflect.Type, to reflect.Type, data any) (any, error) {
//	    if to == reflect.TypeOf(time.Duration(0)) {
//	        return time.ParseDuration(data.(string))
//	    }
//	    return data, nil
//	}
//	err := ext.Unmarshal(ctx, cfg, "server", &config, ext.WithDecodeHook(hook))
func WithDecodeHook(hook mapstructure.DecodeHookFunc) UnmarshalOption {
	return func(c *unmarshalConfig) {
		c.decodeHooks = append(c.decodeHooks, hook)
	}
}

// WithMetadata provides a pointer to store decoding metadata.
// The metadata includes information about which keys were decoded,
// which were unused, and which struct fields were unset.
func WithMetadata(metadata *mapstructure.Metadata) UnmarshalOption {
	return func(c *unmarshalConfig) {
		c.metadata = metadata
	}
}

// WithMapstructureTag is a convenience option that sets the tag name to "mapstructure".
// This is useful when working with structs that use viper-style tags.
func WithMapstructureTag() UnmarshalOption {
	return WithTagName("mapstructure")
}

// WithJSONTag is a convenience option that sets the tag name to "json".
// This is useful when working with structs that use JSON tags.
func WithJSONTag() UnmarshalOption {
	return WithTagName("json")
}

// WithYAMLTag is a convenience option that sets the tag name to "yaml".
// This is useful when working with structs that use YAML tags.
func WithYAMLTag() UnmarshalOption {
	return WithTagName("yaml")
}

// WithStrictMode enables strict unmarshalling that errors on both
// unused config keys and unset struct fields.
func WithStrictMode() UnmarshalOption {
	return func(c *unmarshalConfig) {
		c.errorUnused = true
		c.errorUnset = true
	}
}
