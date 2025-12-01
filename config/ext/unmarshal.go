package ext

import (
	"github.com/go-viper/mapstructure/v2"
)

// buildDecoderConfig creates a mapstructure.DecoderConfig from unmarshalConfig.
func buildDecoderConfig(target any, cfg *unmarshalConfig) *mapstructure.DecoderConfig {
	decoderConfig := &mapstructure.DecoderConfig{
		Result:           target,
		TagName:          cfg.tagName,
		WeaklyTypedInput: cfg.weaklyTypedInput,
		Squash:           cfg.squash,
		ErrorUnused:      cfg.errorUnused,
		ErrorUnset:       cfg.errorUnset,
		Metadata:         cfg.metadata,
	}

	// Add decode hooks
	if len(cfg.decodeHooks) > 0 {
		// Compose all hooks with default hooks
		hooks := append([]mapstructure.DecodeHookFunc{
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		}, cfg.decodeHooks...)
		decoderConfig.DecodeHook = mapstructure.ComposeDecodeHookFunc(hooks...)
	} else {
		// Default hooks for common conversions
		decoderConfig.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}

	return decoderConfig
}

// newDecoder creates a new mapstructure decoder with the given config.
func newDecoder(config *mapstructure.DecoderConfig) (*mapstructure.Decoder, error) {
	return mapstructure.NewDecoder(config)
}

// decodeWithMapstructure decodes a map into the target struct using mapstructure.
// This is a shared helper used by both unmarshalWithMapstructure and unmarshalAny.
func decodeWithMapstructure(configMap map[string]any, target any, options ...any) error {
	// Apply options - only process UnmarshalOption types
	unmarshalCfg := defaultUnmarshalConfig()
	for _, opt := range options {
		if fn, ok := opt.(UnmarshalOption); ok {
			fn(unmarshalCfg)
		}
	}

	// Build mapstructure decoder config
	decoderConfig := buildDecoderConfig(target, unmarshalCfg)

	decoder, err := newDecoder(decoderConfig)
	if err != nil {
		return ErrExtUnmarshalFailed.New("failed to create decoder", "error", err)
	}

	if err := decoder.Decode(configMap); err != nil {
		return ErrExtUnmarshalFailed.New("failed to decode config", "error", err)
	}

	return nil
}
