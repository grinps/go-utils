package system

import (
	"context"
	"github.com/grinps/go-utils/errext"
	"testing"
)

var counterTrackers = map[string]int{}

func WithServiceRetrievalTracker(counterName string) GetOption {
	counterTrackers[counterName] = 0
	return func(context context.Context, system System, serviceIdentifier string, serviceTypeIdentifier string,
		options []GetOption, retrievedService any, previousApplicableValue any, err *error) (applicableValue any) {
		counterTrackers[counterName] = counterTrackers[counterName] + 1
		return retrievedService
	}
}

type notASystemImpl struct{}

func (system *notASystemImpl) System() System {
	return system
}

func TestChangeInitializationGetOptions(t *testing.T) {
	defaultSystem := NewSystem()
	t.Run("SingleImpl", func(t *testing.T) {
		ChangeInitializationGetOptions(WithServiceRetrievalTracker("SingleImpl"))
		systemWith1Tracker := NewSystem()
		if asSystemImpl, isSystemImpl := systemWith1Tracker.(*systemImpl); isSystemImpl {
			if numberOfGetOptions := len(asSystemImpl.defaultGetOptions); numberOfGetOptions != 1 {
				t.Errorf("Expected 1 Get options, actual %d", numberOfGetOptions)
			}
		} else {
			t.Errorf("Expected implementation of systemImpl, actual %#v", defaultSystem)
		}
		_, err := GetServiceFromSystem(context.TODO(), systemWith1Tracker, "SomeService", TestServiceType)
		if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
		if counterTrackers["SingleImpl"] != 1 {
			t.Errorf("Expected value of counter 1, actual %d", counterTrackers["SingleImpl"])
		}
	})
	t.Run("EmptyValue", func(t *testing.T) {
		ChangeInitializationGetOptions()
		systemWithEmptyValue := NewSystem()
		if asSystemImpl, isSystemImpl := systemWithEmptyValue.(*systemImpl); isSystemImpl {
			if len(asSystemImpl.defaultGetOptions) != 0 {
				t.Errorf("Expected no Get options since changed option to be empty.")
			}
		} else {
			t.Errorf("Expected implementation of systemImpl, actual %#v", defaultSystem)
		}
	})
}

func TestChangeSystemGetOptions(t *testing.T) {
	defaultSystem := NewSystem()
	if asSystemImpl, isSystemImpl := defaultSystem.(*systemImpl); isSystemImpl {
		if len(asSystemImpl.defaultGetOptions) > 0 {
			t.Errorf("Expected no Get options since default is empty.")
		}
	} else {
		t.Errorf("Expected implementation of systemImpl, actual %#v", defaultSystem)
	}
	t.Run("SingleImpl", func(t *testing.T) {
		changeErr := ChangeSystemGetOptions(defaultSystem, WithServiceRetrievalTracker("SingleImpl"))
		if changeErr != nil {
			t.Errorf("Expected no error, actual %#v", changeErr)
		}
		if asSystemImpl, isSystemImpl := defaultSystem.(*systemImpl); isSystemImpl {
			if numberOfGetOptions := len(asSystemImpl.defaultGetOptions); numberOfGetOptions != 1 {
				t.Errorf("Expected 1 Get options, actual %d", numberOfGetOptions)
			}
		} else {
			t.Errorf("Expected implementation of systemImpl, actual %#v", defaultSystem)
		}
		_, err := GetServiceFromSystem(context.TODO(), defaultSystem, "SomeService", TestServiceType)
		if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
		if counterTrackers["SingleImpl"] != 1 {
			t.Errorf("Expected value of counter 1, actual %d", counterTrackers["SingleImpl"])
		}
	})
	t.Run("EmptyValue", func(t *testing.T) {
		anotherSystem := NewSystem()
		changeErr := ChangeSystemGetOptions(defaultSystem)
		if changeErr != nil {
			t.Errorf("Expected no error, actual %#v", changeErr)
		}
		if asSystemImpl, isSystemImpl := anotherSystem.(*systemImpl); isSystemImpl {
			if len(asSystemImpl.defaultGetOptions) != 0 {
				t.Errorf("Expected no Get options since changed option to be empty.")
			}
		} else {
			t.Errorf("Expected implementation of systemImpl, actual %#v", defaultSystem)
		}
	})
	t.Run("NotAsystemImpl", func(t *testing.T) {
		var aRandomSystem = &notASystemImpl{}
		changeErr := ChangeSystemGetOptions(aRandomSystem, WithServiceRetrievalTracker("NotAsystemImpl"))
		if changeErr == nil {
			t.Errorf("Expected error, actual no err")
		} else if !errext.Is(changeErr, ErrChangeSystemGetOptions) {
			t.Errorf("Expected ErrChangeSystemGetOptions, actual %#v", changeErr)
		}
	})
}

