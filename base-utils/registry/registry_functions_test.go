package registry

import (
	"math/rand"
	"sync"
	"testing"
	"time"
)

type compareFunction[Key comparable] func(key Key, currentValue interface{}, expectedValue interface{}) (interface{}, bool)

type operationFunction[Key comparable, Value any] func(registry *Register[Key, Value], key Key, value interface{}) interface{}

type Test2Key bool

type KeyType int

const (
	STRING   = 1
	TESTKEY  = 2
	TEST2KEY = 3
)

type TestKey struct {
	keyName string
}

func (key TestKey) String() string {
	return key.keyName
}

type structValue struct {
	valueString string
}

type testCase struct {
	name  string
	value interface{}
}

var keyTest2Key Test2Key = true

var VALUES = []testCase{
	{"nil value", nil},
	{"primitive bool", false},
	{"primitive string", "VALUE_A"},
	{"array", [3]string{"VAL1", "VAL2", "VAL3"}},
	{"Struct value", structValue{valueString: "VALUE_B"}},
	//{"Struct slice", []structValue{{valueString: "VALUE_C"}, {valueString: "VALUE_D"}}}, // panics since non-nil value can not be compared
	{"Struct slice pointer", &[]structValue{{valueString: "VALUE_E"}, {valueString: "VALUE_F"}}},
	//{"Function", func() string { return "NewValue" }}, // panics since non-nil value can not be compared
}

var KEYS = map[KeyType][]testCase{
	TESTKEY: {
		{"Empty struct Key", &TestKey{}},
		{"Struct Key", &TestKey{"Key3"}},
	},
	TEST2KEY: {{"Primitive bool key", keyTest2Key}},
	STRING:   {{"String key", "test"}},
}

var nilTestKey *TestKey
var nilTest2Key Test2Key = false
var NilKey = map[KeyType][]testCase{
	TESTKEY:  {{"TESTKEY Nil key", nilTestKey}},
	TEST2KEY: {{"TEST2KEY Nil key", nilTest2Key}},
	STRING:   {{"STRING Nil key", ""}},
}

func TestRegister_NilRegister(t *testing.T) {
	var nilRegister *Register[*TestKey, any] = nil
	t.Run("Nil Register Get Operation", func(t *testing.T) {
		value := nilRegister.Get(nil)
		if value != nil {
			t.Error("Failed test. Expected", nil, "Actual", value)
		}
	})
	t.Run("Nil Register register operation", func(t *testing.T) {
		value := nilRegister.Register(nil, nil)
		if value != nil {
			t.Error("Failed test. Expected", nil, "Actual", value)
		}
	})
	t.Run("Nil Register Get registration record Operation", func(t *testing.T) {
		value := nilRegister.getRegistrationRecord(nil)
		if value != nil {
			t.Error("Failed test. Expected", nil, "Actual", value)
		}
	})
	t.Run("Nil Register Set registration record Operation", func(t *testing.T) {
		value := nilRegister.setRegistrationRecord(nil, nil)
		if value != nil {
			t.Error("Failed test. Expected", nil, "Actual", value)
		}
	})
	t.Run("Nil Register Unregister Operation", func(t *testing.T) {
		value := nilRegister.Unregister(nil)
		if value != nil {
			t.Error("Failed test. Expected", nil, "Actual", value)
		}
	})
}

func TestRegister_DefaultRegister(t *testing.T) {
	//t.Setenv("TRACE_LOG_UTIL_ENABLE", "TRUE")
	//logger.Initialize()
	t.Run("Default Register Get Operation", func(t *testing.T) {
		var regTestKey = &Register[*TestKey, any]{}
		runAllKeyValue[*TestKey, any](t, TESTKEY, regTestKey, t.Name(), getOperation[*TestKey, any], compareNil[*TestKey])
		var regT2 = &Register[Test2Key, any]{}
		runAllKeyValue[Test2Key](t, TEST2KEY, regT2, t.Name(), getOperation[Test2Key, any], compareNil[Test2Key])
		var regString = &Register[string, any]{}
		runAllKeyValue[string](t, STRING, regString, t.Name(), getOperation[string, any], compareNil[string])
	})
	t.Run("Default Register Set Get Operation", func(t *testing.T) {
		var regTestKey = &Register[*TestKey, any]{}
		runAllKeyValue[*TestKey](t, TESTKEY, regTestKey, t.Name(), setGetOperation[*TestKey, any], compareNil[*TestKey])
		var regT2 = &Register[Test2Key, any]{}
		runAllKeyValue[Test2Key](t, TEST2KEY, regT2, t.Name(), setGetOperation[Test2Key, any], compareNil[Test2Key])
		var regString = &Register[string, any]{}
		runAllKeyValue[string](t, STRING, regString, t.Name(), setGetOperation[string, any], compareNil[string])
	})
}

