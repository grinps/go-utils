package logger

import (
	"context"
	utils "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
)

type ContextKey string

const (
	MarkerContextKey                           ContextKey = "GoUtilsLoggerContext"
	SimpleLoggerWrapperContextMarkerErrorCodes            = "SimpleLoggerWrapperContextMarkerErrorCodes"
)

type LogPlugin interface {
	Log(marker Marker, level Level, message string, v ...interface{})
}

type logConfigType int

const (
	SimpleLoggerConfigNotInitialized logConfigType = 0
	SimpleLoggerConfigInherited      logConfigType = 1
	SimpleLoggerConfigSet            logConfigType = 2
)

//TODO add auto-updates from parent down for levels.

type SimpleLoggerWrapper struct {
	parent        *SimpleLoggerWrapper
	logConfig     LogConfig
	logConfigType logConfigType
	logPlugin     LogPlugin
	level         Level
	marker        Marker
}

func getApplicableLogConfig(marker Marker, logConfig LogConfig, logConfigIsInherited bool) (LogConfig, logConfigType) {
	var returnLogConfig = logConfig
	var returnLogConfigType = SimpleLoggerConfigSet
	if logConfigIsInherited {
		returnLogConfigType = SimpleLoggerConfigInherited
	}
	if logConfigWithMarker, ok := logConfig.(LogConfigForMarker); ok {
		logConfigForMarker := logConfigWithMarker.GetConfig(marker)
		if logConfigForMarker != nil {
			returnLogConfig = logConfigForMarker
			returnLogConfigType = SimpleLoggerConfigInherited
			utils.Log("Located applicable config using marker", "marker", marker)
		}
	}
	return returnLogConfig, returnLogConfigType
}

func getApplicableLogLevel(applicableLogConfig LogConfig, defaultValue Level) Level {
	var applicableLogLevel = defaultValue
	if applicableLogConfigWithLevel, ok := applicableLogConfig.(LogConfigLevel); ok {
		configLevel := applicableLogConfigWithLevel.GetLevel()
		if configLevel != nil {
			applicableLogLevel = configLevel
			utils.Log("Got applicable level from config", "level", applicableLogLevel)
		}
	}
	return applicableLogLevel
}

func NewSimpleLogWrapper(loggerName string, marker Marker, logConfig LogConfig, logPlugin LogPlugin) *SimpleLoggerWrapper {
	if marker == nil {
		marker = NewSimpleMarker(loggerName)
		utils.Log("Created new marker from loggerName", "loggerName", loggerName)
	}
	applicableLogConfig, applicableLogConfigType := getApplicableLogConfig(marker, logConfig, false)
	applicableLogLevel := getApplicableLogLevel(applicableLogConfig, Info)
	return &SimpleLoggerWrapper{
		parent:        nil,
		logConfigType: applicableLogConfigType,
		logConfig:     applicableLogConfig,
		logPlugin:     logPlugin,
		marker:        marker,
		level:         applicableLogLevel,
	}
}

