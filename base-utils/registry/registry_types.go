package registry

import (
	"sync"
)

type registrationRecord[Key comparable] struct {
	key   Key
	value interface{}
	lock  *sync.RWMutex
}

type Register[Key comparable] struct {
	register map[Key]*registrationRecord[Key]
	lock     *sync.RWMutex
}
