package logger

import (
	"context"
	"sync"
)

type Marker interface {
	String() string
	Append(aspects ...string) Marker
	Add(aspects ...string) Marker
	GetAspects() []string
	AddMethod(callerMarker Marker, method string, values ...interface{}) Marker
	Push(marker Marker) Marker
	Pop() Marker
}

type CompareResult int

const (
	Equal         CompareResult = 0
	Less                        = -1
	Greater                     = 1
	NotApplicable               = 10
)

type LogSystem struct {
	defaultLoggerName string
	defaultLogger     Logger
	defaultLoggerLock *sync.Mutex
	loggers           Loggers
	logDrivers        LogDrivers
}

type Loggers struct {
	loggers map[string]*RootLoggerRegistration
	lock    *sync.Mutex
}

type RootLoggerRegistration struct {
	name       string
	rootLogger Logger
	lock       *sync.Mutex
}

type Level interface {
	String() string
	Compare(level Level) CompareResult
}

type LogConfig interface {
	GetLogDriver() LogDriver
	String() string
}

type LogConfigLevel interface {
	GetLevel() Level
}

type LogConfigForMarker interface {
	GetConfig(marker Marker) LogConfig
}

type BaseLogger interface {
	GetLogDriver() LogDriver
	Log(message string, v ...interface{}) Logger
	Named(name string) Logger
}

type LevelLogger interface {
	Trace(message string, v ...interface{}) Logger
	Debug(message string, v ...interface{}) Logger
	Info(message string, v ...interface{}) Logger
	Warn(message string, v ...interface{}) Logger
	Error(message string, v ...interface{}) Logger
	Fatal(message string, v ...interface{}) Logger
	LogL(level Level, message string, v ...interface{}) Logger
	IsLoggable(level Level) bool
	GetLevel() Level
	SetLevel(level Level) Logger
}

type MethodLogger interface {
	PushTo(parentContext context.Context, methodName string, v ...interface{}) context.Context
	Entering(parentContext context.Context, methodName string, v ...interface{}) (Logger, context.Context)
	EnteringM(parentContext context.Context, methodName string, v ...interface{}) Logger
	Exiting(v ...interface{})
}

type Logger interface {
	BaseLogger
	LevelLogger
	MethodLogger
}