func NewSimpleLogWrapperFromParent(parent *SimpleLoggerWrapper, marker Marker) *SimpleLoggerWrapper {
	if marker == nil {
		marker = parent.marker
	}
	applicableLogConfig, applicableLogConfigType := getApplicableLogConfig(marker, parent.logConfig, true)
	applicableLogLevel := getApplicableLogLevel(applicableLogConfig, parent.level)
	return &SimpleLoggerWrapper{
		parent:        parent,
		logConfigType: applicableLogConfigType,
		logConfig:     applicableLogConfig,
		logPlugin:     parent.logPlugin,
		marker:        marker,
		level:         applicableLogLevel,
	}
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) SetLevel(level Level) Logger {
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to set level since the logger is not initialized.", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		simpleLoggerWrapper.level = level //TODO: Need to support child logger levels auto-updates.
		utils.Log("Setting log level", "level", level)
		break
	}
	return simpleLoggerWrapper
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) GetLevel() Level {
	var returnValue Level = NoLevel
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to get level since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		returnValue = simpleLoggerWrapper.level
		break
	}
	return returnValue
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) GetLogDriver() LogDriver {
	var returnValue LogDriver = nil
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to get log driver since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		returnValue = simpleLoggerWrapper.logConfig.GetLogDriver()
		break
	}
	return returnValue
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) getLogPlugin() LogPlugin {
	return simpleLoggerWrapper.logPlugin
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Log(message string, v ...interface{}) Logger {
	simpleLoggerWrapper.LogL(Info, message, v...)
	return simpleLoggerWrapper
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Named(name string) Logger {
	var returnLogger Logger = simpleLoggerWrapper
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to create named logger since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		var newMarker = simpleLoggerWrapper.marker.Add(name)
		return NewSimpleLogWrapperFromParent(simpleLoggerWrapper, newMarker)
	}
	return returnLogger
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Trace(message string, v ...interface{}) Logger {
	simpleLoggerWrapper.LogL(Trace, message, v...)
	return simpleLoggerWrapper
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Debug(message string, v ...interface{}) Logger {
	simpleLoggerWrapper.LogL(Debug, message, v...)
	return simpleLoggerWrapper
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Info(message string, v ...interface{}) Logger {
	simpleLoggerWrapper.LogL(Info, message, v...)
	return simpleLoggerWrapper
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Warn(message string, v ...interface{}) Logger {
	simpleLoggerWrapper.LogL(Warn, message, v...)
	return simpleLoggerWrapper
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Error(message string, v ...interface{}) Logger {
	simpleLoggerWrapper.LogL(ERROR, message, v...)
	return simpleLoggerWrapper
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Fatal(message string, v ...interface{}) Logger {
	simpleLoggerWrapper.LogL(Fatal, message, v...)
	return simpleLoggerWrapper
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) LogL(level Level, message string, v ...interface{}) Logger {
	var returnValue Logger = simpleLoggerWrapper
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
	default:
		utils.Log("Failed to log message since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		if levelCompareResult := simpleLoggerWrapper.level.Compare(level); levelCompareResult == Less || levelCompareResult == Equal {
			simpleLoggerWrapper.logPlugin.Log(simpleLoggerWrapper.marker, level, message, v...)
		}
		break
	}
	return returnValue
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) IsLoggable(level Level) bool {
	var returnResult = false
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to check loggable since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		if result := simpleLoggerWrapper.level.Compare(level); result == Less || result == Equal {
			returnResult = true
		}
	}
	return returnResult
}

var GetMarkerErrorIncorrectObjectType = errext.NewUniqueErrorCodeOfType(1, SimpleLoggerWrapperContextMarkerErrorCodes)

func GetMarker(parentContext context.Context) (Marker, error) {
	var returnMarker Marker = nil
	var returnError error = nil
	if parentContext != nil {
		existingMarkerValue := parentContext.Value(MarkerContextKey)
		if existingMarkerValue != nil {
			if existingMarker, ok := existingMarkerValue.(Marker); !ok {
				returnError = GetMarkerErrorIncorrectObjectType.NewF("existingMarkerValue", existingMarkerValue)
			} else {
				returnMarker = existingMarker
			}
		}
	}
	return returnMarker, returnError
}

func PutMarker(parentContext context.Context, newMarker Marker) context.Context {
	var returnContext = parentContext
	if parentContext == nil {
		returnContext = context.WithValue(context.Background(), MarkerContextKey, newMarker)
	} else {
		returnContext = context.WithValue(parentContext, MarkerContextKey, newMarker)
	}
	return returnContext
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) pushToMarker(parentContext context.Context, methodName string, v ...interface{}) Marker {
	var returnMarker Marker = DefaultSimpleMarker
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to push marker since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		existingMarker, exitingMarkerExtractError := GetMarker(parentContext)
		if exitingMarkerExtractError != nil {
			utils.Log("Failed to extract marker", "parentContext", parentContext, "exitingMarkerExtractError", exitingMarkerExtractError)
			existingMarker = nil
		}
		returnMarker = simpleLoggerWrapper.marker.AddMethod(existingMarker, methodName, v...)
	}
	return returnMarker
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) PushTo(parentContext context.Context, methodName string, v ...interface{}) context.Context {
	var returnContext = parentContext
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to push to context since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		newMarker := simpleLoggerWrapper.pushToMarker(parentContext, methodName, v...)
		if newMarker != DefaultSimpleMarker {
			returnContext = PutMarker(parentContext, newMarker)
		} else {
			utils.Log("Skipping putting default simple marker to parent context.", "simpleLoggerWrapper", simpleLoggerWrapper)
		}
	}
	return returnContext
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Entering(parentContext context.Context, methodName string, v ...interface{}) (Logger, context.Context) {
	var returnContext = parentContext
	var returnLogger Logger = simpleLoggerWrapper
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to print entering method since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		newMarker := simpleLoggerWrapper.pushToMarker(parentContext, methodName, v...)
		returnContext = PutMarker(parentContext, newMarker)
		returnLogger = NewSimpleLogWrapperFromParent(simpleLoggerWrapper, newMarker)
		returnLogger.Trace("Entering method")
	}
	return returnLogger, returnContext
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) EnteringM(parentContext context.Context, methodName string, v ...interface{}) Logger {
	var returnLogger = simpleLoggerWrapper
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to print entering method and return logger since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		newMarker := simpleLoggerWrapper.pushToMarker(parentContext, methodName, v...)
		returnLogger = NewSimpleLogWrapperFromParent(simpleLoggerWrapper, newMarker)
		returnLogger.Trace("Entering method")
	}
	return returnLogger
}

func (simpleLoggerWrapper *SimpleLoggerWrapper) Exiting(values ...interface{}) {
	switch simpleLoggerWrapper.logConfigType {
	case SimpleLoggerConfigNotInitialized:
		utils.Log("Failed to print exiting method since logger not initialized", "simpleLoggerWrapper", simpleLoggerWrapper)
		break
	case SimpleLoggerConfigSet, SimpleLoggerConfigInherited:
		simpleLoggerWrapper.Trace("Leaving method", values...)
	}
}
