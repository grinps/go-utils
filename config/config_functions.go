package config

import (
	"context"
	logger "github.com/grinps/go-utils/base-utils/logs"
)

var defaultConfig Config = NewConfig[SimpleConfigDriver, SimpleConfig](context.Background(), OptionSimpleConfigWithMap(map[string]any{}))

func init() {
	logger.Log("Initializing configuration module.")
}

func GetValue[T any](ctx context.Context, key string, options ...GetOption) T {
	retValue, _ := GetValueE[T](ctx, key, options...)
	return retValue
}

func GetValueP[T any](ctx context.Context, key string, options ...GetOption) T {
	retValue, err := GetValueE[T](ctx, key, options...)
	if err != nil {
		panic(err)
	}
	return retValue
}

func GetValueE[T any](ctx context.Context, key string, options ...GetOption) (applicableValue T, getValueErr error) {
	if key != "" {
		value, valueGetErr := defaultConfig.GetValue(ctx, key, options...)
		if valueGetErr == nil {
			var isT bool
			applicableValue, isT = value.(T)
			if !isT {
				getValueErr = ErrWithParameters(ctx, ErrConfigInvalidValueType, defaultConfig, key, value, nil)
			}
		} else {
			getValueErr = valueGetErr
		}
	} else {
		getValueErr = ErrWithParameters(ctx, ErrConfigEmptyKey, defaultConfig, key, applicableValue, nil)
	}
	return applicableValue, getValueErr
}

func ClassifyOptions[C Config](options ...DriverOption) (sources []InitOption[C], keyGens []KeyParser, getValOpts []ValueRetriever[C]) {
	for _, option := range options {
		if option == nil {
			continue
		}
		if asInitOption, isInitOption := option.(InitOption[C]); isInitOption {
			sources = append(sources, asInitOption)
		} else if asKeyParser, isKeyParser := option.(KeyParser); isKeyParser {
			keyGens = append(keyGens, asKeyParser)
		} else if asValueRetriever, isValueRetriever := option.(ValueRetriever[C]); isValueRetriever {
			getValOpts = append(getValOpts, asValueRetriever)
		}
	}
	return
}

func ClassifyGetOptions[C Config](options ...GetOption) (keyGens []KeyParser, getValOpts []ValueRetriever[C]) {
	for _, option := range options {
		if option == nil {
			continue
		}
		if asKeyParser, isKeyParser := option.(KeyParser); isKeyParser {
			keyGens = append(keyGens, asKeyParser)
		} else if asValueRetriever, isValueRetriever := option.(ValueRetriever[C]); isValueRetriever {
			getValOpts = append(getValOpts, asValueRetriever)
		}
	}
	return
}

type contextConfigType int

const ContextConfigReference contextConfigType = 1

func ContextWithConfig(ctx context.Context, config Config) context.Context {
	if ctx != nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, ContextConfigReference, config)
}

func ContextConfig(ctx context.Context, defaultIfNotAvailable bool) Config {
	if ctx != nil {
		config := ctx.Value(ContextConfigReference)
		if config != nil {
			if asConfig, isConfig := config.(Config); isConfig {
				return asConfig
			}
		}
	}
	if defaultIfNotAvailable {
		return defaultConfig
	} else {
		return nil
	}
}
