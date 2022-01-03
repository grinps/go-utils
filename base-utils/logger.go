package logger

import (
	"fmt"
	goLogger "log"
	"os"
	"strings"
)

const (
	LogUtilEnableTraceEnvironmentName  string = "TRACE_LOG_UTIL_ENABLE"
	LogUtilEnableTraceEnvironmentValue string = "ENABLE"
	LogUtilTraceFormat                 string = "TRACE_LOG_UTIL_FORMAT"
	LogUtilTraceFormatDefault          string = "Trace Log Util: %s: %s"
)

var traceLogUtil bool = true
var traceLogUtilFormat = LogUtilTraceFormatDefault

func init() {
	initialize()
}

func initialize() {
	traceLogUtilValue := os.Getenv(LogUtilEnableTraceEnvironmentName)
	if strings.ToUpper(traceLogUtilValue) == LogUtilEnableTraceEnvironmentValue {
		traceLogUtil = true
		goLogger.Printf("Trace log Util: Enabling logging. Change value of %s to a value other than case-insensitive %s to stop trace logging", LogUtilEnableTraceEnvironmentName, LogUtilEnableTraceEnvironmentValue)
	} else {
		traceLogUtil = false
		goLogger.Printf("Trace log Util: Logging not enabled. Set value of %s to case-insensitive %s instead of current value %s to start trace logging", LogUtilEnableTraceEnvironmentName, LogUtilEnableTraceEnvironmentValue, traceLogUtilValue)
	}
	traceLogUtilFormatTmp := os.Getenv(LogUtilTraceFormat)
	if traceLogUtilFormatTmp != "" && strings.Contains(traceLogUtilFormatTmp, "%s") {
		goLogger.Printf("Trace log Util: Changing the current logging format from %s to %s.", traceLogUtilFormat, traceLogUtilFormatTmp)
		traceLogUtilFormat = traceLogUtilFormatTmp
	} else {
		goLogger.Printf("Trace log Util: The environment variable %s with value %s does not contain any %s for formatting log. Using log format as %s.", "TRACE_LOG_UTIL_FORMAT", traceLogUtilFormatTmp, "%s", traceLogUtilFormat)
	}
	goLogger.SetFlags(goLogger.LstdFlags | goLogger.Lshortfile)
}

func Log(message string, values ...interface{}) {
	if traceLogUtil {
		keyValues := strings.TrimSuffix(fmt.Sprintln(values...), "\n")
		_ = goLogger.Output(2, fmt.Sprintf(traceLogUtilFormat, message, keyValues))
	}
}
