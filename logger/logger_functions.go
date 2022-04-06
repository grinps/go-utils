package logger

import (
	"context"
	utils "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
	"sync"
)

const (
	contextReference string = "LOGGER"
	DefaultLogger    string = "DefaultLogger"
	ErrorCodeType    string = "LoggerError"
)

var defaultLogConfig *GoLogConfig = &GoLogConfig{
	Prefix:     "",
	OutputFile: "",
	Flags: GoLogFlags{
		Date:                true,
		Time:                true,
		TimeInMicrosecond:   false,
		TimeInUTC:           false,
		LongFile:            false,
		ShortFile:           false,
		PrefixAtStartOfLine: false,
	},
	populated:     true,
	logConfigName: GoLogConfigName + "(DefaultLog)",
}

var DefaultLogSystem = LogSystem{
	defaultLoggerName: DefaultLogger,
	defaultLogger:     nil,
	defaultLoggerLock: &sync.Mutex{},
	loggers: Loggers{
		loggers: map[string]*RootLoggerRegistration{},
		lock:    &sync.Mutex{},
	},
	logDrivers: LogDrivers{
		logDrivers: map[string]*LogDriverRegistration{},
		lock:       &sync.Mutex{},
	},
}

var DefaultLoggerErrorCode = errext.NewErrorCodeOfType(1, ErrorCodeType)

func DefaultE() (Logger, error) {
	return DefaultLogSystem.DefaultE()
}

func Default() Logger {
	return DefaultLogSystem.Default()
}

func (logSystem *LogSystem) Default() Logger {
	return logSystem.DefaultP(true)
}

func (logSystem *LogSystem) DefaultP(panicOnError bool) Logger {
	logger, err := logSystem.DefaultE()
	if err != nil && panicOnError {
		panic(DefaultLoggerErrorCode.NewWithError("Failed to create default logger", err))
	}
	return logger
}

func (logSystem *LogSystem) DefaultE() (Logger, error) {
	utils.Log("Entering DefaultE")
	var defaultLogger Logger = nil
	var loggerError error = nil
	if logSystem.defaultLogger == nil {
		logSystem.defaultLoggerLock.Lock()
		defer logSystem.defaultLoggerLock.Unlock()
		if logSystem.defaultLogger == nil {
			if generatedDefaultLogger, initErr := logSystem.InitializeLoggerE(DefaultLogger, defaultLogConfig); initErr != nil {
				loggerError = DefaultLoggerErrorCode.NewWithError("Initialization of default logger failed.", initErr)
			} else if generatedDefaultLogger == nil {
				loggerError = DefaultLoggerErrorCode.New("Default logger could not be initialized.")
			} else {
				logSystem.defaultLogger = generatedDefaultLogger
				defaultLogger = generatedDefaultLogger
			}
		} else {
			defaultLogger = logSystem.defaultLogger
		}
	} else {
		defaultLogger = logSystem.defaultLogger
	}
	utils.Log("Leaving DefaultE", "defaultLogger", defaultLogger, "loggerError", loggerError)
	return defaultLogger, loggerError
}

var ChangeLoggerErrorCode = errext.NewErrorCodeOfType(2, ErrorCodeType)

func ChangeDefault(newLogger string) error {
	return DefaultLogSystem.ChangeDefaultE(newLogger)
}

func (logSystem *LogSystem) ChangeDefault(newLogger string) *LogSystem {
	return logSystem.ChangeDefaultP(newLogger, true)
}

func (logSystem *LogSystem) ChangeDefaultP(newLogger string, panicOnError bool) *LogSystem {
	err := logSystem.ChangeDefaultE(newLogger)
	if err != nil && panicOnError {
		panic(ChangeLoggerErrorCode.NewWithErrorF(err, "Failed to change default logger", "newLogger", newLogger))
	}
	return logSystem
}

