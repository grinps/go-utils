package logger

import (
	"io"
	"log"
	"strings"
	"testing"
)

func resetLogWriter(currentWriter io.Writer) {
	log.SetOutput(currentWriter)
}

func TestLog_NoEnvSet(t *testing.T) {
	defer resetLogWriter(log.Default().Writer())
	var logCollector = &strings.Builder{}
	log.Default().SetOutput(logCollector)
	t.Run("Empty message", func(t *testing.T) {
		testCase(logCollector, t, "Empty message was logged incorrectly.", "", "")
	})
	t.Run("Simple message", func(t *testing.T) {
		testCase(logCollector, t, "Simple message was logged incorrectly.", "", "SimpleMessage1")
	})
	t.Run("Message with nil value", func(t *testing.T) {
		testCase(logCollector, t, "Message with nil value was logged incorrectly.", "", "MessageWithNil1", nil)
	})
	t.Run("Message with value", func(t *testing.T) {
		testCase(logCollector, t, "Message with value was logged incorrectly.", "", "MessageWithValue1", "value")
	})
	t.Run("Message with multiple values", func(t *testing.T) {
		arrayValues := []string{"a", "b"}
		testCase(logCollector, t, "Message with multiple values was logged incorrectly.", "", "MessageWithMultiValue1", "value", nil, 23, 24.5, false, arrayValues)
	})
}

func TestLog_EnvSetInvalidValue(t *testing.T) {
	defer resetLogWriter(log.Default().Writer())
	var logCollector = &strings.Builder{}
	log.Default().SetOutput(logCollector)
	t.Setenv(LogUtilEnableTraceEnvironmentName, "InvalidValue")
	initialize()
	t.Run("Empty message", func(t *testing.T) {
		testCase(logCollector, t, "Empty message was logged incorrectly.", "", "")
	})
	t.Run("Simple message", func(t *testing.T) {
		testCase(logCollector, t, "Simple message was logged incorrectly.", "", "SimpleMessage1")
	})
	t.Run("Message with nil value", func(t *testing.T) {
		testCase(logCollector, t, "Message with nil value was logged incorrectly.", "", "MessageWithNil1", nil)
	})
	t.Run("Message with value", func(t *testing.T) {
		testCase(logCollector, t, "Message with value was logged incorrectly.", "", "MessageWithValue1", "value")
	})
	t.Run("Message with multiple values", func(t *testing.T) {
		arrayValues := []string{"a", "b"}
		testCase(logCollector, t, "Message with multiple values was logged incorrectly.", "", "MessageWithMultiValue1", "value", nil, 23, 24.5, false, arrayValues)
	})
}

func TestLog_TraceLogSet(t *testing.T) {
	defer resetLogWriter(log.Default().Writer())
	var logCollector = &strings.Builder{}
	log.Default().SetOutput(logCollector)
	t.Setenv(LogUtilEnableTraceEnvironmentName, LogUtilEnableTraceEnvironmentValue)
	initialize()
	t.Run("Empty message", func(t *testing.T) {
		testCase(logCollector, t, "Empty message was logged incorrectly.", "Trace Log Util: ", "")
	})
	t.Run("Simple message", func(t *testing.T) {
		testCase(logCollector, t, "Simple message could not be logged correctly.", "Trace Log Util: SimpleMessage2", "SimpleMessage2")
	})
	t.Run("Message with nil value", func(t *testing.T) {
		testCase(logCollector, t, "Message with nil value could not be logged.", "Trace Log Util: MessageWithNil2: <nil>", "MessageWithNil2", nil)
	})
	t.Run("Message with value", func(t *testing.T) {
		testCase(logCollector, t, "Message with value was logged incorrectly.", "Trace Log Util: MessageWithValue2: value2", "MessageWithValue2", "value2")
	})
	t.Run("Message with multiple values", func(t *testing.T) {
		arrayValues := []string{"a", "b"}
		testCase(logCollector, t, "Message with multiple values was logged incorrectly.", "Trace Log Util: MessageWithMultiValue2: value <nil> 23 24.5 false [a b]", "MessageWithMultiValue2", "value", nil, 23, 24.5, false, arrayValues)
	})
}

func TestLog_TraceLogAndDifferentFormatting(t *testing.T) {
	defer resetLogWriter(log.Default().Writer())
	var logCollector = &strings.Builder{}
	log.Default().SetOutput(logCollector)
	t.Setenv(LogUtilEnableTraceEnvironmentName, LogUtilEnableTraceEnvironmentValue)
	t.Setenv(LogUtilTraceFormat, "TLU: %s [%s]")
	initialize()
	t.Run("Empty message", func(t *testing.T) {
		testCase(logCollector, t, "Empty message was logged incorrectly.", "TLU:  []", "")
	})
	t.Run("Simple message", func(t *testing.T) {
		testCase(logCollector, t, "Simple message could not be logged correctly.", "TLU: SimpleMessage3 []", "SimpleMessage3")
	})
	t.Run("Message with nil value", func(t *testing.T) {
		testCase(logCollector, t, "Message with nil value could not be logged.", "TLU: MessageWithNil3 [<nil>]", "MessageWithNil3", nil)
	})
	t.Run("Message with value", func(t *testing.T) {
		testCase(logCollector, t, "Message with value was logged incorrectly.", "TLU: MessageWithValue3 [value3]", "MessageWithValue3", "value3")
	})
	t.Run("Message with multiple values", func(t *testing.T) {
		arrayValues := []string{"a", "b"}
		testCase(logCollector, t, "Message with multiple values was logged incorrectly.", "TLU: MessageWithMultiValue3 [value <nil> 23 24.5 false [a b]]", "MessageWithMultiValue3", "value", nil, 23, 24.5, false, arrayValues)
	})
}

func testCase(logCollector *strings.Builder, t *testing.T, testName string, expectedSubstring string, message string, values ...interface{}) {
	logCollector.Reset()
	if values == nil {
		Log(message)
	} else {
		Log(message, values...)
	}
	var loggedValue = logCollector.String()
	if !strings.Contains(loggedValue, expectedSubstring) {
		t.Error(testName, ", Expected ending with : ", expectedSubstring, ", Actual: ", loggedValue)
	}
	logCollector.Reset()
}