func TestNewRegister(t *testing.T) {
	t.Run("Empty Register Get operation", func(t *testing.T) {
		var regTestKey = NewRegister[*TestKey, any]()
		runAllKeyValue[*TestKey](t, TESTKEY, regTestKey, t.Name(), getOperation[*TestKey, any], compareNil[*TestKey])
		var regT2 = NewRegister[Test2Key, any]()
		runAllKeyValue[Test2Key](t, TEST2KEY, regT2, t.Name(), getOperation[Test2Key, any], compareNil[Test2Key])
		var regString = NewRegister[string, any]()
		runAllKeyValue[string](t, STRING, regString, t.Name(), getOperation[string, any], compareNil[string])
	})
	t.Run("Empty Register Set Get operation", func(t *testing.T) {
		var regTestKey = NewRegister[*TestKey, any]()
		runAllKeyValue[*TestKey](t, TESTKEY, regTestKey, t.Name(), setGetOperation[*TestKey, any], compareEquals[*TestKey])
		var regT2 = NewRegister[Test2Key, any]()
		runAllKeyValue[Test2Key](t, TEST2KEY, regT2, t.Name(), setGetOperation[Test2Key, any], compareEquals[Test2Key])
		var regString = NewRegister[string, any]()
		runAllKeyValue[string](t, STRING, regString, t.Name(), setGetOperation[string, any], compareEquals[string])
	})
	t.Run("Empty Register Nil Key Set Get Operation", func(t *testing.T) {
		var regTestKey = NewRegister[*TestKey, any]()
		for _, value := range VALUES {
			runTest[*TestKey](t, regTestKey, ":*TestKey", NilKey[TESTKEY][0], value, setGetOperation[*TestKey, any], compareNil[*TestKey])
		}
		var regT2 = NewRegister[Test2Key, any]()
		for _, value := range VALUES {
			runTest[Test2Key](t, regT2, ":Test2Key", NilKey[TEST2KEY][0], value, setGetOperation[Test2Key, any], compareNil[Test2Key])
		}
		var regString = NewRegister[string, any]()
		for _, value := range VALUES {
			runTest[string](t, regString, ":string", NilKey[STRING][0], value, setGetOperation[string, any], compareNil[string])
		}
	})
	t.Run("Multiple Get operation", func(t *testing.T) {
		var regString = NewRegister[string, any]()
		values := createRandomKeyValues(STRING, regString)
		runAllKeyValue[string](t, STRING, regString, t.Name(), getOperation[string, any], func(key string, currentValue interface{}, expectedValue interface{}) (interface{}, bool) {
			if currentValue == values[key] {
				return values[key], true
			}
			return values[key], false
		})
		var regT2 = NewRegister[Test2Key, any]()
		valuesT2 := createRandomKeyValues(TEST2KEY, regT2)
		runAllKeyValue[Test2Key](t, TEST2KEY, regT2, t.Name(), getOperation[Test2Key, any], func(key Test2Key, currentValue interface{}, expectedValue interface{}) (interface{}, bool) {
			if currentValue == valuesT2[key] {
				return valuesT2[key], true
			}
			return valuesT2[key], false
		})
		var regTestKey = NewRegister[*TestKey, any]()
		valTKey := createRandomKeyValues(TESTKEY, regTestKey)
		runAllKeyValue[*TestKey](t, TESTKEY, regTestKey, t.Name(), getOperation[*TestKey, any], func(key *TestKey, currentValue interface{}, expectedValue interface{}) (interface{}, bool) {
			if currentValue == valTKey[key] {
				return valTKey[key], true
			}
			return valTKey[key], false
		})

	})
	t.Run("Multiple Set and Get operation", func(t *testing.T) {
		var regTestKey = NewRegister[*TestKey, any]()
		createRandomKeyValues(TESTKEY, regTestKey)
		runAllKeyValue[*TestKey](t, TESTKEY, regTestKey, t.Name(), setGetOperation[*TestKey, any], compareEquals[*TestKey])
		var regT2 = NewRegister[Test2Key, any]()
		createRandomKeyValues(TEST2KEY, regT2)
		runAllKeyValue[Test2Key](t, TEST2KEY, regT2, t.Name(), setGetOperation[Test2Key, any], compareEquals[Test2Key])
		var regString = NewRegister[string, any]()
		createRandomKeyValues(STRING, regString)
		runAllKeyValue[string](t, STRING, regString, t.Name(), setGetOperation[string, any], compareEquals[string])
	})
}

