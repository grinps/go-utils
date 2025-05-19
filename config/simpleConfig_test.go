package config_test

import "testing"
import "github.com/grinps/go-utils/config"

func TestSimpleConfig_GetConfig(t *testing.T) {
	configObject := config.NewConfig[config.SimpleConfigDriver, config.SimpleConfig](ctx, config.OptionSimpleConfigWithMap(configData))
	if configObject == nil {
		t.Errorf("Expected not nil configuration, actual nil")
	}
	t.Run("ValidConfigValidKey", func(t *testing.T) {
		cfgVal, cfgValErr := configObject.GetConfig(ctx, "K2", config.OptionSimpleConfigGetValue(config.Default))
		if cfgValErr != nil {
			t.Errorf("Expected no error actual %s", cfgValErr)
		}
		if cfgVal == nil {
			t.Errorf("Expected valid value actual nil")
		}
	})
}

func TestSimpleConfig_GetValue(t *testing.T) {
	configObject := config.NewConfig[config.SimpleConfigDriver, config.SimpleConfig](ctx, config.OptionSimpleConfigWithMap(configData))
	if configObject == nil {
		t.Errorf("Expected not nil configuration, actual nil")
	}
	interVal, interValErr := configObject.GetValue(ctx, "K2-K21",
		config.DefaultKeyParser(ctx, config.OptionKeyParserHierarchyKeyDelimiter("-")))
	if interValErr != nil {
		t.Errorf("Expected no error actual %v", interValErr)
	}
	if interVal != "V21" {
		t.Errorf("Expected V21 actual %#v", interVal)
	}

}