func TestGetSystemGetOptions(t *testing.T) {
	t.Run("NotAsystemImpl", func(t *testing.T) {
		var aRandomSystem = &notASystemImpl{}
		options := GetSystemGetOptions(aRandomSystem)
		if options != nil {
			t.Errorf("Expected no error, actual %#v", options)
		}
	})
}

func TestSystemImpl_GetService(t *testing.T) {
	aValidSystem := NewSystem()
	notInitailizedSystem := &systemImpl{}
	notASystem := &notASystemImpl{}
	t.Run("LookupNotRegisteredNoOptions", func(t *testing.T) {
		aService, getErr := GetServiceFromSystem(context.TODO(), aValidSystem, "someService", TestServiceType)
		if getErr != nil {
			t.Errorf("Expected no error actual error %#v", getErr)
		}
		if aService != nil {
			t.Errorf("Expected nil, actual %#v", aService)
		}
	})
	t.Run("LookupNotASystemImplNotRegisteredNoOptions", func(t *testing.T) {
		aService, getErr := GetServiceFromSystem(context.TODO(), notASystem, "someService", TestServiceType)
		if getErr == nil {
			t.Errorf("Expected error actual no error")
		} else if !errext.Is(getErr, ErrSystemGetServiceFailed) {
			t.Errorf("Expected ErrSystemGetServiceFailed, actual %#v", getErr)
		}
		if aService != nil {
			t.Errorf("Expected nil, actual %#v", aService)
		}
	})
	t.Run("LookupNotInitializedSystemNotRegisteredNoOptions", func(t *testing.T) {
		aService, getErr := GetServiceFromSystem(context.TODO(), notInitailizedSystem, "someService", TestServiceType)
		if getErr == nil {
			t.Errorf("Expected error actual no error")
		} else if !errext.Is(getErr, ErrSystemGetServiceFailed) {
			t.Errorf("Expected ErrSystemGetServiceFailed, actual %#v", getErr)
		}
		if aService != nil {
			t.Errorf("Expected nil, actual %#v", aService)
		}
	})
	t.Run("LookupNilSystemNotRegisteredNoOptions", func(t *testing.T) {
		aService, getErr := GetServiceFromSystem(context.TODO(), nil, "someService", TestServiceType)
		if getErr == nil {
			t.Errorf("Expected error actual no error")
		} else if !errext.Is(getErr, ErrSystemGetServiceFailed) {
			t.Errorf("Expected ErrSystemGetServiceFailed, actual %#v", getErr)
		}
		if aService != nil {
			t.Errorf("Expected nil, actual %#v", aService)
		}
	})
	t.Run("LookupNilSystemNilServicetypeNotRegisteredNoOptions", func(t *testing.T) {
		aService, getErr := GetServiceFromSystem(context.TODO(), nil, "someService", ServiceType[TestInterface](nil))
		if getErr == nil {
			t.Errorf("Expected error actual no error")
		} else if !errext.Is(getErr, ErrSystemGetServiceFailed) {
			t.Errorf("Expected ErrSystemGetServiceFailed, actual %#v", getErr)
		}
		if aService != nil {
			t.Errorf("Expected nil, actual %#v", aService)
		}
	})
}
