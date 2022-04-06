package logger

import (
	"reflect"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	logger, loggerError := DefaultE()
	if logger == nil {
		t.Error("Default should always return logger", "loggerError", loggerError)
	} else if _, ok := logger.(*SimpleLoggerWrapper); !ok {
		t.Error("Expected logger of type SimpleLoggerWrapper", "Type", reflect.TypeOf(logger))
	} else {
		logger.Trace("This is trace message with no parameters.")
		logger.Trace("This is trace message with one parameter", "param1")
		logger.Trace("This is trace message with one parameter & value", "param1=", true)
		logger.Debug("This is debug message with no parameters.")
		logger.Debug("This is debug message with one parameter", "param1")
		logger.Debug("This is debug message with one parameter & value", "param1=", "value1")
		logger.Info("This is info message with no parameters.")
		logger.Info("This is info message with one parameter", "param1")
		logger.Info("This is info message with one parameter & value", "param1=", 1.0)
		logger.Warn("This is warn message with no parameters.")
		logger.Warn("This is warn message with one parameter", "param1")
		logger.Warn("This is warn message with one parameter & value", "param1=", logger)
		logger.Error("This is error message with no parameters.")
		logger.Error("This is error message with one parameter", "param1")
		logger.Error("This is error message with one parameter & value", "param1=", 1)
	}

}
