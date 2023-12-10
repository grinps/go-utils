package system

import (
	"github.com/grinps/go-utils/errext"
	"testing"
)

type testService struct {
}
type TestService *testService

var testServiceInstace = &testService{}

func TestSystemImpl_System(t *testing.T) {
	t.Run("Nil", func(t *testing.T) {
		var nilSystemImpl *systemImpl
		if sysReturned := nilSystemImpl.System(); sysReturned != (*systemImpl)(nil) {
			t.Errorf("Expected nil value, actual %#v", sysReturned)
		} else if isInit, initErr := nilSystemImpl.isInitialized(); isInit {
			t.Errorf("Expected not initialized, actual initialized")
		} else if !errext.Is(initErr, ErrSystemNotInitialized) {
			t.Errorf("Expected ErrSystemNotInitialized, actual %#v", initErr)
		}
	})
	t.Run("default", func(t *testing.T) {
		var defaultSysImpl *systemImpl = &systemImpl{}
		if sysReturned := defaultSysImpl.System(); sysReturned == nil {
			t.Errorf("Expected not nil value, actual nil")
		} else if isInit, initErr := defaultSysImpl.isInitialized(); isInit {
			t.Errorf("Expected not initialized, actual initialized")
		} else if !errext.Is(initErr, ErrSystemNotInitialized) {
			t.Errorf("Expected ErrSystemNotInitialized, actual %#v", initErr)
		}
	})
	t.Run("Valid", func(t *testing.T) {
		var validSystem System = NewSystem()
		if sysReturned := validSystem.System(); sysReturned == nil {
			t.Errorf("Expected not nil value, actual nil")
		} else if asSysImpl, isSysImpl := validSystem.(*systemImpl); !isSysImpl {
			t.Errorf("Expected implementation of *systemImpl, actual %#v", validSystem)
		} else if asSysImpl == nil {
			t.Errorf("Expected not nil implementation of *systemImpl, actual nil")
		} else if isInit, initErr := asSysImpl.isInitialized(); !isInit {
			t.Errorf("Expected initialized object actual not initialized. initErr %#v", initErr)
		}
	})
}

func TestNewSystemWithOptions(t *testing.T) {
	t.Run("No Options", func(t *testing.T) {
		sysNoOptions := NewSystemWithOptions()
		if sysReturned := sysNoOptions.System(); sysReturned == nil {
			t.Errorf("Expected not nil value, actual nil")
		} else if asSysImpl, isSysImpl := sysNoOptions.(*systemImpl); !isSysImpl {
			t.Errorf("Expected implementation of *systemImpl, actual %#v", asSysImpl)
		} else if asSysImpl == nil {
			t.Errorf("Expected not nil implementation of *systemImpl, actual nil")
		} else {
			if isInit, initErr := asSysImpl.isInitialized(); !isInit {
				t.Errorf("Expected initialized object actual not initialized. initErr %#v", initErr)
				if len(asSysImpl.defaultGetOptions) != len(defaultGetOptions) {
					t.Errorf("Expected default options to match defaultGetOptions %#v actual %#v", defaultGetOptions, asSysImpl.defaultGetOptions)
				}
				if len(asSysImpl.defaultRegistrationOptions) != len(defaultRegistrationOptions) {
					t.Errorf("Expected default options to match defaultRegistrationOptions %#v actual %#v", defaultRegistrationOptions, asSysImpl.defaultRegistrationOptions)
				}
			}
		}
	})
	t.Run("One Options", func(t *testing.T) {
		sysNoOptions := NewSystemWithOptions(WithGetOptions(GetOptionDoNothing()), WithRegistrationOptions(RegistrationOptionDoNothing()))
		if sysReturned := sysNoOptions.System(); sysReturned == nil {
			t.Errorf("Expected not nil value, actual nil")
		} else if asSysImpl, isSysImpl := sysNoOptions.(*systemImpl); !isSysImpl {
			t.Errorf("Expected implementation of *systemImpl, actual %#v", asSysImpl)
		} else if asSysImpl == nil {
			t.Errorf("Expected not nil implementation of *systemImpl, actual nil")
		} else {
			if isInit, initErr := asSysImpl.isInitialized(); !isInit {
				t.Errorf("Expected initialized object actual not initialized. initErr %#v", initErr)
			}
			if len(asSysImpl.defaultGetOptions) != 1 {
				t.Errorf("Expected 1 default option actual %#v", asSysImpl.defaultGetOptions)
			}
			if len(asSysImpl.defaultRegistrationOptions) != 1 {
				t.Errorf("Expected 1 default option actual %#v", asSysImpl.defaultRegistrationOptions)
			}
		}
	})
}

