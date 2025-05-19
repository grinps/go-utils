package config

import (
	"context"
	"testing"
)

var ctx = context.Background()
var configData = map[string]any{"K1": "V1", "K2": map[string]any{"K21": "V21"}}

func TestSimpleConfigDriver_NewConfigMgr(t *testing.T) {
	t.Run("NilDriverNoOption", func(t *testing.T) {
		var nilDriver *simpleConfigDriver
		config, err := nilDriver.NewConfigMgr(ctx)
		if config == nil {
			t.Errorf("Expected not nil configuration, actual nil")
		}
		if err != nil {
			t.Errorf("Expected no error actual %s", err.Error())
		}
	})
	t.Run("NilDriverDefaultOption", func(t *testing.T) {
		var nilDriver *simpleConfigDriver
		config, err := nilDriver.NewConfigMgr(ctx, OptionSimpleConfigWithMap(configData))
		if config == nil {
			t.Errorf("Expected not nil configuration, actual nil")
		}
		if err != nil {
			t.Errorf("Expected no error actual %s", err.Error())
		}
	})
	t.Run("NilDriverNilOption", func(t *testing.T) {
		var nilDriver *simpleConfigDriver
		config, err := nilDriver.NewConfigMgr(ctx, nil)
		if config == nil {
			t.Errorf("Expected not nil configuration, actual nil")
		}
		if err != nil {
			t.Errorf("Expected no error actual %s", err.Error())
		}
	})
	t.Run("defaultDriverNilOption", func(t *testing.T) {
		var defaultDriver *simpleConfigDriver = &simpleConfigDriver{}
		config, err := defaultDriver.NewConfigMgr(ctx, nil)
		if config == nil {
			t.Errorf("Expected not nil configuration, actual nil")
		}
		if err != nil {
			t.Errorf("Expected no error actual %s", err.Error())
		}
	})
	//t.Run("")
}

func KeyParserSameValue(keyValue string) KeyParser {
	return func(ctx context.Context, keyParsers []KeyParser, key string, previousParsedKeys []Key) (parsedKey []Key, parsingErr error) {
		var aKey = simpleStringKey(keyValue)
		return []Key{&aKey}, nil
	}
}

func KeyParserWithErr(errCode ErrCode, key string) KeyParser {
	return func(ctx context.Context, keyParsers []KeyParser, key string, previousParsedKeys []Key) (parsedKey []Key, parsingErr error) {
		return nil, ErrWithParameters(context.Background(), errCode, nil, key, nil, nil)
	}
}
