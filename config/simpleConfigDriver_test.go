package config_test

import (
	"context"
	"github.com/grinps/go-utils/config"
	"testing"
)

var ctx = context.Background()
var configData = map[string]any{"K1": "V1", "K2": map[string]any{"K21": "V21"}}

func TestSimpleConfigDriver_NewConfigMgr2(t *testing.T) {
	t.Run("ValidConfigDefaultGet", func(t *testing.T) {
		cfgObject := config.NewConfig[config.SimpleConfigDriver,
			config.SimpleConfig](ctx, config.OptionSimpleConfigWithMap(configData))
		if cfgObject == nil {
			t.Errorf("Expected not nil configuration, actual nil")
		}
		val, valErr := cfgObject.GetValue(ctx, "K1")
		if valErr != nil {
			t.Errorf("Expected no error actual %s", valErr)
		}
		if val != "V1" {
			t.Errorf("Expected V1 actual %#v", val)
		}
	})
}
