package system

import (
	"github.com/grinps/go-utils/errext"
	"testing"
)

type simpleService struct {
}
type SimpleService *simpleService

var aSimpleService *simpleService = &simpleService{}

const TestServiceTypeName = "TestServiceType"

var TestServiceType = NewServiceType[TestInterface](TestServiceTypeName)

func TestNewNamedServiceE(t *testing.T) {
	t.Run("EmptyServiceName", func(t *testing.T) {
		sName, err := NewNamedServiceE("", aSimpleService)
		if sName != nil {
			t.Errorf("Expected namedService to nil actual %#v", sName)
		}
		if err == nil {
			t.Errorf("Expected error actual nil error")
		} else if _, isValidErr := ErrServiceCreationFailed.AsError(err); !isValidErr {
			t.Errorf("Expected error as ErrServiceCreationFailed, Actual %#v", err)
		}
	})
	t.Run("NilService", func(t *testing.T) {
		sName, err := NewNamedServiceE[SimpleService]("ValidService", nil)
		if sName != nil {
			t.Errorf("Expected namedService name to be nil actual %#v", sName)
		}
		if err == nil {
			t.Errorf("Expected error actual nil error")
		} else if _, isValidErr := ErrServiceCreationFailed.AsError(err); !isValidErr {
			t.Errorf("Expected error as ErrServiceCreationFailed, Actual %#v", err)
		}
	})
	t.Run("Valid", func(t *testing.T) {
		servName := "ValidService"
		sName, err := NewNamedServiceE(servName, aSimpleService)
		if sName == nil {
			t.Errorf("Expected namedService name to not be nil actual nil")
		}
		if err != nil {
			t.Errorf("Expected no error actual %#v", err)
		} else {
			if sName.String() != servName {
				t.Errorf("Expected namedService name %#v actual %#v", servName, sName.String())
			}
		}
	})
	t.Run("Duplicate", func(t *testing.T) {
		t.Skip("Skipping since duplicate services are supported.")
		servName := "DuplicateService"
		sName, err := NewNamedServiceE(servName, aSimpleService)
		if sName == nil {
			t.Errorf("Expected service type to not be nil actual nil")
		}
		if err != nil {
			t.Errorf("Expected no error actual %#v", err)
		}
		_, err2 := NewNamedServiceE(servName, aSimpleService)
		if err2 == nil {
			t.Errorf("Expected error actual no error")
		} else if !errext.Is(err2, ErrServiceCreationFailed) {
			t.Errorf("Expected error of type ErrServiceCreationFailed actual %#v", err2)
		}
	})
}

func TestNewNamedServiceP(t *testing.T) {
	t.Run("InvalidName", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Errorf("Expected a panic, actual no panic")
			}
		}()
		var _ = NewNamedServiceP("", aSimpleService)
	})
	t.Run("ValidName", func(t *testing.T) {
		defer func() {
			if recover() != nil {
				t.Errorf("Expected no panic, actual panic %#v", recover())
			}
		}()
		var _ = NewNamedServiceP("AValidService", aSimpleService)
	})
}

func TestNewNamedService(t *testing.T) {
	t.Run("InvalidName", func(t *testing.T) {
		_ = NewNamedService("", aSimpleService)
	})
	t.Run("ValidName", func(t *testing.T) {
		defer func() {
			recovered := recover()
			if recovered != nil {
				t.Errorf("Expected no panic while creating valid named service, actual error %#v", recovered)
			}
		}()
		_ = NewNamedService("SomeRandomServiceName", aSimpleService)
	})
}

func TestNewServiceTypeE(t *testing.T) {
	t.Run("EmptyServiceTypeName", func(t *testing.T) {
		sName, err := NewServiceTypeE[*simpleService]("")
		if sName != nil {
			t.Errorf("Expected serviceType as nil actual %#v", sName)
		}
		if err == nil {
			t.Errorf("Expected error actual nil error")
		} else if !errext.Is(err, ErrServiceTypeCreationFailed) {
			t.Errorf("Expected error as ErrServiceCreationFailed, Actual %#v", err)
		}
	})
	t.Run("Valid", func(t *testing.T) {
		servName := "ValidServiceType"
		sName, err := NewServiceTypeE[*simpleService](servName)
		if sName == nil {
			t.Errorf("Expected service type to not be nil actual nil")
		}
		if err != nil {
			t.Errorf("Expected no error actual %#v", err)
		} else {
			if sName.String() != servName {
				t.Errorf("Expected serviceType name %#v actual %#v", servName, sName.String())
			}
		}
	})
	t.Run("Duplicate", func(t *testing.T) {
		servName := "DuplicateServiceType"
		sName, err := NewServiceTypeE[*simpleService](servName)
		if sName == nil {
			t.Errorf("Expected service type to not be nil actual nil")
		}
		if err != nil {
			t.Errorf("Expected no error actual %#v", err)
		}
		_, err2 := NewServiceTypeE[*simpleService](servName)
		if err2 == nil {
			t.Errorf("Expected error actual no error")
		} else if !errext.Is(err2, ErrServiceTypeCreationFailed) {
			t.Errorf("Expected error of type ErrServiceTypeCreationFailed actual %#v", err2)
		}
	})
}

