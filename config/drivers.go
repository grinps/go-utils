package config

import (
	"context"
	"fmt"
	"reflect"
)

func NewConfig[D Driver[C], C Config](ctx context.Context, driverOptions ...DriverOption) C {
	var driverObj D
	driver, driverErr := NewConfigFromDriver[D, C](ctx, driverObj, driverOptions...)
	if driverErr != nil {
		panic(fmt.Errorf("failed to create config from driver %s due to error %#w", reflect.TypeOf(driverObj), driverErr))
	}
	return driver
}

func NewConfigFromDriver[D Driver[C], C Config](ctx context.Context, driver D, driverOptions ...DriverOption) (C, error) {
	return driver.NewConfigMgr(ctx, driverOptions...)
}

func NewConfigFromDriverName[D Driver[C], C Config](ctx context.Context, driverName string, driverOptions ...DriverOption) (C, error) {
	var nilC C
	if driver := commonRegistries.DriverRegistry.Get(driverName); driver != nil {
		if asGivenDriverType, isGivenDriverType := driver.(D); isGivenDriverType {
			return NewConfigFromDriver[D, C](ctx, asGivenDriverType, driverOptions...)
		} else {
			var zero [0]D
			var driverType = reflect.TypeOf(zero).Elem()
			return nilC, fmt.Errorf("can not create new config since registered driver %#v for name %s is not of expected type %s", driver, driverName, driverType)
		}
	} else {
		return nilC, fmt.Errorf("can not create new config since no driver corresponding to given name %s has been registered", driverName)
	}
}

func WithDriver[D Driver[C], C Config]() RegistrationOption {
	var driverObj D
	var driverName string
	if driverName = driverObj.Name(); driverName == "" {
		panic(fmt.Errorf("expected configuration driver %#v to return non-empty string as name of driver", driverObj))
	}
	return RegistrationOptionF(func(ctx context.Context, r *registries) {
		driverN := driverName
		r.DriverRegistry.Register(driverN, driverObj)
	})
}
