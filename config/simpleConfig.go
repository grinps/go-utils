package config

import (
	"context"
	"errors"
	"fmt"
)

func init() {
	Register(context.Background(), WithDriver[SimpleConfigDriver, SimpleConfig]())
}

type SimpleConfig = *simpleConfig

type simpleConfig struct {
	configurationMap  map[string]any
	defaultKeyParsers []KeyParser
	defaultGetOption  []ValueRetriever[*simpleConfig]
}

func (cfg *simpleConfig) GetValue(ctx context.Context, key string, options ...GetOption) (applicableValue any, errs error) {
	if cfg != nil && cfg.configurationMap != nil {
		keyGenerators, valueGetters := ClassifyGetOptions[*simpleConfig](options...)
		if len(keyGenerators) == 0 {
			if len(cfg.defaultKeyParsers) > 0 {
				keyGenerators = cfg.defaultKeyParsers
			} else {
				keyGenerators = []KeyParser{DefaultKeyParser(ctx)}
			}
		}
		if len(valueGetters) == 0 {
			if len(cfg.defaultGetOption) > 0 {
				valueGetters = cfg.defaultGetOption
			} else {
				valueGetters = []ValueRetriever[*simpleConfig]{OptionSimpleConfigGetValue(First)}
			}
		}
		var generatedKeys []Key
		// current key generation trys all generators and uses the first key generated
		for _, keyGenerator := range keyGenerators {
			newKey, err := keyGenerator(ctx, keyGenerators, key, generatedKeys)
			if err != nil {
				return nil, err
			} else if len(newKey) > 0 {
				generatedKeys = newKey
				break
			}
		}
		if len(generatedKeys) == 0 {
			return nil, ErrWithParameters(ctx, ErrConfigKeyParsingFailed, cfg, key, nil, nil)
		}
		var extractedValues []any
		// Current implementation returns the first value returned by value getters.
		for _, valueGetter := range valueGetters {
			newValue, err := valueGetter(ctx, cfg, key, generatedKeys, extractedValues)
			if err != nil {
				if !errors.Is(err, ErrConfigMissingValue) {
					return nil, ErrWithParameters(ctx, ErrConfigInvalidValue, cfg, key, newValue, err)
				}
			} else {
				extractedValues = append(extractedValues, newValue)
				break
			}
		}
		numberOfExtractedValues := len(extractedValues)
		switch {
		// use the first value if multiple values returned
		case numberOfExtractedValues >= 1:
			applicableValue = extractedValues[0]
		case numberOfExtractedValues == 0:
			fallthrough
		default:
			return nil, ErrWithParameters(ctx, ErrConfigMissingValue, cfg, key, applicableValue, nil)
		}
	} else {
		return nil, ErrWithParameters(ctx, ErrConfigNilConfig, cfg, key, applicableValue, fmt.Errorf("either config %v or associated map is nil", cfg))
	}
	return
}

func (cfg *simpleConfig) GetConfig(ctx context.Context, key string, options ...GetOption) (Config, error) {
	getValue, getErr := cfg.GetValue(ctx, key, options...)
	if getValue == nil {
		return nil, ErrWithParameters(ctx, ErrConfigMissingValue, cfg, key, getValue, getErr)
	}
	if getErr != nil {
		return nil, ErrWithParameters(ctx, ErrConfigInvalidValue, cfg, key, getValue, getErr)
	}
	var applicableConfigMap, mapGenErr = GetAsMap(ctx, getValue, ExtensibleErr(ctx, cfg, key))
	if mapGenErr != nil {
		return nil, mapGenErr
	}
	applicableConfig, cfgErr := NewConfigFromDriver[SimpleConfigDriver, SimpleConfig](ctx, defaultConfigDriver, OptionSimpleConfigWithMap(applicableConfigMap))
	if cfgErr != nil {
		return nil, ErrWithParameters(ctx, ErrConfigInvalidValue, cfg, key, getValue, fmt.Errorf("creation of configuration from map %v failed with error %#w", applicableConfigMap, cfgErr))
	}
	return applicableConfig, nil
}

func OptionSimpleConfigWithMap(cfg map[string]any) InitOption[*simpleConfig] {
	return func(ctx context.Context, mgr *simpleConfig) error {
		mgr.configurationMap = cfg
		mgr.defaultKeyParsers = []KeyParser{DefaultKeyParser(ctx)}
		mgr.defaultGetOption = []ValueRetriever[*simpleConfig]{OptionSimpleConfigGetValue(First)}
		return nil
	}
}
