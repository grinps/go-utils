package registry

import "sync"
import logger "github.com/grinps/go-utils/base-utils/logs"

func (record *registrationRecord[Key, Value]) Get() Value {
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

func (record *registrationRecord[Key, Value]) Set(newValue Value) Value {
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

func (record *registrationRecord[Key, Value]) isInitialized() bool {
	var defaultKeyValue Key
	if record != nil && record.lock != nil && record.key != defaultKeyValue {
		return true
	}
	return false
}

func (registry *Register[Key, Value]) Get(key Key) Value {
	var value Value
	if registry == nil {
		logger.Warn("Register.Get was called on nil register. This should not happen in normal flow of code", "key", key)
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

func (registry *Register[Key, Value]) Register(key Key, newValue Value) Value {
	var currentValue Value
	if registry == nil {
		logger.Warn("Register.Register was called on nil register. This should not happen in normal flow of code", "key", key, "newValue", newValue)
	} else {
		regRecord := registry.getRegistrationRecord(key)
		logger.Log("Found registration record", regRecord)
		if regRecord != nil {
			logger.Log("Setting value on registration record", regRecord, "value", newValue)
			currentValue = regRecord.Set(newValue)
		} else {
			logger.Log("Creating registration recordInternal for key", key)
			regRecord := &registrationRecord[Key, Value]{
				key:   key,
				value: newValue,
				lock:  &sync.RWMutex{},
			}
			registry.setRegistrationRecord(key, regRecord)
		}
	}
	return currentValue
}

func (registry *Register[Key, Value]) isInitialized() bool {
	if registry != nil && registry.register != nil && registry.lock != nil {
		return true
	}
	return false
}

func (registry *Register[Key, Value]) getRegistrationRecord(key Key) *registrationRecord[Key, Value] {
	if registry == nil || !registry.isInitialized() {
		logger.Warn("Register.getRegistrationRecord was called on nil register or not initialized. This should not happen in normal flow of code", "key", key)
		return nil
	}
	var record *registrationRecord[Key, Value] = nil
	logger.Log("Locking read lock with deferred unlock on entry", registry)
	registry.lock.RLock()
	defer deferredRUnlock(registry, registry.lock)
	if recordInternal, ok := registry.register[key]; ok && recordInternal != nil {
		record = recordInternal
		logger.Log("Located existing record on entry", registry)
	}
	logger.Log("Returning record", record)
	return record
}

func (registry *Register[Key, Value]) setRegistrationRecord(key Key, record *registrationRecord[Key, Value]) *registrationRecord[Key, Value] {
	var exitingRecord *registrationRecord[Key, Value] = nil
	if registry == nil || !registry.isInitialized() {
		logger.Warn("Register.setRegistrationRecord was called on nil register or not initialized. This should not happen in normal flow of code", "key", key, "record", record)
	} else {
		logger.Log("Locking lock on entry", registry)
		registry.lock.Lock()
		defer deferredUnlock(registry, registry.lock)
		if internalCurrentValue, ok := registry.register[key]; ok {
			exitingRecord = internalCurrentValue
			logger.Log("Got current value from register ", registry, " for key ", key, " as ", internalCurrentValue)
		}
		registry.register[key] = record
		logger.Log("Set current value in register ", registry, " for key ", key, " as ", record)
	}
	return exitingRecord
}

func (registry *Register[Key, Value]) Unregister(key Key) Value {
	var currentEntry Value
	if registry == nil {
		logger.Warn("Register.Unregister was called on nil register. This should not happen in normal flow of code", "key", key)
	} else {
		logger.Log("Locking lock with deferred unlock on entry", registry)
		registry.lock.Lock()
		defer deferredUnlock(registry, registry.lock)
		if currentEntryInMap, ok := registry.register[key]; ok {
			logger.Log("Registration record located for key", key, "value", currentEntryInMap)
			delete(registry.register, key)
			logger.Log("Deleted key from map")
			if currentEntryInMap != nil {
				currentEntry = currentEntryInMap.Get()
				logger.Log("Retrieved current value as", currentEntry)
			}
		} else {
			logger.Log("No registration record was located for key", key)
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

func NewRegister[Key comparable, Value any]() *Register[Key, Value] {
	return &Register[Key, Value]{register: map[Key]*registrationRecord[Key, Value]{}, lock: &sync.RWMutex{}}
}