func TestNewServiceTypeP(t *testing.T) {
	t.Run("InvalidName", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Errorf("Expected a panic, actual no panic")
			}
		}()
		var _ = NewServiceTypeP[*simpleService]("")
	})
	t.Run("ValidName", func(t *testing.T) {
		defer func() {
			if recover() != nil {
				t.Errorf("Expected no panic, actual panic %#v", recover())
			}
		}()
		var _ = NewServiceTypeP[*simpleService]("AValidService")
	})
}

func TestNewServiceType(t *testing.T) {
	t.Run("InvalidName", func(t *testing.T) {
		serviceType := NewServiceType[*simpleService]("")
		if serviceType != nil {
			t.Errorf("Expected nil service type, actual %#v", serviceType)
		}
	})
	t.Run("ValidName", func(t *testing.T) {
		serviceType := NewServiceType[*simpleService]("SomeRandomServiceName")
		if serviceType == nil {
			t.Errorf("Expected not nil service type, actual nil")
		}
	})
}

func TestNamedService_String(t *testing.T) {
	t.Run("DefaultObject", func(t *testing.T) {
		namedServiceVal := namedService[TestInterface]{}
		if namedServiceVal.String() != InvalidServiceName {
			t.Errorf("Expected %s, actual %s", InvalidServiceName, namedServiceVal.String())
		}
	})
}

func TestNamedService_As(t *testing.T) {
	t.Run("ValidServiceType", func(t *testing.T) {
		servType := NewServiceType[TestInterface]("TestInterface")
		aService := TestStruct{input: 42}
		outServ, err := servType.As(aService)
		if err != nil {
			t.Errorf("Expected success, actual %#v", err)
		}
		if outServ == nil {
			t.Errorf("Expected not nil output, actual nil")
		} else if outVal := outServ.aFunc(39); outVal != "42:39" {
			t.Errorf("Expected 42:39, actual %s", outVal)
		}
	})
}

func TestNamedService_AsService(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		var nilNamedService *namedService[TestInterface] = nil
		var nilService Service[TestInterface] = nilNamedService
		if getService := nilNamedService.AsService(); getService != TestInterface(nil) {
			t.Errorf("nilNamedService: Expected TestInterface(nil), Actual %#v", getService)
		}
		if anotherService := nilService.AsService(); anotherService != TestInterface(nil) {
			t.Errorf("nilService: Expected TestInterface(nil), Actual %#v", anotherService)
		}
	})
	t.Run("ValidService", func(t *testing.T) {
		aService := &TestStruct{input: 42}
		service := NewNamedService[TestInterface]("TestStruct", aService)
		outServ := service.AsService()
		if outServ == nil {
			t.Errorf("Expected not nil output, actual nil")
		} else if outVal := outServ.aFunc(39); outVal != "42:39" {
			t.Errorf("Expected 42:39, actual %s", outVal)
		}
	})
	t.Run("InvalidServiceWithDefaultService", func(t *testing.T) {
		service := namedService[TestInterface]{name: "JunkService"}
		outServ := service.AsService()
		if outServ != TestInterface(nil) {
			t.Errorf("Expected nil output, actual %#v", outServ)
		}
	})

	t.Run("InvalidServiceWithNilService", func(t *testing.T) {
		service := namedService[TestInterface]{name: "JunkService", service: nil}
		outServ := service.AsService()
		if outServ != TestInterface(nil) {
			t.Errorf("Expected nil output, actual %#v", outServ)
		}
	})
}

func TestNamedService_Key(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		var nilNamedService *namedService[TestInterface] = nil
		var nilService Service[TestInterface] = nilNamedService
		if serviceKey := nilNamedService.Key(); serviceKey != InvalidServiceName {
			t.Errorf("nilNamedService: Expected %s, Actual %#v", InvalidServiceName, serviceKey)
		}
		if anotherServiceKey := nilService.(Comparable[string]).Key(); anotherServiceKey != InvalidServiceName {
			t.Errorf("nilService: Expected %s, Actual %#v", InvalidServiceName, anotherServiceKey)
		}
	})
	t.Run("ValidService", func(t *testing.T) {
		aService := &TestStruct{input: 42}
		aNamedService, namedServiceErr := NewNamedServiceE[TestInterface]("TestStruct1", aService)
		if namedServiceErr != nil {
			t.Errorf("Expected no error, actual %#v", namedServiceErr)
		} else {
			if service, ok := aNamedService.(Comparable[string]); ok {
				keyValue := service.Key()
				if keyValue != "TestStruct1" {
					t.Errorf("Expected TestStruct1, actual %s", keyValue)
				}
			} else {
				t.Errorf("Expected named service to have a string key, actual %#v", aNamedService)
			}
		}
	})
	t.Run("InvalidServiceWithDefaultService", func(t *testing.T) {
		service := namedService[TestInterface]{name: "JunkService"}
		keyValue := service.Key()
		if keyValue != "JunkService" {
			t.Errorf("Expected JunkService, actual %s", keyValue)
		}
	})
	t.Run("DefaultNamedService", func(t *testing.T) {
		service := namedService[TestInterface]{}
		keyValue := service.Key()
		if keyValue != InvalidServiceName {
			t.Errorf("Expected %s, actual %s", InvalidServiceName, keyValue)
		}
	})
}