func TestRegister_Unregister(t *testing.T) {
	t.Run("Single Unregister operation", func(t *testing.T) {
		var registry = NewRegister[*TestKey, any]()
		values := createRandomKeyValues(TESTKEY, registry)
		for _, key := range KEYS[TESTKEY] {
			var tKey *TestKey = key.value.(*TestKey)
			currentValue := registry.Unregister(tKey)
			if currentValue != values[tKey] {
				t.Error("Unregister failed for key", key, "Expected Value", values[tKey], "Actual Values", currentValue)
			}
		}
	})
	t.Run("Multiple Unregister operation", func(t *testing.T) {
		var registry = NewRegister[*TestKey, any]()
		createRandomKeyValues(TESTKEY, registry)
		for counter := 0; counter < 10; counter++ {
			for _, key := range KEYS[TESTKEY] {
				var tKey *TestKey = key.value.(*TestKey)
				registry.Unregister(tKey)
			}
		}
		for _, key := range KEYS[TESTKEY] {
			var tKey *TestKey = key.value.(*TestKey)
			currentValue := registry.Unregister(tKey)
			if currentValue != nil {
				t.Error("Multiple unregister failed for key", key, "Expected Value", nil, "Actual Values", currentValue)
			}
		}

	})
}

func TestRegister_MultiThread(b *testing.T) {
	var registry = NewRegister[*TestKey, any]()
	values := createRandomKeyValues(TESTKEY, registry)
	random := rand.New(rand.NewSource(time.Now().Unix()))
	var waitGroup sync.WaitGroup
	var registerAndGetCounter = 0
	var unregisterAndGetCounter = 0
	for counter := 0; counter < 10; counter++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for settingCounter := 0; settingCounter < 100; settingCounter++ {
				keyIndex := random.Intn(len(KEYS[TESTKEY]))
				valueOption := random.Intn(2)
				var applicableValue interface{} = nil
				if valueOption == 0 {
					var tValue *TestKey = KEYS[TESTKEY][keyIndex].value.(*TestKey)
					applicableValue = values[tValue]
				} else {
					applicableValue = VALUES[0].value
				}
				registry.Register(KEYS[TESTKEY][keyIndex].value.(*TestKey), applicableValue)
				time.Sleep(0)
				newValue := registry.Get(KEYS[TESTKEY][keyIndex].value.(*TestKey))
				if applicableValue != newValue {
					registerAndGetCounter++
				}
			}
		}()
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for settingCounter := 0; settingCounter < 100; settingCounter++ {
				keyIndex := random.Intn(len(KEYS[TESTKEY]))
				returnedValue := registry.Get(KEYS[TESTKEY][keyIndex].value.(*TestKey))
				if !(returnedValue == values[KEYS[TESTKEY][keyIndex].value.(*TestKey)]) && !(returnedValue == VALUES[0].value) {
					b.Error("Failed test", "Key", KEYS[TESTKEY][keyIndex].value, "Actual", returnedValue, "Expected", values[KEYS[TESTKEY][keyIndex].value.(*TestKey)])
					break
				}
			}
		}()
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for settingCounter := 0; settingCounter < 100; settingCounter++ {
				keyIndex := random.Intn(len(KEYS[TESTKEY]))
				returnedValue := registry.Unregister(KEYS[TESTKEY][keyIndex].value.(*TestKey))
				if !(returnedValue == values[KEYS[TESTKEY][keyIndex].value.(*TestKey)]) && !(returnedValue == VALUES[0].value) && returnedValue != nil {
					b.Error("Failed test", "Key", KEYS[TESTKEY][keyIndex].value, "Actual", returnedValue, "Expected", values[KEYS[TESTKEY][keyIndex].value.(*TestKey)], "or", nil, "or", returnedValue == VALUES[0].value)
					break
				}
				time.Sleep(0)
				nilValue := registry.Get(KEYS[TESTKEY][keyIndex].value.(*TestKey))
				if nilValue != nil {
					unregisterAndGetCounter++
				}
			}
		}()
	}
	waitGroup.Wait()
	b.Log("Registry updates between Register & get calls happened", registerAndGetCounter, "times and between unregister and get happened", unregisterAndGetCounter, "times")

}

