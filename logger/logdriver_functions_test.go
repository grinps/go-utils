package logger

import (
	"strings"
	"testing"
)

func getSimpleGoLogConfig() (*GoLogConfig, *strings.Builder) {
	logCollector := &strings.Builder{}
	loggerOpts := &GoLogConfig{
		Prefix:              "PREFIX--",
		outputFileReference: logCollector,
		Flags:               GoLogFlags{},
		populated:           true,
		logConfigName:       "BasicLogConfig",
	}
	return loggerOpts, logCollector
}

func TestInitializeLogger(t *testing.T) {
	t.Run("NilOptions", func(t *testing.T) {
		logger, err := InitializeLogger("nilLogger", nil)
		if err == nil {
			t.Error("Expected call to fail due to missing opts")
		} else if logger != nil {
			t.Error("Expected call to fail due to missing opts", "logger", logger)
		} else if _, isInstanceOf := MissingLogConfig.AsError(err); !isInstanceOf {
			t.Error("Expected error of type MissingLogConfig", "err", err)
		}
	})
	loggerOpts, logCollector := getSimpleGoLogConfig()
	t.Run("ValidLoggerConfiguration", func(t *testing.T) {
		logger, err := InitializeLogger("", loggerOpts)
		if err != nil {
			t.Error("Expected call to succeed due to valid details.", "err", err)
		} else if logger == nil {
			t.Error("Expected call to succeed due to valid details")
		} else {
			logger.Warn("This is warn message")
			logOutput := logCollector.String()
			if !strings.Contains(logOutput, "This is warn message") {
				t.Error("Expected logging with following details", "expected value", "This is warn message", "actual value", logOutput)
			}
		}
	})
}

func TestLogSystem_InitializeLogger(t *testing.T) {

}