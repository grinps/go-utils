package logger

import (
	"fmt"
	goLogger "log"
	"os"
	"strings"
)

const (
	LogUtilEnableTraceEnvironmentName string = "TRACE_LOG_UTIL_ENABLE"
	LogUtilTraceFormat                string = "TRACE_LOG_UTIL_FORMAT"
	LogUtilTraceFormatDefault         string = "Trace Log Util: %s: %s"
	LogUtilDisableWarnEnvironmentName string = "TRACE_WARN_UTIL_DISABLE"
	LogUtilWarnFormat                 string = "TRACE_WARN_UTIL_FORMAT"
	LogUtilWarnFormatDefault          string = "WARN: %s: %s"
)

var LogUtilEnableTraceEnvironmentValue []string = []string{"1", "TRUE", "ENABLE"}

var traceLogUtil bool = false
var traceLogUtilFormat = LogUtilTraceFormatDefault
var warnLogUtil bool = true
var warnLogUtilFormat = LogUtilWarnFormatDefault

func init() {
	initialize()
}

func initialize() {
	traceLogUtilValue := os.Getenv(LogUtilEnableTraceEnvironmentName)
	if contains(strings.ToUpper(traceLogUtilValue), LogUtilEnableTraceEnvironmentValue) {
		traceLogUtil = true
		goLogger.Printf("Trace log Util: Enabling logging. Change value of %s to a value other than one of the %s to stop trace logging", LogUtilEnableTraceEnvironmentName, LogUtilEnableTraceEnvironmentValue)
	} else {
		traceLogUtil = false
		goLogger.Printf("Trace log Util: Logging not enabled. Set value of %s to one of the %s values instead of current value %s to start trace logging", LogUtilEnableTraceEnvironmentName, LogUtilEnableTraceEnvironmentValue, traceLogUtilValue)
	}
	traceLogUtilFormatTmp := os.Getenv(LogUtilTraceFormat)
	if traceLogUtilFormatTmp != "" && strings.Contains(traceLogUtilFormatTmp, "%s") {
		goLogger.Printf("Trace log Util: Changing the current logging format from %s to %s.", traceLogUtilFormat, traceLogUtilFormatTmp)
		traceLogUtilFormat = traceLogUtilFormatTmp
	} else {
		goLogger.Printf("Trace log Util: The environment variable %s with value %s does not contain any %s for formatting log. Using log format as %s.", LogUtilTraceFormat, traceLogUtilFormatTmp, "%s", traceLogUtilFormat)
	}
	warnUtilValue := os.Getenv(LogUtilDisableWarnEnvironmentName)
	if contains(strings.ToUpper(warnUtilValue), LogUtilEnableTraceEnvironmentValue) {
		warnLogUtil = false
		goLogger.Printf("Warn log Util: Disabling warn messages. Change value of %s to a value other than one of the %s values to start warning logging", LogUtilDisableWarnEnvironmentName, LogUtilEnableTraceEnvironmentValue)
	} else {
		warnLogUtil = true
		goLogger.Printf("Warn log Util: Warn messages enabled. Set value of environment variable %s to any of the values %s to stop logging of warn message.", LogUtilDisableWarnEnvironmentName, LogUtilEnableTraceEnvironmentValue)
	}
	warnLogUtilFormatTmp := os.Getenv(LogUtilWarnFormat)
	if warnLogUtilFormatTmp != "" && strings.Contains(warnLogUtilFormatTmp, "%s") {
		goLogger.Printf("Warn log Util: Changing the current logging format from %s to %s.", warnLogUtilFormat, warnLogUtilFormatTmp)
		warnLogUtilFormat = warnLogUtilFormatTmp
	} else {
		goLogger.Printf("Warn log Util: The environment variable %s with value %s does not contain any %s for formatting log. Using log format as %s.", LogUtilWarnFormat, warnLogUtilFormatTmp, "%s", warnLogUtilFormat)
	}
	goLogger.SetFlags(goLogger.LstdFlags | goLogger.Lshortfile)
}

func contains(value string, list []string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func Log(message string, values ...interface{}) {
	if traceLogUtil {
		keyValues := strings.TrimSuffix(fmt.Sprintln(values...), "\n")
		_ = goLogger.Output(2, fmt.Sprintf(traceLogUtilFormat, message, keyValues))
	}
}

func Warn(message string, values ...interface{}) {
	if warnLogUtil {
		keyValues := strings.TrimSuffix(fmt.Sprintln(values...), "\n")
		_ = goLogger.Output(2, fmt.Sprintf(warnLogUtilFormat, message, keyValues))
	}
}
