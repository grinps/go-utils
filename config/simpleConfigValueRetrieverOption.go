package config

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

type HitPolicy int

const (
	Default HitPolicy = iota
	First
)

func OptionSimpleConfigGetValue(hitPolicy HitPolicy) ValueRetriever[*simpleConfig] {
	if hitPolicy == Default {
		hitPolicy = First
	}
	return func(ctx context.Context, config *simpleConfig, origKey string, keys []Key, previousValue any) (value any, getErr error) {
		applicableHitPolicy := hitPolicy
		extErr := ExtensibleErr(ctx, config, origKey)
		if config == nil {
			return nil, extErr(WithErrorCode(ErrConfigNilConfig), WithCause(fmt.Errorf("nil config")))
		}
		if config.configurationMap == nil {
			return nil, extErr(WithErrorCode(ErrConfigNilConfig), WithCause(fmt.Errorf("nil configuration map associated with simple config")))
		}
		if len(config.configurationMap) == 0 {
			return nil, extErr(WithErrorCode(ErrConfigNilConfig), WithCause(fmt.Errorf("empty configuration map associated with simple config")))

		}
		if applicableHitPolicy != First {
			return nil, extErr(WithErrorCode(ErrConfigInvalidValue), WithCause(fmt.Errorf("unsupported hit policy %d", applicableHitPolicy)))
		}
		if len(keys) > 0 {
			applicableValue := value
			for _, key := range keys {
				if key == nil {
					continue
				}
				getValue, getErr := getDataUsingKey(ctx, config.configurationMap, key, ExtensibleErr(ctx, config, origKey))
				if errors.Is(getErr, ErrConfigMissingValue) {
					continue
				} else if getErr != nil {
					return nil, getErr
				}
				applicableValue = getValue
				break
			}
			value = applicableValue
			return value, nil
		} else {
			return nil, extErr(WithErrorCode(ErrConfigInvalidKey), WithCause(fmt.Errorf("given key is of size 0")))
		}
	}
}

func getDataUsingKey(ctx context.Context, applicableMap map[string]any, key Key, extErr ExtensibleErrFunc) (any, error) {
	keyType := reflect.TypeOf(key)
	switch {
	case keyType.Implements(hierarchyKeyType):
		if asHierarchyKey, isHierarchyKey := key.(HierarchyKey); isHierarchyKey {
			var returnValue any = applicableMap
			if keyType.Implements(simpleKeyType) {
				getValue, getErr := getDataUsingSimpleKey(ctx, applicableMap, key, extErr)
				if getErr != nil {
					return nil, getErr
				}
				returnValue = getValue
			}
			applicableHierarchyKey := asHierarchyKey
			if applicableHierarchyKey.HasChild(ctx) {
				transformedMap, transformErr := GetAsMap(ctx, returnValue, extErr)
				if transformErr != nil {
					return nil, transformErr
				}
				childKey := applicableHierarchyKey.Child(ctx)
				//recursion - should we change to for loop
				getValue, getErr := getDataUsingKey(ctx, transformedMap, childKey, extErr)
				if getErr != nil {
					return nil, getErr
				}
				returnValue = getValue
			}
			return returnValue, nil
		} else {
			return nil, extErr(WithErrorCode(ErrConfigInvalidKey), WithCause(fmt.Errorf("failed to cast expected hierarchy key %#v of type %T", key, key)))
		}
	case keyType.Implements(simpleKeyType):
		return getDataUsingSimpleKey(ctx, applicableMap, key, extErr)
	default:
		return nil, extErr(WithErrorCode(ErrConfigInvalidKey), WithCause(fmt.Errorf("key %#v type %T is not supported (only SimpleKey & HierarchyKey are currently supported)", key, key)))
	}
}

func getDataUsingSimpleKey(ctx context.Context, applicableMap map[string]any, key Key, extErr ExtensibleErrFunc) (any, error) {
	if asSimpleKey, isSimpleKey := key.(SimpleKey); isSimpleKey {
		if simpleKeyName := asSimpleKey.Name(ctx); simpleKeyName != "" && simpleKeyName != InvalidKeyName {
			if mappedValue, hasValue := applicableMap[simpleKeyName]; hasValue {
				return mappedValue, nil
			} else {
				return nil, extErr(WithErrorCode(ErrConfigMissingValue))
			}
		} else {
			return nil, extErr(WithErrorCode(ErrConfigInvalidKey), WithCause(fmt.Errorf("failed to cast expected simple key %#v that is part of hierarchy key of type %T", key, key)))
		}
	} else {
		return nil, extErr(WithErrorCode(ErrConfigInvalidKey), WithCause(fmt.Errorf("failed to cast expected simple key %#v of type %T", key, key)))
	}
}
