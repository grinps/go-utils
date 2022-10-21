package logger

import (
	"github.com/grinps/go-utils/errext"
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

func HandlePanic(t *testing.T, errCode errext.ErrorCode) {
	if r := recover(); r != nil {
		if asErr, isErr := r.(error); !isErr {
			t.Error("Expected recovered value of type err. Received ", r)
		} else if _, isInstanceOf := errCode.AsError(asErr); !isInstanceOf {
			t.Error("Expected panic error of type InitializeErrorCode. Received ", r)
		}
	} else {
		t.Error("Expected the call to fail with a panic. Either no panic or panic didn't result in an error")
	}
}

func TestLogSystem_InitializeLogger(t *testing.T) {
	t.Run("NilLogSystemWithEmptyLoggerNameAndNilOpts", func(t *testing.T) {
		var logSystem LogSystem
		defer HandlePanic(t, InitializeErrorCode)
		logger := logSystem.InitializeLogger("", nil)
		if logger != nil {
			t.Error("Expected call to fail due to missing opts", "logger", logger)
		}
	})
	t.Run("NilLogSystemWithLoggerNameAndNilOpts", func(t *testing.T) {
		defer HandlePanic(t, InitializeErrorCode)
		var logSystem LogSystem
		logger := logSystem.InitializeLogger("SomeLogger", nil)
		if logger != nil {
			t.Error("Expected call to fail due to missing opts", "logger", logger)
		}
	})
}
