package logger

import (
	"strings"
	"testing"
)

func CommonSimpleMarkerInterfaceTests(t *testing.T, inputMarker Marker, isDefault bool, baseString string, callStack ...string) {
	t.Run("MarkerStringValue", func(t *testing.T) {
		if len(callStack) > 0 {
			if inputMarker.String() != callStack[0] {
				t.Errorf("Marker should contain string %s got %s", callStack[0], inputMarker)
			}
		} else if inputMarker.String() != baseString {
			t.Errorf("Marker should contain string %s got %s", baseString, inputMarker)
		}
	})
	t.Run("MarkerAspects", func(t *testing.T) {
		if isDefault {
			t.Skipf("Skipping check for default since this is expected to fail.")
		}
		expectedAspects := strings.Split(baseString, SimpleMarkerLoggerNameDelimiter)
		markerAspects := inputMarker.GetAspects()
		if markerAspects == nil {
			t.Errorf("Marker should contain aspects %s", inputMarker)
		}
		if len(markerAspects) != len(expectedAspects) {
			t.Errorf("Marker aspects # does not match expected %d got %d", len(expectedAspects), len(markerAspects))
		}
		markerAspectWithDelimit := strings.Join(markerAspects, "_")
		expectedAspectsWithDemit := strings.Join(expectedAspects, "_")
		if markerAspectWithDelimit != expectedAspectsWithDemit {
			t.Errorf("Marker aspects does not match expected %s got %s", expectedAspectsWithDemit, markerAspectWithDelimit)
		}
	})
	t.Run("StackCheck", func(t *testing.T) {
		currentMarker := inputMarker
		totalItems := len(callStack)
		for counter, callStackItem := range callStack {
			if currentMarker.String() != callStackItem {
				t.Errorf("Call stack at level %d does not match expected. expected %s got %s", counter, callStackItem, currentMarker.String())
			}
			var previousMarker = currentMarker.Pop()
			if previousMarker == nil && counter < totalItems-1 {
				t.Errorf("No additional item is available in marker stack at %d. Expected call %s", counter, callStackItem)
				break
			}
			currentMarker = previousMarker
		}
	})
}

func TestDefaultSimpleMarker(t *testing.T) {
	var defaultMarker Marker = &simpleMarker{}
	CommonSimpleMarkerInterfaceTests(t, defaultMarker, true, "")
}

func TestNewSimpleMarker(t *testing.T) {
	var defaultMarker Marker = NewSimpleMarker("pkg1")
	CommonSimpleMarkerInterfaceTests(t, defaultMarker, false, "pkg1")
}

func TestMarkerAppend(t *testing.T) {
	var defaultMarker Marker = NewSimpleMarker("pkg1")
	var appendedMarker = defaultMarker.Append("pkg2")
	var appendedAnotherMarker = appendedMarker.Append("file")
	t.Run("defaultMarker", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, defaultMarker, false, "pkg1.pkg2.file")
	})
	t.Run("appendedMarker", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, appendedMarker, false, "pkg1.pkg2.file")
	})
	t.Run("appendedAnotherMarker", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, appendedAnotherMarker, false, "pkg1.pkg2.file")
	})
}

func TestMarkerAdd(t *testing.T) {
	var defaultMarker Marker = NewSimpleMarker("pkg1")
	var addedMarker = defaultMarker.Add("pkg2", "pkg3")
	var addedAnotherMarker = addedMarker.Add("file")
	var appendedMarker = addedAnotherMarker.Append("file1")
	t.Run("defaultMarker", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, defaultMarker, false, "pkg1")
	})
	t.Run("addedMarker", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, addedMarker, false, "pkg1.pkg2.pkg3")
	})
	t.Run("addedAnotherMarker", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, addedAnotherMarker, false, "pkg1.pkg2.pkg3.file.file1")
	})
	t.Run("appendedMarker", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, appendedMarker, false, "pkg1.pkg2.pkg3.file.file1")
	})
}

func TestNewHierarchicalMarker(t *testing.T) {
	var defaultMarker Marker = NewSimpleMarker("pk1.pkg2")
	CommonSimpleMarkerInterfaceTests(t, defaultMarker, false, "pk1.pkg2")
}

func TestNewHierarchical2LevelMarker(t *testing.T) {
	var defaultMarker Marker = NewSimpleMarker("pk1.pkg2.file1")
	CommonSimpleMarkerInterfaceTests(t, defaultMarker, false, "pk1.pkg2.file1")
}

func TestMethodNoParam(t *testing.T) {
	updatedMarkerValue := NewSimpleMarker("pkg1").Append("pkg2", "file1").AddMethod(nil, "method1NoParam")
	CommonSimpleMarkerInterfaceTests(t, updatedMarkerValue, false, "pkg1.pkg2.file1", "pkg1.pkg2.file1.method1NoParam()")
}