func runAllKeyValue[Key comparable, Value any](t *testing.T, registryType KeyType, registry *Register[Key, Value], testType string, operation operationFunction[Key, Value], compare compareFunction[Key]) {
	for _, key := range KEYS[registryType] {
		for _, value := range VALUES {
			runTest(t, registry, testType, key, value, operation, compare)
		}
	}
}

func createRandomKeyValues[Key comparable, Value any](registryType KeyType, registry *Register[Key, Value]) map[Key]Value {
	numberOfValues := len(VALUES)
	random := rand.New(rand.NewSource(time.Now().Unix()))
	var valueSequence = make(map[Key]Value)
	for _, key := range KEYS[registryType] {
		value := VALUES[random.Intn(numberOfValues)]
		registry.Register(key.value.(Key), value.value.(Value))
		valueSequence[key.value.(Key)] = value.value.(Value)
	}
	return valueSequence
}

func runTest[Key comparable, Value any](t *testing.T, registry *Register[Key, Value], testType string, key testCase, value testCase, operation operationFunction[Key, Value], compare compareFunction[Key]) {
	t.Run(testType+":"+key.name+"("+value.name+")", func(t *testing.T) {
		var nilKey Key
		if key.value == nilKey {
			retrievedValue := operation(registry, nilKey, value.value)
			if expectedValue, equals := compare(nilKey, retrievedValue, value.value); !equals {
				t.Error("Failed test ", key.name, "(", value.value, ") registry", registry, "Expected", expectedValue, "Actual", retrievedValue)
			}
		} else {
			retrievedValue := operation(registry, key.value.(Key), value.value)
			if expectedValue, equals := compare(key.value.(Key), retrievedValue, value.value); !equals {
				t.Error("Failed test ", key.name, "(", value.value, ") registry", registry, "Expected", expectedValue, "Actual", retrievedValue)
			}
		}
	})
}

func compareEquals[Key comparable](key Key, value interface{}, setValue interface{}) (interface{}, bool) {
	if value == setValue {
		return setValue, true
	}
	return setValue, false
}

func compareNil[Key comparable](key Key, value interface{}, expectedValue interface{}) (interface{}, bool) {
	if value == nil {
		return nil, true
	}
	return nil, false
}

func getOperation[Key comparable, Value any](registry *Register[Key, Value], key Key, value interface{}) interface{} {
	return registry.Get(key)
}

func setGetOperation[Key comparable, Value any](registry *Register[Key, Value], key Key, value Value) interface{} {
	registry.Register(key, value)
	return registry.Get(key)
}

func unregisterOperation[Key comparable, Value any](registry *Register[Key, Value], key Key, value interface{}) interface{} {
	return registry.Unregister(key)
}

func unRegisterAndGet[Key comparable, Value any](registry *Register[Key, Value], key Key, value interface{}) interface{} {
	registry.Unregister(key)
	return registry.Get(key)
}