func (logSystem *LogSystem) ChangeDefaultE(newLogger string) error {
	utils.Log("Entering ChangeDefaultE", "newLogger", newLogger)
	if newLoggerObject, err := logSystem.GetRootLoggerE(newLogger); err == nil {
		logSystem.defaultLoggerLock.Lock()
		defer logSystem.defaultLoggerLock.Unlock()
		utils.Log("Switching to new logger", "defaultLoggerName", logSystem.defaultLoggerName, "newLogger", newLogger)
		logSystem.defaultLoggerName = newLogger
		logSystem.defaultLogger = newLoggerObject
		utils.Log("Leaving ChangeDefaultE", "err", err)
		return err
	} else {
		utils.Log("Leaving ChangeDefaultE", "err", err)
		return err
	}
}

var GetRootLoggerErrorCode = errext.NewErrorCodeOfType(3, ErrorCodeType)

func GetRootLogger(loggerName string) (Logger, error) {
	return DefaultLogSystem.GetRootLoggerE(loggerName)
}

func (logSystem *LogSystem) GetRootLogger(loggerName string) Logger {
	return logSystem.GetRootLoggerP(loggerName, true)
}

func (logSystem *LogSystem) GetRootLoggerP(loggerName string, panicOnError bool) Logger {
	loggerObject, err := logSystem.GetRootLoggerE(loggerName)
	if err != nil && panicOnError {
		panic(GetRootLoggerErrorCode.NewWithErrorF(err, "Failed to get root logger", "loggerName", loggerName))
	}
	return loggerObject
}

func (logSystem *LogSystem) GetRootLoggerE(loggerName string) (Logger, error) {
	utils.Log("Entering GetRootLoggerE", "loggerName", loggerName)
	var loggerError error = nil
	var returnValue Logger = nil
	if loggerName == "" {
		utils.Log("Getting log tracker for default logger", "defaultLogger", logSystem.defaultLoggerName)
		loggerName = logSystem.defaultLoggerName
	}
	utils.Log("Checking if log tracker is already setup", "loggerName", loggerName)
	returnValue = logSystem.getRootLoggerRegistration(loggerName).GetRootLogger()
	if returnValue == nil {
		loggerError = GetRootLoggerErrorCode.NewF("No root logger corresponding to given name has been initialised.", "loggerName", loggerName)
	}
	utils.Log("Exiting GetRootLoggerE", "returnValue", returnValue, "loggerError", loggerError)
	return returnValue, loggerError
}

func (logSystem *LogSystem) getRootLoggerRegistration(loggerName string) *RootLoggerRegistration {
	var logTracker *RootLoggerRegistration = nil
	logSystem.loggers.lock.Lock()
	defer logSystem.loggers.lock.Unlock()
	if logTrackerInternal, ok := logSystem.loggers.loggers[loggerName]; !ok || logTrackerInternal == nil {
		utils.Log("Creating log tracker")
		logTrackerInternal = &RootLoggerRegistration{
			name:       loggerName,
			rootLogger: nil,
			lock:       &sync.Mutex{},
		}
		utils.Log("Assigning log tracker")
		logSystem.loggers.loggers[loggerName] = logTrackerInternal
		logTracker = logTrackerInternal
	} else {
		utils.Log("Located existing log tracker")
		logTracker = logTrackerInternal
	}
	return logTracker
}

func (rootLogger *RootLoggerRegistration) Attach(newRootLogger Logger) (Logger, error) {
	utils.Log("Entering Attach", "newRootLogger", newRootLogger)
	var loggerError error = nil
	var existingRootLogger Logger = nil
	rootLogger.lock.Lock()
	defer rootLogger.lock.Unlock()
	existingRootLogger = rootLogger.rootLogger
	rootLogger.rootLogger = newRootLogger
	utils.Log("Leaving Attach", "existingRootLogger", existingRootLogger, "loggerError", loggerError)
	return existingRootLogger, loggerError
}

func (rootLogger *RootLoggerRegistration) GetRootLogger() Logger {
	rootLogger.lock.Lock()
	defer rootLogger.lock.Unlock()
	return rootLogger.rootLogger
}

func WithLogger(parent context.Context, logger Logger) context.Context {
	return context.WithValue(parent, contextReference, logger)
}

func FromContext(parent context.Context) (Logger, bool) {
	logger, ok := parent.Value(contextReference).(Logger)
	return logger, ok
}
