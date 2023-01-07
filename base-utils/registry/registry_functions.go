package registry

import "sync"
import logger "github.com/grinps/go-utils/base-utils/logs"

func (record *registrationRecord[Value]) Get() Value {
	var emptyValue Value
	if record == nil || !record.isInitialized() {
		logger.Warn("registrationRecord.Get() was called on nil registration record or not initialized. This should not happen in normal flow of code")
		return emptyValue
	}
	logger.Log("Locking read lock with deferred unlock on entry", record)
	record.lock.RLock()
	defer deferredRUnlock(record, record.lock)
	return record.value
}

func (record *registrationRecord[Value]) Set(newValue Value) Value {
	var nilValue Value
	if record == nil || !record.isInitialized() {
		logger.Warn("registrationRecord.Set() was called on nil registration record or not initialized. This should not happen in normal flow of code.", "newValue", newValue)
		return nilValue
	}
	logger.Log("Locking lock with deferred unlock on entry", record)
	record.lock.Lock()
	defer deferredUnlock(record, record.lock)
	var oldValue = record.value
	record.value = newValue
	return oldValue
}

func (record *registrationRecord[Value]) isInitialized() bool {
	if record != nil && record.lock != nil && record.valueSet {
		return true
	}
	return false
}

func (registry *registryComparable[T, Key, Value]) Get(key Key) Value {
	var value Value
	if registry == nil {
		logger.Warn("registryComparable.Get was called on nil register. This should not happen in normal flow of code", "key", key)
	} else {
		regRecord := registry.getRegistrationRecord(key)
		logger.Log("Found registration record", regRecord)
		if regRecord != nil {
			logger.Log("Getting value from registration record", regRecord)
			value = regRecord.Get()
		}
	}
	logger.Log("Returning value from registry(", registry, ") for key(", key, ") as ", value)
	return value
}

func (registry *registryComparable[T, Key, Value]) Register(key Key, newValue Value) Value {
	var currentValue Value
	if registry == nil {
		logger.Warn("registryComparable.Register was called on nil register. This should not happen in normal flow of code", "key", key, "newValue", newValue)
	} else {
		regRecord := registry.getRegistrationRecord(key)
		logger.Log("Found registration record", regRecord)
		if regRecord != nil {
			logger.Log("Setting value on registration record", regRecord, "value", newValue)
			currentValue = regRecord.Set(newValue)
		} else {
			logger.Log("Creating registration recordInternal for key", key)
			regRecord := &registrationRecord[Value]{
				valueSet: true,
				value:    newValue,
				lock:     &sync.RWMutex{},
			}
			registry.setRegistrationRecord(key, regRecord)
		}
	}
	return currentValue
}

func (registry *registryComparable[T, Key, Value]) isInitialized() bool {
	if registry != nil && registry.register != nil && registry.lock != nil {
		return true
	}
	return false
}

func (registry *registryComparable[T, Key, Value]) getRegistrationRecord(key Key) *registrationRecord[Value] {
	if registry == nil || !registry.isInitialized() {
		logger.Warn("registryComparable.getRegistrationRecord was called on nil register or not initialized. This should not happen in normal flow of code", "key", key)
		return nil
	}
	var record *registrationRecord[Value] = nil
	logger.Log("Locking read lock with deferred unlock on entry", registry)
	registry.lock.RLock()
	defer deferredRUnlock(registry, registry.lock)
	applicableKey, err := ComparableValueConverter[T](key)
	var nilObject T
	if err != nil {
		logger.Log("Failed to convert key", key, "to map key", registry)
	} else if applicableKeyValue := applicableKey.Unique(); applicableKeyValue != nilObject {
		if recordInternal, ok := registry.register[applicableKey.Unique()]; ok && recordInternal != nil {
			record = recordInternal
			logger.Log("Located existing record on entry", registry)
		}
	} else {
		logger.Log("Key lookup is nil value", nilObject, "for key", key, "on registry", registry)
	}
	logger.Log("Returning record", record)
	return record
}

func (registry *registryComparable[T, Key, Value]) setRegistrationRecord(key Key, record *registrationRecord[Value]) *registrationRecord[Value] {
	var exitingRecord *registrationRecord[Value] = nil
	if registry == nil || !registry.isInitialized() {
		logger.Warn("registryComparable.setRegistrationRecord was called on nil register or not initialized. This should not happen in normal flow of code", "key", key, "record", record)
	} else {
		logger.Log("Locking lock on entry", registry)
		registry.lock.Lock()
		defer deferredUnlock(registry, registry.lock)
		applicableKey, err := ComparableValueConverter[T](key)
		var nilObject T
		if err != nil {
			logger.Log("Failed to convert key", key, "to map key", registry)
		} else if applicableKeyValue := applicableKey.Unique(); applicableKeyValue != nilObject {
			if internalCurrentValue, ok := registry.register[applicableKeyValue]; ok {
				exitingRecord = internalCurrentValue
				logger.Log("Got current value from register ", registry, " for key ", key, " as ", internalCurrentValue)
			}
			registry.register[applicableKeyValue] = record
		} else {
			logger.Log("Key lookup is nil value", nilObject, "for key", key, "on registry", registry)
		}
		logger.Log("Set current value in register ", registry, " for key ", key, " as ", record)
	}
	return exitingRecord
}

func (registry *registryComparable[T, Key, Value]) Unregister(key Key) Value {
	var currentEntry Value
	if registry == nil {
		logger.Warn("registryComparable.Unregister was called on nil register. This should not happen in normal flow of code", "key", key)
	} else {
		logger.Log("Locking lock with deferred unlock on entry", registry)
		registry.lock.Lock()
		defer deferredUnlock(registry, registry.lock)
		applicableKey, err := ComparableValueConverter[T](key)
		var nilObject T
		if err != nil {
			logger.Log("Failed to convert key", key, "to map key", registry)
		} else if applicableKeyValue := applicableKey.Unique(); applicableKeyValue != nilObject {
			if currentEntryInMap, ok := registry.register[applicableKeyValue]; ok {
				logger.Log("Registration record located for key", key, "value", currentEntryInMap)
				delete(registry.register, applicableKeyValue)
				logger.Log("Deleted key from map")
				if currentEntryInMap != nil {
					currentEntry = currentEntryInMap.Get()
					logger.Log("Retrieved current value as", currentEntry)
				}
			} else {
				logger.Log("No registration record was located for key", key)
			}
		} else {
			logger.Log("Key lookup is nil value", nilObject, "for key", key, "on registry", registry)
		}
	}
	return currentEntry
}

func deferredUnlock(registry interface{}, lock *sync.RWMutex) {
	logger.Log("Unlocking lock on entry", registry)
	lock.Unlock()

}

func deferredRUnlock(registry interface{}, lock *sync.RWMutex) {
	logger.Log("Unlocking read lock on entry", registry)
	lock.RUnlock()
}

func NewRegister[Key comparable, Value any]() Register[Key, Value] {
	return &registryComparable[Key, Key, Value]{register: map[Key]*registrationRecord[Value]{}, lock: &sync.RWMutex{}}
}

func NewRegisterWithAnyKey[C comparable, Key CustomKey[C], Value any]() Register[Key, Value] {
	return &registryComparable[C, Key, Value]{register: map[C]*registrationRecord[Value]{}, lock: &sync.RWMutex{}}
}
