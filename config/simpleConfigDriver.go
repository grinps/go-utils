package config

import (
	"context"
	"errors"
	"fmt"
)

const SimpleConfigDriverName = "SimpleConfigDriverName"

type SimpleConfigDriver = *simpleConfigDriver

var defaultConfigDriver SimpleConfigDriver = &simpleConfigDriver{}

type simpleConfigDriver struct {
}

func (driver *simpleConfigDriver) Name() string {
	return SimpleConfigDriverName
}

func (driver *simpleConfigDriver) NewConfigMgr(ctx context.Context, options ...DriverOption) (*simpleConfig, error) {
	var config = &simpleConfig{
		configurationMap:  map[string]any{},
		defaultGetOption:  []ValueRetriever[*simpleConfig]{},
		defaultKeyParsers: []KeyParser{},
	}
	var optionErrors []error
	initOptions, keyParsers, getValueOptions := ClassifyOptions[*simpleConfig](options...)
	for _, source := range initOptions {
		optionErr := source(ctx, config)
		if optionErr != nil {
			optionErrors = append(optionErrors, optionErr)
		}
	}
	if len(keyParsers) > 0 {
		config.defaultKeyParsers = keyParsers
	}
	if len(getValueOptions) > 0 {
		config.defaultGetOption = getValueOptions
	}
	if optionErrors != nil {
		return nil, fmt.Errorf("failed to create simple config due to error %w", errors.Join(optionErrors...))
	} else {
		return config, nil
	}
}