func TestMethodOneParam(t *testing.T) {
	updatedMarkerValue := NewSimpleMarker("pkg1").Append("pkg2", "file1").AddMethod(nil, "method2OneParam", "param1string", "param1value1")
	CommonSimpleMarkerInterfaceTests(t, updatedMarkerValue, false, "pkg1.pkg2.file1", "pkg1.pkg2.file1.method2OneParam(param1string param1value1)")
}

func TestMethodMultiParam(t *testing.T) {
	type testObjectStruct struct {
		value bool
	}
	testObject := testObjectStruct{value: true}
	updatedMarkerValue := NewSimpleMarker("pkg1.file1").AddMethod(nil, "method3MultiParam", "param1string", "param1value1",
		"param2bool", true, "param3number", 1, "param3object", testObject)
	CommonSimpleMarkerInterfaceTests(t, updatedMarkerValue, false, "pkg1.file1", "pkg1.file1.method3MultiParam(param1string param1value1 param2bool true param3number 1 param3object {true})")
}

func TestMethodStackOneLevel(t *testing.T) {
	mainFunction := NewSimpleMarker("pkg1").Append("pkg2", "file1").AddMethod(nil, "method1NoParam")
	levelOneFunction := NewSimpleMarker("pkg1").Append("pkg2", "file1").AddMethod(mainFunction, "method2OneParam", "param1string", "param1value1")
	levelOneFunctionBaseMethod := NewSimpleMarker("pkg1").Append("pkg2", "file1").AddMethod(nil, "method2OneParam", "param1string", "param1value1")
	levelOneFunctionWithPush := levelOneFunctionBaseMethod.Push(mainFunction)
	t.Run("mainFunction", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, mainFunction, false, "pkg1.pkg2.file1", "pkg1.pkg2.file1.method1NoParam()")
	})
	t.Run("levelOneFunction", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, levelOneFunction, false, "pkg1.pkg2.file1", "pkg1.pkg2.file1.method2OneParam(param1string param1value1)", "pkg1.pkg2.file1.method1NoParam()")
	})
	t.Run("levelOneFunctionBaseMethod", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, levelOneFunctionBaseMethod, false, "pkg1.pkg2.file1", "pkg1.pkg2.file1.method2OneParam(param1string param1value1)")
	})
	t.Run("levelOneFunctionWithPush", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, levelOneFunctionWithPush, false, "pkg1.pkg2.file1", "pkg1.pkg2.file1.method2OneParam(param1string param1value1)", "pkg1.pkg2.file1.method1NoParam()")
	})
}

func TestMethodStackTwoLevel(t *testing.T) {
	type testObjectStruct struct {
		value bool
	}
	testObject := testObjectStruct{value: true}
	mainFunction := NewSimpleMarker("pkg1").Append("pkg2", "file1").AddMethod(nil, "method1NoParam")
	levelOneFunction := NewSimpleMarker("pkg1").Append("pkg2", "file1").AddMethod(mainFunction, "method2OneParam", "param1string", "param1value1")
	levelTwoFunction := NewSimpleMarker("pkg1.file1").AddMethod(levelOneFunction, "method3MultiParam", "param1string", "param1value1",
		"param2bool", true, "param3number", 1, "param3object", testObject)
	levelTwoFunctionBaseMethod := NewSimpleMarker("pkg1.file1").AddMethod(nil, "method3MultiParam", "param1string", "param1value1",
		"param2bool", true, "param3number", 1, "param3object", testObject)
	levelTwoFunctionWithPush := levelTwoFunctionBaseMethod.Push(levelOneFunction)

	t.Run("levelTwoFunction", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, levelTwoFunction, false, "pkg1.file1", "pkg1.file1.method3MultiParam(param1string param1value1 param2bool true param3number 1 param3object {true})", "pkg1.pkg2.file1.method2OneParam(param1string param1value1)", "pkg1.pkg2.file1.method1NoParam()")
	})
	t.Run("levelTwoFunctionBaseMethod", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, levelTwoFunctionBaseMethod, false, "pkg1.file1", "pkg1.file1.method3MultiParam(param1string param1value1 param2bool true param3number 1 param3object {true})")
	})
	t.Run("levelTwoFunctionWithPush", func(t *testing.T) {
		CommonSimpleMarkerInterfaceTests(t, levelTwoFunctionWithPush, false, "pkg1.file1", "pkg1.file1.method3MultiParam(param1string param1value1 param2bool true param3number 1 param3object {true})", "pkg1.pkg2.file1.method2OneParam(param1string param1value1)", "pkg1.pkg2.file1.method1NoParam()")
	})
}
