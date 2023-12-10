package system

/*
import (
	"context"
	"errors"
	"github.com/grinps/go-utils/errext"
	"testing"
)

func TestGetService(t *testing.T) {
	t.Run("AllNil", func(t *testing.T) {
		service, err := TransformService[*TestServiceB](nil, nil, nil)
		if service != nil {
			t.Errorf("Expected nil namedService actual %#v", service)
		}
		if err == nil {
			t.Errorf("Expected error,actual no error")
		}
		if !errext.Is(err, ErrSystemGetServiceFailed) {
			t.Errorf("Expected error ErrSystemGetServiceFailed, actual %#v", err)
		}
	})

	testServiceBServiceName := NewServiceName("TestName", AsServiceType[*TestServiceB]("TestServiceB"))
	t.Run("serviceNil", func(t *testing.T) {
		service, err := TransformService[*TestServiceB](context.TODO(), testServiceBServiceName, nil)
		if service != nil {
			t.Errorf("Expected nil namedService actual %#v", service)
		}
		if err == nil {
			t.Errorf("Expected error,actual no error")
		}
		if !errext.Is(err, ErrSystemGetServiceFailed) {
			t.Errorf("Expected error ErrSystemGetServiceFailed, actual %#v", err)
		}
	})
	t.Run("ServiceValid", func(t *testing.T) {
		var testService ServiceType = &TestServiceB{serviceName: "ServiceValid"}
		service, err := TransformService[*TestServiceB](context.TODO(), testServiceBServiceName, testService)
		if service == nil {
			t.Errorf("Expected not nil namedService actual nil namedService")
		}
		if err != nil {
			t.Errorf("Expected no error,actual %#v", err)
		}
		if testService != service {
			t.Errorf("Expected namedService to match actually does not match, returned namedService %#v", service)
		}
	})
	testErrorServiceName := &AltServiceName[*TestServiceB]{name: "error", st: AsServiceType[*TestServiceB]("AnotherTestService")}
	t.Run("ServiceErrGetServiceType", func(t *testing.T) {
		var testService ServiceType = &TestServiceB{serviceName: "ServiceValid"}
		service, err := TransformService[*TestServiceB](context.TODO(), testErrorServiceName, testService)
		if service != nil {
			t.Errorf("Expected nil namedService actual %#v", service)
		}
		if err == nil {
			t.Errorf("Expected error,actual no error")
		}
		if !errext.Is(err, ErrSystemGetServiceFailed) {
			t.Errorf("Expected ErrSystemGetServiceFailed, actual %#v", err)
		}
	})
}

func TestNewDefaultSystem(t *testing.T) {
	t.Run("NilValues", func(t *testing.T) {
		sys, err := NewDefaultSystem[*TestServiceB](nil, nil, nil)
		if sys != nil {
			t.Errorf("Expected nil system, actual system %#v", sys)
		}
		if err == nil {
			t.Errorf("Expected error actual no error")
		}
		if !errext.Is(err, ErrNewDefaultSystem) {
			t.Errorf("Expected error ErrNewDefaultSystem, actual %#v", err)
		}
		t.Logf("Error %#v", err)
	})
	t.Run("NilService", func(t *testing.T) {
		sName := NewServiceName[*TestServiceB]("TestNewDefaultSystemNilService", AsServiceType[*TestServiceB]("TestNewDefaultSystemType"))
		sys, err := NewDefaultSystem[*TestServiceB](context.TODO(),
			sName,
			nil)
		if sys != nil {
			t.Errorf("Expected nil system, actual %#v", sys)
		}
		if err == nil {
			t.Errorf("Expected error actual no error")
		}
		if !errext.Is(err, ErrNewDefaultSystem) {
			t.Errorf("Expected ErrNewDefaultSystem, actual %#v", err)
		}
	})
}

func TestRegisterService(t *testing.T) {
	t.Run("NilValue", func(t *testing.T) {
		err := RegisterService[*TestServiceB](nil, nil, nil, nil)
		if err == nil {
			t.Errorf("Expected error, actual no error")
		}
		if !errext.Is(err, ErrSystemRegisterServiceFailed) {
			t.Errorf("Expected ErrSystemRegisterServiceFailed, actual %#v", err)
		}
	})
	t.Run("NotSystemMgr", func(t *testing.T) {
		sName := NewServiceName[*TestServiceB]("TestNewDefaultSystemNilService", AsServiceType[*TestServiceB]("TestNewDefaultSystemType"))
		asystem := &altSystem{}
		aService := &TestServiceB{serviceName: "NotSystemMgr"}
		err := RegisterService[*TestServiceB](context.TODO(), asystem, sName, aService)
		if err == nil {
			t.Errorf("Expected error, actual no error")
		}
		if !errext.Is(err, ErrSystemRegisterServiceFailed) {
			t.Errorf("Expected ErrSystemRegisterServiceFailed, actual %#v", err)
		}
	})

}

type AltServiceName[T ServiceType] struct {
	name string
	st   ServiceType[T]
}

func (name *AltServiceName[T]) String() string {
	return name.name
}

func (name *AltServiceName[T]) Equals(dst ServiceName) bool {
	if dst == name {
		return true
	}
	if dst != nil && name != nil {
		return dst.String() == name.name
	}
	return false
}

func (name *AltServiceName[T]) GetServiceType(ctx context.Context) (ServiceType[T], error) {
	if name.name == "error" {
		return nil, errors.New("Random Error")
	}
	return name.st, nil
}

type altSystem struct {
}

func (system *altSystem) System() System {
	return system
}
*/
