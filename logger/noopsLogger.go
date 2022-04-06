package logger

import "context"

type NoOpsLogger struct {
}

func (logger *NoOpsLogger) GetLevel() Level {
	return NoLevel
}

func (logger *NoOpsLogger) SetLevel(level Level) Logger {
	return logger
}

func (logger *NoOpsLogger) PushTo(parentContext context.Context, methodName string, v ...interface{}) context.Context {
	return parentContext
}

func (logger *NoOpsLogger) Entering(parentContext context.Context, methodName string, v ...interface{}) (Logger, context.Context) {
	return logger, parentContext
}

func (logger *NoOpsLogger) EnteringM(parentContext context.Context, methodName string, v ...interface{}) Logger {
	return logger
}

func (logger *NoOpsLogger) Exiting(v ...interface{}) {
}

func (logger *NoOpsLogger) GetLogDriver() LogDriver {
	return defaultNoOpsLogDriver
}

func (logger *NoOpsLogger) Log(message string, v ...interface{}) Logger {
	return logger
}

func (logger *NoOpsLogger) Named(name string) Logger {
	return logger
}

func (logger *NoOpsLogger) Trace(message string, v ...interface{}) Logger {
	return logger
}

func (logger *NoOpsLogger) Debug(message string, v ...interface{}) Logger {
	return logger
}

func (logger *NoOpsLogger) Info(message string, v ...interface{}) Logger {
	return logger
}

func (logger *NoOpsLogger) Warn(message string, v ...interface{}) Logger {
	return logger
}

func (logger *NoOpsLogger) Error(message string, v ...interface{}) Logger {
	return logger
}

func (logger *NoOpsLogger) Fatal(message string, v ...interface{}) Logger {
	return logger
}

func (logger *NoOpsLogger) LogL(level Level, message string, v ...interface{}) Logger {
	return logger
}

func (logger *NoOpsLogger) IsLoggable(level Level) bool {
	return true
}

type NoOpsLogDriver struct{}

var defaultNoOpsLogDriver = &NoOpsLogDriver{}
var defaultNoOpsLogger = &NoOpsLogger{}

func (logDriver *NoOpsLogDriver) Initialize(loggerName string, config LogConfig) (Logger, error) {
	return defaultNoOpsLogger, nil
}

func (logDriver *NoOpsLogDriver) GetName() string {
	return "NoOpsLogDriver"
}
