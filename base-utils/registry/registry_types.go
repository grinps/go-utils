package registry

import (
	"sync"
)

type registrationRecord[Key comparable, Value any] struct {
	key   Key
	value Value
	lock  *sync.RWMutex
}

type Register[Key comparable, Value any] struct {
	register map[interface{}]*registrationRecord[Key, Value]
	lock     *sync.RWMutex
}

type CustomKey interface {
	Unique() string
}
