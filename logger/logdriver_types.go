package logger

import "sync"

type LogCapability struct {
	name         string
	capabilityId int
}

type LogDriver interface {
	GetName() string
	Initialize(loggerName string, config LogConfig) (Logger, error)
}

type LogDrivers struct {
	logDrivers map[string]*LogDriverRegistration
	lock       *sync.Mutex
}

type LogDriverRegistration struct {
	name      string
	logDriver LogDriver
	lock      *sync.Mutex
}
