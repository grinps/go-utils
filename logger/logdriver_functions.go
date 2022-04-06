package logger

import (
	utils "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
)

const (
	LogDriverErrorCodeType = "LogDriver"
)

var InitializeErrorCode = errext.NewErrorCodeOfType(1, LogDriverErrorCodeType)

func InitializeLogger(loggerName string, opts LogConfig) (Logger, error) {
	return DefaultLogSystem.InitializeLoggerE(loggerName, opts)
}

func (logSystem *LogSystem) InitializeLogger(loggerName string, opts LogConfig) Logger {
	return logSystem.InitializeLoggerP(loggerName, opts, true)
}

func (logSystem *LogSystem) InitializeLoggerP(loggerName string, opts LogConfig, panicOnError bool) Logger {
	loggerInstance, err := logSystem.InitializeLoggerE(loggerName, opts)
	if err != nil && panicOnError {
		panic(InitializeErrorCode.NewWithErrorF(err, "Failed to initialize logger", "loggerName", loggerName, "opts", opts, "panicOnError", panicOnError))
	}
	return loggerInstance
}

var MissingLogConfig = errext.NewErrorCodeOfType(2, LogDriverErrorCodeType)
var MissingLogDriver = errext.NewErrorCodeOfType(3, LogDriverErrorCodeType)
var LogDriverInitializationFailed = errext.NewErrorCodeOfType(4, LogDriverErrorCodeType)
var LogAttachFailed = errext.NewErrorCodeOfType(5, LogDriverErrorCodeType)

func (logSystem *LogSystem) InitializeLoggerE(loggerName string, opts LogConfig) (Logger, error) {
	utils.Log("Entering InitializeLoggerE", "loggerName", loggerName, "opts", opts)
	var loggerInstance Logger = nil
	var loggerError error = nil
	if opts == nil {
		loggerError = MissingLogConfig.NewF("No log config was passed for initialization.", "loggerName", loggerName)
	} else {
		logDriver := opts.GetLogDriver()
		utils.Log("Log Driver retrieved", "logDriver", logDriver)
		if logDriver == nil {
			loggerError = MissingLogDriver.NewF("Log config did not return any log driver", "opts", opts)
		} else {
			utils.Log("Trying to initialize logger using log driver")
			logger, loggerInitError := logDriver.Initialize(loggerName, opts)
			if loggerInitError == nil {
				previousAttachedRootLogger, attachError := logSystem.getRootLoggerRegistration(loggerName).Attach(logger)
				utils.Log("Replaced existing root logger with new logger", "previousAttachedRootLogger", previousAttachedRootLogger)
				if attachError != nil {
					loggerError = LogAttachFailed.NewWithErrorF(attachError, "Failed to attach the initialized logger", "loggerName", loggerName, "logger", logger, "logDriver", logDriver)
				}
				loggerInstance = logger
			} else {
				utils.Log("Failed to initialize log driver", "error", loggerInitError)
				loggerError = LogDriverInitializationFailed.NewWithErrorF(loggerInitError, "Failed to initialize new logger", "loggerName", loggerName, "opts", opts)
			}
		}
	}
	utils.Log("Exiting InitializeLoggerE", "loggerInstance", loggerInstance, "loggerError", loggerError)
	return loggerInstance, loggerError
}
