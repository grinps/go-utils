package logger

import (
	golog "log"
	"os"
	"path"
	"strings"
	"testing"
)

func TestGoLogConfig_GetLogDriver(t *testing.T) {
	t.Run("DefaultGoLogConfig", func(t *testing.T) {
		defaultGoLogConfig := GoLogConfig{}
		logDriver := defaultGoLogConfig.GetLogDriver()
		if logDriver == nil {
			t.Error("GoLogConfig should always return log driver")
		} else if _, isInstanceOf := logDriver.(*goLogDriver); !isInstanceOf {
			t.Error("GoLogConfig should always return an instance of pointer to goLogDriver")
		}
	})
	t.Run("NullGoLogConfig", func(t *testing.T) {
		var nilGoLogConfig *GoLogConfig = nil
		if logDriver := nilGoLogConfig.GetLogDriver(); logDriver == nil {
			t.Error("GoLogConfig should always return log driver")
		} else if _, isInstanceOf := logDriver.(*goLogDriver); !isInstanceOf {
			t.Error("GoLogConfig should always return an instance of pointer to goLogDriver")
		}
	})
}

func TestGoLogConfig_String(t *testing.T) {
	t.Run("DefaultGoLogConfig", func(t *testing.T) {
		defaultGoLogConfig := GoLogConfig{}
		logDriverName := defaultGoLogConfig.String()
		if logDriverName != GoLogConfigName {
			t.Error("GoLogConfig should always return " + GoLogConfigName)
		}
	})
	t.Run("NullGoLogConfig", func(t *testing.T) {
		var nilGoLogConfig *GoLogConfig = nil
		if logDriverName := nilGoLogConfig.String(); logDriverName != GoLogConfigName {
			t.Error("GoLogConfig should always return " + GoLogConfigName)
		}
	})
	t.Run("Initialized GoLogConfig", func(t *testing.T) {
		logConfigName := "initGoLogConfig"
		var initGoLogConfig *GoLogConfig = &GoLogConfig{
			logConfigName: logConfigName,
		}
		if logDriverName := initGoLogConfig.String(); logDriverName != logConfigName {
			t.Error("GoLogConfig should return " + logConfigName)
		}
	})
}

func TestGoLogConfig_GetLevel(t *testing.T) {
	t.Run("DefaultGoLogConfig", func(t *testing.T) {
		defaultGoLogConfig := GoLogConfig{}
		logLevel := defaultGoLogConfig.GetLevel()
		if logLevel != Info {
			t.Error("GoLogConfig should always return " + Info.String())
		}
	})
	t.Run("NullGoLogConfig", func(t *testing.T) {
		var nilGoLogConfig *GoLogConfig = nil
		if logLevel := nilGoLogConfig.GetLevel(); logLevel != Info {
			t.Error("GoLogConfig should always return " + Info.String())
		}
	})
}

func TestGoLogConfig_getWriter(t *testing.T) {
	t.Run("DefaultGoLogConfig", func(t *testing.T) {
		defaultGoLogConfig := GoLogConfig{}
		logWriter := defaultGoLogConfig.getWriter()
		if logWriter != os.Stderr {
			t.Error("Default GoLogConfig should return os.Stderr ", "logWriter", logWriter)
		}
	})
	t.Run("NilGoLogConfig", func(t *testing.T) {
		var defaultGoLogConfig *GoLogConfig = nil
		logWriter := defaultGoLogConfig.getWriter()
		if logWriter != nil {
			t.Error("Nil GoLogConfig should return nil", "logWriter", logWriter)
		}
	})
	t.Run("ValidGoLogFileName", func(t *testing.T) {
		configWdFileName := "logoutput.deleteMe"
		workDirLocation, _ := os.Getwd()
		workDirConfigLocation := path.Join(workDirLocation, configWdFileName)
		defer func() {
			t.Log("Trying to delete local file", "workDirConfigLocation", workDirConfigLocation)
			err := os.Remove(workDirConfigLocation)
			if err != nil {
				t.Log("Failed to delete local file", "workDirConfigLocation", workDirConfigLocation, "err", err)
			}
		}()
		var defaultGoLogConfig *GoLogConfig = &GoLogConfig{
			OutputFile: workDirConfigLocation,
		}
		logWriter := defaultGoLogConfig.getWriter()
		if logWriter == nil {
			t.Error("The writer should have been returned.", "workDirConfigLocation", workDirConfigLocation)
		}
	})
	t.Run("InvalidGoLogFileName", func(t *testing.T) {
		workDirLocation, _ := os.Getwd()
		workDirConfigLocation := workDirLocation
		var defaultGoLogConfig *GoLogConfig = &GoLogConfig{
			OutputFile: workDirConfigLocation,
		}
		logWriter := defaultGoLogConfig.getWriter()
		if logWriter != os.Stderr {
			t.Error("The os.Stderr should have been returned since the given location is a directory.", "workDirConfigLocation", workDirConfigLocation, "logWriter", logWriter)
		}
	})
}

