package system

import (
	"fmt"
	"github.com/grinps/go-utils/errext"
	"reflect"
	"testing"
)

type TestInterface interface {
	aFunc(input int) string
}

type TestFunction func(input int) string

func (f TestFunction) aFunc(input int) string {
	return f(input)
}

func TFunctionImpl(input int) string {
	return fmt.Sprintf("Func:%d", input)
}

type TestStruct struct {
	input int
}

func (tStruct TestStruct) aFunc(input int) string {
	return fmt.Sprintf("%d:%d", tStruct.input, input)
}

type TestStructWithAs struct {
	someInputVal int
}

func (tStruct TestStructWithAs) As(output any) bool {
	if outVal := reflect.ValueOf(output); !outVal.IsNil() {
		outVal.Elem().Set(reflect.ValueOf(TestStruct{input: tStruct.someInputVal}))
		return true
	}
	return false
}

func TestAsType(t *testing.T) {
	var intVal int // intX, rune,byte
	var boolVal bool
	var stringVal string
	var stringPtr *string
	var stringArr []string
	var funcVal = func(intVal int) string { return fmt.Sprintf("funcVal(%d)", intVal) }
	var tInterface TestInterface = struct {
		TestInterface
	}{}
	var tStruct = TestStruct{input: 1}
	var tFunction = TFunctionImpl
	t.Run("NilInput", func(t *testing.T) {
		runTest(t, "int value", func() error { return AsType(nil, &intVal) }, true, ErrAsFailedInvalidInput)
		runTest(t, "bool value", func() error { return AsType(nil, &boolVal) }, true, ErrAsFailedInvalidInput)
		runTest(t, "string value", func() error { return AsType(nil, &stringVal) }, true, ErrAsFailedInvalidInput)
		runTest(t, "string pointer value", func() error { return AsType(nil, &stringPtr) }, true, ErrAsFailedInvalidInput)
		runTest(t, "function value", func() error { return AsType(nil, &funcVal) }, true, ErrAsFailedInvalidInput)
		runTest(t, "string[] value", func() error { return AsType(nil, &stringArr) }, true, ErrAsFailedInvalidInput)
		runTest(t, "anonymous interface value", func() error { return AsType(nil, &tInterface) }, true, ErrAsFailedInvalidInput)
		runTest(t, "struct value", func() error { return AsType(nil, &tStruct) }, true, ErrAsFailedInvalidInput)
		runTest(t, "function interface value", func() error { return AsType(nil, &tFunction) }, true, ErrAsFailedInvalidInput)
	})
	t.Run("NilOutput", func(t *testing.T) {
		runTest(t, "int value", func() error { return AsType(intVal, nil) }, true, ErrAsFailedInvalidInput)
		runTest(t, "bool value", func() error { return AsType(boolVal, nil) }, true, ErrAsFailedInvalidInput)
		runTest(t, "string value", func() error { return AsType(stringVal, nil) }, true, ErrAsFailedInvalidInput)
		runTest(t, "string pointer value", func() error { return AsType(stringPtr, nil) }, true, ErrAsFailedInvalidInput)
		runTest(t, "function value", func() error { return AsType(funcVal, nil) }, true, ErrAsFailedInvalidInput)
		runTest(t, "string[] value", func() error { return AsType(stringArr, nil) }, true, ErrAsFailedInvalidInput)
		runTest(t, "anonymous interface value", func() error { return AsType(tInterface, nil) }, true, ErrAsFailedInvalidInput)
		runTest(t, "struct value", func() error { return AsType(tStruct, nil) }, true, ErrAsFailedInvalidInput)
		runTest(t, "function interface value", func() error { return AsType(tFunction, nil) }, true, ErrAsFailedInvalidInput)
	})
	t.Run("Int2", func(t *testing.T) {
		runTest(t, "int2*int", func() error { return AsType(1, &intVal) }, false, nil)
		runTest(t, "int2int", func() error { return AsType(1, intVal) }, true, ErrAsFailedInvalidInput)
		runTest(t, "int2*string", func() error { return AsType(1, &stringVal) }, true, ErrAsFailedTransformation)
		runTest(t, "int2*bool", func() error { return AsType(1, &boolVal) }, true, ErrAsFailedTransformation)
		var listOfInt []int
		runTest(t, "int2*bool", func() error { return AsType(1, &listOfInt) }, true, ErrAsFailedTransformation)
	})
	t.Run("String2", func(t *testing.T) {
		runTest(t, "string value", func() error { return AsType("ThisIsATest", &stringVal) }, false, nil)
		var aStringInterface any = string("AnOldValue")
		runTest(t, "string as interface var", func() error { return AsType("ThisIsATest", &aStringInterface) }, false, nil)
		runTest(t, "string pointer value", func() error { return AsType("ThisIsATest", stringPtr) }, true, ErrAsFailedInvalidInput)
		runTest(t, "string[] value", func() error { return AsType("ThisIsATest", &stringArr) }, true, ErrAsFailedTransformation)
		runTest(t, "function value", func() error { return AsType("ThisIsATest", &funcVal) }, true, ErrAsFailedTransformation)
		runTest(t, "anonymous interface value", func() error { return AsType("ThisIsATest", &tInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "struct value", func() error { return AsType("ThisIsATest", &tStruct) }, true, ErrAsFailedTransformation)
		runTest(t, "function interface value", func() error { return AsType("ThisIsATest", &tFunction) }, true, ErrAsFailedTransformation)
	})
	t.Run("2Interface", func(t *testing.T) {
		var outputInterface TestInterface
		// TODO: Make this work, currently ERROR
		runTest(t, "function value", func() error {
			return AsType(funcVal, &outputInterface)
		}, true, ErrAsFailedTransformation)
		var funcAsTestFunction TestFunction = funcVal
		runTest(t, "function value as interface", func() error { return AsType(funcAsTestFunction, &outputInterface) }, false, nil)
		runTest(t, "anonymous interface value", func() error { return AsType(tInterface, &outputInterface) }, false, nil)
		runTest(t, "struct value", func() error { return AsType(tStruct, &outputInterface) }, false, nil)
		// TODO: Make this work, currently ERROR
		runTest(t, "function interface value", func() error {
			return AsType(tFunction, &outputInterface)
		}, true, ErrAsFailedTransformation)
		var tFunctionAsTestFunction TestFunction = tFunction
		runTest(t, "function interface value as interface", func() error { return AsType(tFunctionAsTestFunction, &outputInterface) }, false, nil)

		runTest(t, "int value", func() error { return AsType(1, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "bool value", func() error { return AsType(true, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "string value", func() error { return AsType("ThisIsATest", &outputInterface) }, true, ErrAsFailedTransformation)
		var aString = "SomeRandomString"
		var anInterface interface{} = aString
		stringPtr = &aString
		runTest(t, "string as interface var", func() error { return AsType(anInterface, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "string pointer value", func() error { return AsType(stringPtr, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "string[] value", func() error { return AsType(stringArr, &outputInterface) }, true, ErrAsFailedTransformation)
	})
	t.Run("2FunctionInterface", func(t *testing.T) {
		var outputInterface TestFunction
		runTest(t, "function value", func() error { return AsType(funcVal, &outputInterface) }, false, nil)
		var funcAsTestFunction TestFunction = funcVal
		runTest(t, "function value as interface", func() error { return AsType(funcAsTestFunction, &outputInterface) }, false, nil)
		runTest(t, "function interface value", func() error { return AsType(tFunction, &outputInterface) }, false, nil)
		var tFunctionAsTestFunction TestFunction = tFunction
		runTest(t, "function interface value as interface", func() error { return AsType(tFunctionAsTestFunction, &outputInterface) }, false, nil)

		runTest(t, "struct value", func() error { return AsType(tStruct, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "anonymous interface value", func() error { return AsType(tInterface, &outputInterface) }, true, ErrAsFailedTransformation)

		runTest(t, "int value", func() error { return AsType(1, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "bool value", func() error { return AsType(true, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "string value", func() error { return AsType("ThisIsATest", &outputInterface) }, true, ErrAsFailedTransformation)
		var aString = "SomeRandomString"
		var anInterface interface{} = aString
		stringPtr = &aString
		runTest(t, "string as interface var", func() error { return AsType(anInterface, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "string pointer value", func() error { return AsType(stringPtr, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "string[] value", func() error { return AsType(stringArr, &outputInterface) }, true, ErrAsFailedTransformation)
	})
	t.Run("2InterfaceInitializedWithStruct", func(t *testing.T) {
		var outputInterface TestInterface = tStruct
		//TODO: Make this work
		runTest(t, "function value", func() error { return AsType(funcVal, &outputInterface) }, true, ErrAsFailedTransformation)
		var funcAsTestFunction TestFunction = funcVal
		runTest(t, "function value as interface", func() error { return AsType(funcAsTestFunction, &outputInterface) }, false, nil)
		//TODO: Make this work
		runTest(t, "function interface value", func() error { return AsType(tFunction, &outputInterface) }, true, ErrAsFailedTransformation)
		var tFunctionAsTestFunction TestFunction = tFunction
		runTest(t, "function interface value as interface", func() error { return AsType(tFunctionAsTestFunction, &outputInterface) }, false, nil)
		runTest(t, "struct value", func() error { return AsType(tStruct, &outputInterface) }, false, nil)
		runTest(t, "anonymous interface value", func() error { return AsType(tInterface, &outputInterface) }, false, nil)

		runTest(t, "int value", func() error { return AsType(1, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "bool value", func() error { return AsType(true, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "string value", func() error { return AsType("ThisIsATest", &outputInterface) }, true, ErrAsFailedTransformation)
		var aString = "SomeRandomString"
		var anInterface interface{} = aString
		stringPtr = &aString
		runTest(t, "string as interface var", func() error { return AsType(anInterface, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "string pointer value", func() error { return AsType(stringPtr, &outputInterface) }, true, ErrAsFailedTransformation)
		runTest(t, "string[] value", func() error { return AsType(stringArr, &outputInterface) }, true, ErrAsFailedTransformation)
	})
	t.Run("UseAsToTransform", func(t *testing.T) {
		var outputInterface TestInterface
		err := AsType(TestStructWithAs{someInputVal: 123}, &outputInterface)
		if err != nil {
			t.Errorf("Expected success actual %#v", err)
		} else if funcOut := outputInterface.aFunc(1); funcOut != "123:1" {
			t.Errorf("Expected 123:1, Actual %s", funcOut)
		}
	})
}

func TestAs(t *testing.T) {
	t.Run("nilInputTestFunc", func(t *testing.T) {
		output := As[TestFunction, TestInterface](nil)
		//NOTE: Check the comparison being performed.
		if output.(TestFunction) != nil {
			t.Errorf("Expected nil output, actual %#v", output)
		}
	})
	t.Run("defaultInputTestStruct", func(t *testing.T) {
		output := As[TestStruct, TestInterface](TestStruct{})
		if output == nil {
			t.Errorf("Expected not nil output, actual nil")
		} else if outString := output.aFunc(1); outString != "0:1" {
			t.Errorf("Expected value 0:1, actual %s", outString)
		}
	})
	t.Run("ValidTestFunction", func(t *testing.T) {
		var aTestFunction TestFunction = TFunctionImpl
		output := As[TestFunction, TestInterface](aTestFunction)
		if output == nil {
			t.Errorf("Expected not nil output, actual nil")
		} else if outString := output.aFunc(1); outString != "Func:1" {
			t.Errorf("Expected value Func:1, actual %s", outString)
		}
	})
}

func TestAsE(t *testing.T) {
	t.Run("nilInputTestFunc", func(t *testing.T) {
		output, err := AsE[TestFunction, TestInterface](nil)
		if err != nil {
			t.Errorf("Expected no error, actual err %#v", err)
		}
		//NOTE: Check the comparison being performed.
		if output.(TestFunction) != nil {
			t.Errorf("Expected nil output, actual %#v", output)
		}
	})
	t.Run("ValidTestFunction", func(t *testing.T) {
		var aTestFunction TestFunction = TFunctionImpl
		output, err := AsE[TestFunction, TestInterface](aTestFunction)
		if err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		}
		if output == nil {
			t.Errorf("Expected not nil output, actual nil")
		} else if outString := output.aFunc(1); outString != "Func:1" {
			t.Errorf("Expected value Func:1, actual %s", outString)
		}
	})
	t.Run("int2Interface", func(t *testing.T) {
		_, err := AsE[int, TestInterface](1)
		if err == nil {
			t.Errorf("Expected error, actual no error")
		} else if !errext.Is(err, ErrAsFailedTransformation) {
			t.Errorf("Expected ErrAsFailedTransformation, actual %#v", err)
		}
	})
}

func TestAsP(t *testing.T) {
	t.Run("ValidTestFunction", func(t *testing.T) {
		defer func() {
			if panErr := recover(); panErr != nil {
				t.Errorf("Expected no error, actual %#v", panErr)
			}
		}()
		var aTestFunction TestFunction = TFunctionImpl
		output := AsP[TestFunction, TestInterface](aTestFunction)
		if output == nil {
			t.Errorf("Expected not nil output, actual nil")
		} else if outString := output.aFunc(1); outString != "Func:1" {
			t.Errorf("Expected value Func:1, actual %s", outString)
		}
	})
	t.Run("int2Interface", func(t *testing.T) {
		defer func() {
			if panErr := recover(); panErr != nil {
				if asErr, isErr := panErr.(error); isErr {
					if !errext.Is(asErr, ErrAsFailedTransformation) {
						t.Errorf("Expected ErrAsFailedTransformation, actual %#v", asErr)
					}
				} else {
					t.Errorf("Expected recovered value of type error, actual %#v", panErr)
				}
			} else {
				t.Errorf("Expected panic error, actual no panic")
			}
		}()
		_ = AsP[int, TestInterface](1)
	})
}

func TestAsB(t *testing.T) {
	t.Run("ValidTestFunction", func(t *testing.T) {
		var aTestFunction TestFunction = TFunctionImpl
		output, ok := AsB[TestFunction, TestInterface](aTestFunction)
		if !ok {
			t.Errorf("Expected no issues, actual it failed.")
		}
		if output == nil {
			t.Errorf("Expected not nil output, actual nil")
		} else if outString := output.aFunc(1); outString != "Func:1" {
			t.Errorf("Expected value Func:1, actual %s", outString)
		}
	})
	t.Run("int2Interface", func(t *testing.T) {
		_, ok := AsB[int, TestInterface](1)
		if ok {
			t.Errorf("Expected error but actual ok")
		}
	})
}

func runTest(t *testing.T, testName string, aFunction func() error, expectedErr bool, errorCode errext.ErrorCode) {
	t.Run(testName, func(t *testing.T) {
		transformErr := aFunction()
		if expectedErr && transformErr == nil {
			t.Errorf("Expected error but actual nil")
		} else if !expectedErr && transformErr != nil {
			t.Errorf("Expected no error but actual %#v", transformErr)
		}
		if transformErr != nil && errorCode != nil && !errext.Is(transformErr, errorCode) {
			t.Errorf("Expected error of type %v but actual %#v", errorCode, transformErr)
		}
	})
}
