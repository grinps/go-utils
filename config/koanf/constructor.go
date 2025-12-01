package koanf

import (
	"context"
	"errors"

	"github.com/grinps/go-utils/config"
	"github.com/knadh/koanf/v2"
)

// NewKoanfConfig creates a new KoanfConfig instance with optional configuration.
// By default, uses "." as the key delimiter and creates an empty koanf instance.
//
// Example:
//
//	// Empty config
//	cfg, err := koanf.NewKoanfConfig(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	// With custom delimiter
//	cfg, err := koanf.NewKoanfConfig(ctx, koanf.WithDelimiter("/"))
//	if err != nil {
//		return nil, err
//	}
//
//	// With existing koanf instance
//	k := koanf.New(".")
//	cfg := koanf.NewKoanfConfig(ctx, koanf.WithKoanfInstance(k))
//
//	// With provider
//	import (
//	    "github.com/knadh/koanf/parsers/json"
//	    "github.com/knadh/koanf/providers/file"
//	)
//	cfg := koanf.NewKoanfConfig(ctx,
//	    koanf.WithProvider(file.Provider("config.json"), json.Parser()),
//	)
func NewKoanfConfig(ctx context.Context, options ...any) (config.Config, error) {
	var err error
	delimiter := config.DefaultKeyDelimiter

	// Pre-scan for delimiter option to use when creating koanf instance
	for _, opt := range options {
		if kOpt, ok := opt.(KoanfOption); ok {
			// Create a temporary config to extract delimiter
			temp := &KoanfConfig{delimiter: delimiter}
			kOpt(temp)
			delimiter = temp.delimiter
		}
	}

	cfg := &KoanfConfig{
		k:         koanf.New(delimiter),
		delimiter: delimiter,
	}

	var loadErrs []error
	// Apply all options
	for _, opt := range options {
		switch o := opt.(type) {
		case KoanfOption:
			o(cfg)
		case ProviderOption:
			// Load from provider
			if o.Provider != nil {
				loadErr := cfg.Load(ctx, o.Provider, o.Parser)
				loadErrs = append(loadErrs, loadErr)
			}
		}
	}
	if len(loadErrs) > 0 {
		err = ErrKoanfLoadFailed.NewWithError("failed to load configuration from provider", errors.Join(loadErrs...))
	}

	return cfg, err
}

// FromKoanf wraps an existing koanf.Koanf instance as a KoanfConfig.
// This is useful when you have an already configured koanf instance.
//
// Example:
//
//	k := koanf.New(".")
//	// ... configure k ...
//	cfg := koanf.FromKoanf(k)
func FromKoanf(k *koanf.Koanf) config.Config {
	if k == nil {
		return nil
	}

	return &KoanfConfig{
		k:         k,
		delimiter: k.Delim(),
	}
}