func TestGoLogConfig_getFlag(t *testing.T) {
	t.Run("DefaultGoLogConfig", func(t *testing.T) {
		defaultGoLogConfig := GoLogConfig{}
		flags := defaultGoLogConfig.getFlag()
		if flags != golog.LstdFlags {
			t.Error("Default GoLogConfig should return golog.LstdFlags", "flags", flags, "golog.LstdFlags", golog.LstdFlags)
		}
	})
	t.Run("NilGoLogConfig", func(t *testing.T) {
		var defaultGoLogConfig *GoLogConfig = nil
		flags := defaultGoLogConfig.getFlag()
		if flags != golog.LstdFlags {
			t.Error("Nil GoLogConfig should return golog.LstdFlags", "flags", flags, "golog.LstdFlags", golog.LstdFlags)
		}
	})
	t.Run("InitializedGoFlags", func(t *testing.T) {
		var defaultGoLogConfig *GoLogConfig = &GoLogConfig{}
		defaultGoLogConfig.Flags = GoLogFlags{
			Date:                false,
			Time:                false,
			TimeInMicrosecond:   false,
			TimeInUTC:           true,
			LongFile:            false,
			ShortFile:           true,
			PrefixAtStartOfLine: true,
		}
		defaultGoLogConfig.populated = true
		defaultGoLogConfig.Prefix = "PREFIX--"
		outputValue := &strings.Builder{}
		defaultGoLogConfig.outputFileReference = outputValue
		logger, err := goLogDriverInstance.Initialize("TestLogger", defaultGoLogConfig)
		if err != nil || logger == nil {
			t.Error("The setting should be able to create test logger", "err", err)
		}
		logger.Warn("This is warning")
		expectedOutputString := "goLogDriver_test.go:153: PREFIX--[Warn]:[TestLogger]:This is warning"
		actualOutputString := outputValue.String()
		if strings.Compare(actualOutputString, expectedOutputString) == 0 {
			t.Error("Output log does not contain expected output.", "expectedOutputString", expectedOutputString, "actualOutputString", actualOutputString)
		}
	})
	t.Run("InitializedOtherGoFlags", func(t *testing.T) {
		var defaultGoLogConfig *GoLogConfig = &GoLogConfig{}
		defaultGoLogConfig.Flags = GoLogFlags{
			Date:                true,
			Time:                true,
			TimeInMicrosecond:   true,
			TimeInUTC:           true,
			LongFile:            true,
			ShortFile:           false,
			PrefixAtStartOfLine: false,
		}
		defaultGoLogConfig.populated = true
		outputValue := &strings.Builder{}
		defaultGoLogConfig.outputFileReference = outputValue
		logger, err := goLogDriverInstance.Initialize("Test2Logger", defaultGoLogConfig)
		if err != nil || logger == nil {
			t.Error("The setting should be able to create test logger", "err", err)
		}
		logger.Warn("This is warning 2")
		expectedOutputString := "goLogDriver_test.go:178: [Warn]:[Test2Logger]:This is warning 2"
		actualOutputString := outputValue.String()
		if !strings.Contains(actualOutputString, expectedOutputString) {
			t.Error("Output log does not contain expected output.", "expectedOutputString", expectedOutputString, "actualOutputString", actualOutputString)
		}
	})
}

func TestGoLogDriver_Initialize(t *testing.T) {
	t.Run("InitializeWithNilConfig", func(t *testing.T) {
		logger, err := goLogDriverInstance.Initialize("genericLogger", nil)
		if err == nil {
			t.Error("The Nil config should have resulted in error")
		} else if logger != nil {
			t.Error("The Nil config should not have returned a logger", "logger", logger)
		} else if _, isInstanceOf := GoLogDriverInitializationMissingConfigErrorCode.AsError(err); !isInstanceOf {
			t.Error("The error should be instance of GoLogDriverInitializationMissingConfigErrorCode", "err", err)
		}
	})
	t.Run("InitializeWithIncorrectConfig", func(t *testing.T) {
		logger, err := goLogDriverInstance.Initialize("genericLogger", &SimpleLogConfig{})
		if err == nil {
			t.Error("The incorrect config should have resulted in error")
		} else if logger != nil {
			t.Error("The incorrect config should not have returned a logger", "logger", logger)
		} else if _, isInstanceOf := GoLogDriverInitializationInvalidLogConfigTypeErrorCode.AsError(err); !isInstanceOf {
			t.Error("The error should be instance of GoLogDriverInitializationInvalidLogConfigTypeErrorCode", "err", err)
		}
	})
	t.Run("InitializeWithUnpopulatedConfig", func(t *testing.T) {
		logger, err := goLogDriverInstance.Initialize("genericLogger", &GoLogConfig{})
		if err == nil {
			t.Error("The unpopulated config should have resulted in error")
		} else if logger != nil {
			t.Error("The unpopulated config should not have returned a logger", "logger", logger)
		} else if _, isInstanceOf := GoLogDriverInitializationNotPopulatedErrorCode.AsError(err); !isInstanceOf {
			t.Error("The error should be instance of GoLogDriverInitializationNotPopulatedErrorCode", "err", err)
		}
	})
	t.Run("GetName", func(t *testing.T) {
		if logDriverName := goLogDriverInstance.GetName(); logDriverName != GoLogDriverName {
			t.Error("The name of gologdriver does not match expected value", "actual value", logDriverName, "expected value", GoLogDriverName)
		}
	})
}