/*
func TestSystemImpl_RegisterService(t *testing.T) {
	ptrReceiver := &TestServiceB{serviceName: "TestServiceB"}
	var ptrReceiverNil *TestServiceB
	objectReceiver := TestService{serviceName: "TestService"}
	var objectReceiverNil TestService
	var serviceNil ServiceType
	var validService ServiceType = &TestService{"TestServiceAsService"}

	serviceInterfaceName := NewServiceName(DefaultServiceName, AsServiceType[ServiceType]("serviceInterfaceName"))
	var serviceInterfaceSystem, _ = NewDefaultSystem[ServiceType](context.TODO(), serviceInterfaceName, validService)
	serviceSystemMgr := serviceInterfaceSystem.(ServiceManager)
	ptrServiceName := NewServiceName(DefaultServiceName, AsServiceType[*TestServiceB]("ptrServiceName"))
	var ptrServiceSystem, _ = NewDefaultSystem[*TestServiceB](context.TODO(), ptrServiceName, ptrReceiver)
	ptrServiceSystemMgr := ptrServiceSystem.(ServiceManager)
	objServiceName := NewServiceName(DefaultServiceName, AsServiceType[TestService]("objServiceName"))
	var objServiceSystem, _ = NewDefaultSystem[TestService](context.TODO(), objServiceName, objectReceiver)
	objServiceSystemMgr := objServiceSystem.(ServiceManager)

	t11 := TestService{serviceName: "TestServiceAsService2"}
	t12 := &TestServiceB{"TestServiceB2"}
	t13 := TestService{"TestService2"}
	runRegisterTest(t, "ServiceSystemNilName", serviceSystemMgr, nil,
		t11, serviceNil, true, ErrSystemRegisterServiceFailed)
	runRegisterTest(t, "PtrServiceNilName", ptrServiceSystemMgr, nil,
		t12, ptrReceiverNil, true, ErrSystemRegisterServiceFailed)
	runRegisterTest(t, "ObjectServiceNilName", objServiceSystemMgr, nil,
		t13, objectReceiverNil, true, ErrSystemRegisterServiceFailed)

	t21 := TestService{serviceName: "TestServiceAsService21"}
	t22 := TestService{serviceName: "TestServiceAsService22"}
	t23 := TestService{serviceName: "TestServiceAsService23"}
	runRegisterTest(t, "ServiceSystemOriginalNameTestService", serviceSystemMgr, serviceInterfaceName,
		t21, validService, false, nil)
	runRegisterTest(t, "PtrServiceOriginalNameTestService", ptrServiceSystemMgr, ptrServiceName,
		t22, ptrReceiverNil, true, ErrSystemRegisterServiceFailed)
	runRegisterTest(t, "ObjectServiceOriginalNameTestService", objServiceSystemMgr, objServiceName,
		t23, objectReceiver, false, nil)

	t31 := &TestServiceB{"TestServiceB31"}
	t32 := &TestServiceB{"TestServiceB32"}
	t33 := &TestServiceB{"TestServiceB33"}
	runRegisterTest(t, "ServiceSystemOriginalNameTB", serviceSystemMgr, serviceInterfaceName,
		t31, t21, false, nil)
	runRegisterTest(t, "PtrServiceOriginalNameTB", ptrServiceSystemMgr, ptrServiceName,
		t32, ptrReceiver, false, nil)
	runRegisterTest(t, "ObjectServiceOriginalNameTB", objServiceSystemMgr, objServiceName,
		t33, objectReceiverNil, true, ErrSystemRegisterServiceFailed)
}

func TestSystemImpl_UnregisterService(t *testing.T) {
	ptrReceiver := &TestServiceB{serviceName: "TestServiceB"}
	var ptrReceiverNil *TestServiceB
	objectReceiver := TestService{serviceName: "TestService"}
	var objectReceiverNil TestService
	var serviceNil ServiceType
	var validService ServiceType = &TestService{"TestServiceAsService"}

	serviceInterfaceName := NewServiceName(DefaultServiceName, AsServiceType[ServiceType]("serviceInterfaceName"))
	var serviceInterfaceSystem, _ = NewDefaultSystem[ServiceType](context.TODO(), serviceInterfaceName, validService)
	serviceSystemMgr := serviceInterfaceSystem.(ServiceManager)
	ptrServiceName := NewServiceName(DefaultServiceName, AsServiceType[*TestServiceB]("ptrServiceName"))
	var ptrServiceSystem, _ = NewDefaultSystem[*TestServiceB](context.TODO(), ptrServiceName, ptrReceiver)
	ptrServiceSystemMgr := ptrServiceSystem.(ServiceManager)
	objServiceName := NewServiceName(DefaultServiceName, AsServiceType[TestService]("objServiceName"))
	var objServiceSystem, _ = NewDefaultSystem[TestService](context.TODO(), objServiceName, objectReceiver)
	objServiceSystemMgr := objServiceSystem.(ServiceManager)
	runUnRegisterTest(t, "ServiceSystemNilName", serviceSystemMgr, nil, serviceNil, true, ErrSystemUnregisterServiceFailed)
	runUnRegisterTest(t, "PtrServiceNilName", ptrServiceSystemMgr, nil, ptrReceiverNil, true, ErrSystemUnregisterServiceFailed)
	runUnRegisterTest(t, "ObjectServiceNilName", objServiceSystemMgr, nil, objectReceiverNil, true, ErrSystemUnregisterServiceFailed)

	runUnRegisterTest(t, "ServiceSystemOriginalName", serviceSystemMgr, serviceInterfaceName, validService, false, nil)
	runUnRegisterTest(t, "PtrServiceOriginalName", ptrServiceSystemMgr, ptrServiceName, ptrReceiver, false, nil)
	runUnRegisterTest(t, "ObjectServiceOriginalName", objServiceSystemMgr, objServiceName, objectReceiver, false, nil)

}

func TestSystemImpl_ChangeDefault(t *testing.T) {
	t.Run("InvalidName", func(t *testing.T) {
		tService := &TestServiceB{serviceName: "TestSystemImpl_ChangeDefault"}
		var nilService *TestServiceB
		defaultServiceName := NewServiceName(DefaultServiceName, AsServiceType[TestService](TestServiceTypeName))
		var asys, _ = NewDefaultSystem[*TestServiceB](context.TODO(), defaultServiceName, tService)
		oldService, err := asys.ChangeDefault(context.TODO(), nil)
		if oldService != nilService {
			t.Errorf("Expected %#v, Actual %#v", nilService, oldService)
		}
		if err == nil {
			t.Errorf("Expected error actual no error")
		}
		if !errext.Is(err, ErrSystemChangeDefaultFailed) {
			t.Errorf("Expected ErrSystemChangeDefaultFailed, actual %#v", err)
		}
	})
}
func runRegisterTest(t *testing.T, testName string, serviceMgr ServiceManager, serviceName ServiceName, serviceValue ServiceType, expectedServiceValue ServiceType, expectedError bool, expectedErrCode errext.ErrorCode) {
	t.Run(testName, func(t *testing.T) {
		oldService, err := serviceMgr.RegisterService(context.TODO(), serviceName, serviceValue)
		if oldService != expectedServiceValue {
			t.Errorf("Expected %#v oldservice, actual %#v", expectedServiceValue, oldService)
		}
		if expectedError {
			if err == nil {
				t.Errorf("Expected error, actual no error")
			}
			if !errext.Is(err, expectedErrCode) {
				t.Errorf("Expected %#v, actual %#v", expectedErrCode, err)
			}
		} else if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
	})
}

func runUnRegisterTest(t *testing.T, testName string, serviceMgr ServiceManager, serviceName ServiceName, expectedServiceValue ServiceType, expectedError bool, expectedErrCode errext.ErrorCode) {
	t.Run(testName, func(t *testing.T) {
		oldService, err := serviceMgr.UnregisterService(context.TODO(), serviceName)
		if oldService != expectedServiceValue {
			t.Errorf("Expected %#v oldservice, actual %#v", expectedServiceValue, oldService)
		}
		if expectedError {
			if err == nil {
				t.Errorf("Expected error, actual no error")
			}
			if !errext.Is(err, expectedErrCode) {
				t.Errorf("Expected %#v, actual %#v", expectedErrCode, err)
			}
		} else if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
	})
}
*/
