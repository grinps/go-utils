package logger

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path"
	"testing"
)

func TestExecutableDirectory(t *testing.T) {
	value, ok := ExecutableDirectory()
	t.Log("executableDirectory", value, "ok", ok)
	if !ok {
		t.Error("Executable Directory should always be returned.")
	}
}

func TestFileExists(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Error("Failed to read Executable can not continue", "err", err)
	}
	if executable != "" {
		t.Run("ValidFileExists", func(t *testing.T) {
			if _, ok := FileExists(executable); !ok {
				t.Error("Executable file should always exist", "executable", executable)
			}
		})
		t.Run("ValidDirectoryIsInvalidFile", func(t *testing.T) {
			executableDirectory := path.Dir(executable)
			if _, ok := FileExists(executableDirectory); ok {
				t.Error("Executable file directory should be identified as exists", "executableDirectory", executableDirectory)
			}
		})
		t.Run("InvalidFileNotExists", func(t *testing.T) {
			invalidFile := executable + "7483843874"
			if _, ok := FileExists(invalidFile); ok {
				t.Error("Invalid file should not be identified as exists", "invalidFile", invalidFile)
			}
		})
	}
	t.Run("EmptyFileCheck", func(t *testing.T) {
		var emptyFileName string
		if _, ok := FileExists(emptyFileName); ok {
			t.Error("Empty file name should not be identified as exists", "emptyFileName", emptyFileName)
		}
	})
}

func TestFileExistsMultiLocation(t *testing.T) {
	executable, _ := os.Executable()
	fileName := path.Base(executable)
	baseDir := path.Dir(executable)
	tempFileName := "./" + fileName
	tempFile, err := os.Create(tempFileName)
	if err == nil {
		_ = tempFile.Close()
		defer func() {
			t.Log("Trying to delete local file", "tempFileName", tempFileName)
			err := os.Remove(tempFileName)
			if err != nil {
				t.Log("Failed to delete local file", "tempFileName", tempFileName, "err", err)
			}
		}()
	}
	t.Run("EmptyFileCheck", func(t *testing.T) {
		var emptyFileName string
		if _, ok := FileExistsMultiLocation(emptyFileName); ok {
			t.Error("Empty file name should not be identified as exists", "emptyFileName", emptyFileName)
		}
	})
	t.Run("EmptyFileNilPrefixes", func(t *testing.T) {
		var emptyFileName string
		if _, ok := FileExistsMultiLocation(emptyFileName, nil); ok {
			t.Error("Empty file name with nil prefixes should not be identified as exists", "emptyFileName", emptyFileName)
		}
	})
	t.Run("ValidFileNoLocationCheck", func(t *testing.T) {
		if _, ok := FileExistsMultiLocation(executable); !ok {
			t.Error("Valid executable file name should be identified as exists", "executable", executable)
		}
	})
	t.Run("ValidFileMultipleNilPrefix", func(t *testing.T) {
		if _, ok := FileExistsMultiLocation(executable, nil, nil); !ok {
			t.Error("Valid executable file name with nil prefix should be identified as exists", "executable", executable)
		}
	})
	t.Run("ValidFileMultiplePrefixOneValid", func(t *testing.T) {
		if _, ok := FileExistsMultiLocation(fileName, nil, func() string {
			return ""
		}, func() string {
			return baseDir + "/invalid"
		}, nil, func() string {
			return "./"
		}); !ok {
			t.Error("Must support empty return, nil, invalid and Valid prefix and file should be identified as exists", "executable", executable)
		}
	})
	t.Run("MultipleValidLocationFirstNonEmptyFile", func(t *testing.T) {
		if readerObject, ok := FileExistsMultiLocation(fileName, func() string {
			return baseDir
		}, func() string {
			return "./"
		}); !ok {
			t.Error("Must support multiple valid prefixes and file should be identified as exists", "baseDir", baseDir, "local", "./")
		} else {
			var dataRead = make([]byte, 10)
			count, readErr := readerObject.Read(dataRead)
			t.Log("count", count, "readErr", readErr)
			if count < 10 || readErr == io.EOF {
				t.Error("Must have read executable rather than local empty file")
			}
		}
	})
	t.Run("MultipleValidLocationFirstEmptyFile", func(t *testing.T) {
		if readerObject, ok := FileExistsMultiLocation(fileName, func() string {
			return "./"
		}, func() string {
			return baseDir
		}); !ok {
			t.Error("Must support multiple valid prefixes and file should be identified as exists", "baseDir", baseDir, "local", "./")
		} else {
			var dataRead = make([]byte, 10)
			count, readErr := readerObject.Read(dataRead)
			t.Log("count", count, "readErr", readErr)
			if count == 10 || readErr != io.EOF {
				t.Error("Must have read local empty file rather than executable")
			}
		}
	})
}

func TestResolveLocation(t *testing.T) {
	t.Run("EmptyLocation", func(t *testing.T) {
		var emptyLocation string
		if _, err := ResolveLocation(LocationTypeUnknown, emptyLocation); err == nil {
			t.Error("Empty location should not be identified as exists", "emptyLocation", emptyLocation)
		}
	})
	t.Run("ValidURL", func(t *testing.T) {
		var googleLocation = "https://www.google.com"
		if reader, err := ResolveLocation(LocationTypeUnknown, googleLocation); err != nil {
			t.Error("URL location should be resolved", "googleLocation", googleLocation)
		} else {
			var readData = make([]byte, 10)
			count, readErr := reader.Read(readData)
			if count < 10 || readErr != nil {
				t.Error("URL should allow reading of 10 bytes", "googleLocation", googleLocation, "count", count, "readErr", readErr)
			}
		}
	})
	t.Run("InvalidURL", func(t *testing.T) {
		var googleInvalidLocation = "https://www.g00gle.com"
		readerObject, err := ResolveLocation(LocationTypeUnknown, googleInvalidLocation)
		if err == nil {
			t.Error("URL location should not be resolved", "googleInvalidLocation", googleInvalidLocation)
		}
		t.Log("readerObject", readerObject)
		if readerObject != nil {
			var readData = make([]byte, 10)
			count, readErr := readerObject.Read(readData)
			if count > 0 || readErr == nil {
				t.Error("Invalid URL should not allow reading of 10 bytes", "googleInvalidLocation", googleInvalidLocation, "count", count, "readErr", readErr)
			}
		}
	})
	t.Run("InvalidURLScheme", func(t *testing.T) {
		var googleInvalidScheme = "ssh://www.g00gle.com"
		readerObject, err := ResolveLocation(LocationTypeUnknown, googleInvalidScheme)
		if err == nil {
			t.Error("URL location should not be resolved", "googleInvalidScheme", googleInvalidScheme)
		}
		t.Log("readerObject", readerObject)
		if readerObject != nil {
			var readData = make([]byte, 10)
			count, readErr := readerObject.Read(readData)
			if count > 0 || readErr == nil {
				t.Error("Invalid URL should not allow reading of 10 bytes", "googleInvalidScheme", googleInvalidScheme, "count", count, "readErr", readErr)
			}
		}
	})
	t.Run("InvalidFileLocation", func(t *testing.T) {
		var invalidFileName = "myConfig.properties"
		if readerObject, err := ResolveLocation(LocationTypeUnknown, invalidFileName); err == nil {
			t.Error("URL location should not be resolved", "invalidFileName", invalidFileName)
		} else if readerObject != nil {
			var readData = make([]byte, 10)
			count, readErr := readerObject.Read(readData)
			if count > 0 || readErr == nil {
				t.Error("Invalid URL should not allow reading of 10 bytes", "invalidFileName", invalidFileName, "count", count, "readErr", readErr)
			}
		}
	})
	configWdFileName := "myconfigwd.temp"
	workDirLocation, _ := os.Getwd()
	workDirConfigLocation := path.Join(workDirLocation, configWdFileName)
	tempWorkFile, err := os.Create(workDirConfigLocation)
	if err == nil {
		_ = tempWorkFile.Close()
	}
	t.Run("ValidFileLocationWorkDirectory", func(t *testing.T) {
		if readerObject, err := ResolveLocation(LocationTypeUnknown, configWdFileName); err != nil {
			t.Error("Config file created in wd should be resolved", "configWdFileName", configWdFileName, "workDirConfigLocation", workDirConfigLocation)
		} else if readerObject != nil {
			var readData = make([]byte, 10)
			count, readErr := readerObject.Read(readData)
			if count > 0 || readErr != io.EOF {
				t.Error("Should not be able to read anything from empty file", "configWdFileName", configWdFileName, "count", count, "readErr", readErr)
			}
			closeErr := readerObject.(io.Closer).Close()
			if closeErr != nil {
				t.Log("Failed to close open workdir config file", "closeErr", closeErr)
			}
			deleteErr := os.Remove(workDirConfigLocation)
			if deleteErr != nil {
				t.Log("Failed to delete workdir config file", "deleteErr", deleteErr)
			}
		}
	})

	configExeFileName := "myconfigexe.temp"
	executableDir, _ := ExecutableDirectory()
	executableConfigLocation := path.Join(executableDir, configExeFileName)
	tempExeFile, err1 := os.Create(executableConfigLocation)
	if err1 == nil {
		_ = tempExeFile.Close()
	}
	t.Run("ValidFileLocationExecutableDirectory", func(t *testing.T) {
		if readerObject, err := ResolveLocation(LocationTypeUnknown, configExeFileName); err != nil {
			t.Error("Config file created in executable directory should be resolved", "configExeFileName", configExeFileName, "executableConfigLocation", executableConfigLocation)
		} else if readerObject != nil {
			var readData = make([]byte, 10)
			count, readErr := readerObject.Read(readData)
			if count > 0 || readErr != io.EOF {
				t.Error("Should not be able to read anything from empty file", "configWdFileName", configWdFileName, "count", count, "readErr", readErr)
			}
			closeErr := readerObject.(io.Closer).Close()
			if closeErr != nil {
				t.Log("Failed to close open executable config file", "closeErr", closeErr)
			}
			deleteErr := os.Remove(executableConfigLocation)
			if deleteErr != nil {
				t.Log("Failed to delete executable config file", "deleteErr", deleteErr)
			}
		}
	})
	t.Run("ValidContent", func(t *testing.T) {
		if readerObject, err := ResolveLocation(LocationValue, "{}"); err != nil || readerObject == nil {
			t.Error("The content should be successful resolved.", "err", err)
		} else {
			var readData = make([]byte, 2)
			count, readErr := readerObject.Read(readData)
			if count != 2 || readErr == io.EOF {
				t.Error("Should be able to read from content", "count", count, "readErr", readErr)
			}
			if "{}" != string(readData) {
				t.Error("Read data is not expected value", "expected", "{}", "got bytes", readData, "as string", string(readData))
			}
			//TODO: The string reader does not provide close implementation.
		}
	})
	t.Run("InvalidLocationType", func(t *testing.T) {
		var locationType = LocationUnsupported
		if readerObject, err := ResolveLocation(locationType, "UnsupportedLocation"); err == nil {
			t.Error("Expected an error to be returned since the location type provided is not supported.", "locationType", locationType)
		} else if readerObject != nil {
			t.Error("Did not expect reader object to be not-nil since the location type is unsupported.", "locationType", locationType)
		} else if mappedError, isMappedError := ResolveLocationNotSupportedErrorCode.AsError(err); !isMappedError {
			t.Error("Expected ResolveLocationNotSupportedErrorCode error.", "mappedError", mappedError)
		}
	})
}

type innerObject struct {
	AnotherStringValue string
}
type JSONLogConfig struct {
	XMLName     xml.Name `xml:"config"`
	StringValue string
	IntValue    int
	BoolValue   bool
	ObjectValue *innerObject
}

func (J JSONLogConfig) GetLogDriver() LogDriver {
	panic("implement me")
}

func (J JSONLogConfig) String() string {
	return fmt.Sprintf("JSONLogConfig(StringValue: %s, IntValue: %d, BoolValue: %t, ObjectValue: %v)", J.StringValue, J.IntValue, J.BoolValue, J.ObjectValue)
}

var sampleJSONConfig = "{\"StringValue\": \"aStringValue\", \"IntValue\": 1234, \"BoolValue\": true, \"ObjectValue\": { \"AnotherStringValue\" :\"AnotherStringValue\"} }"
var sampleXMLConfig = "<?xml version=\"1.0\" encoding=\"UTF-8\"?><config><StringValue>aStringValue</StringValue><IntValue>1234</IntValue><BoolValue>true</BoolValue><ObjectValue><AnotherStringValue>AnotherStringValue</AnotherStringValue></ObjectValue></config>"

func TestParseConfig(t *testing.T) {
	t.Run("FormatNotSupported", func(t *testing.T) {
		var formatType = FormatUnsupported + 1
		if returnedLogConfig, err := ParseConfig(formatType, []byte("{}"), nil); err == nil {
			t.Error("Expected an error to be returned since the format type provided is not supported.", "formatType", formatType)
		} else if returnedLogConfig != nil {
			t.Error("Did not expect logConfig to be not-nil since the format type is unsupported.", "formatType", formatType, "returnedLogConfig", returnedLogConfig)
		} else if _, isMappedError := ConfigReaderParseInvalidFormatErrorCode.AsError(err); !isMappedError {
			t.Error("Expected ConfigReaderParseInvalidFormatErrorCode error.", "err", err)
		}
	})
	t.Run("FormatSupportedButNotImplemented", func(t *testing.T) {
		var formatType = FormatYAML
		if returnedLogConfig, err := ParseConfig(formatType, []byte("{}"), nil); err == nil {
			t.Error("Expected an error to be returned since the format type provided is not supported.", "formatType", formatType)
		} else if returnedLogConfig != nil {
			t.Error("Did not expect logConfig to be not-nil since the format type is unsupported.", "formatType", formatType, "returnedLogConfig", returnedLogConfig)
		} else if _, isMappedError := ConfigReaderParseInvalidFormatErrorCode.AsError(err); !isMappedError {
			t.Error("Expected ConfigReaderParseInvalidFormatErrorCode error.", "err", err)
		}
	})
	t.Run("FormatUnknownWithJSONValue", func(t *testing.T) {
		var formatType = FormatUnknown
		var jsonLogConfig = &JSONLogConfig{}
		jsonLogConfig.ObjectValue = &innerObject{}
		if returnedLogConfig, err := ParseConfig(formatType, []byte(sampleJSONConfig), jsonLogConfig); err != nil {
			t.Error("Though unknown format, system should have parsed the json config successfully.", "err", err)
		} else if returnedLogConfig != jsonLogConfig {
			t.Error("Did not expect returned logConfig to be different from provided log config.", "returnedLogConfig", returnedLogConfig)
		} else if jsonLogConfig.StringValue != "aStringValue" || jsonLogConfig.IntValue != 1234 || !jsonLogConfig.BoolValue ||
			jsonLogConfig.ObjectValue.AnotherStringValue != "AnotherStringValue" {
			t.Error("Did not expect logconfig to have incorrect value.", "jsonLogConfig", jsonLogConfig)
		}
	})
	t.Run("FormatXMLWithJSONValue", func(t *testing.T) {
		var formatType = FormatXML
		var jsonLogConfig = &JSONLogConfig{}
		jsonLogConfig.ObjectValue = &innerObject{}
		if _, err := ParseConfig(formatType, []byte(sampleJSONConfig), jsonLogConfig); err == nil {
			t.Error("Expected parsing to fail since wrong format was passed..")
		} else if _, isMappedError := ConfigReaderParseParsingErrorCode.AsError(err); !isMappedError {
			t.Error("Expected ConfigReaderParseParsingErrorCode error.", "err", err)
		}
	})
	t.Run("FormatXMLWithXMLValue", func(t *testing.T) {
		var formatType = FormatXML
		var jsonLogConfig = &JSONLogConfig{}
		jsonLogConfig.ObjectValue = &innerObject{}
		if returnedLogConfig, err := ParseConfig(formatType, []byte(sampleXMLConfig), jsonLogConfig); err != nil {
			t.Error("Expected parsing to succeed since XML format with valid string.", "err", err)
		} else if returnedLogConfig != jsonLogConfig {
			t.Error("Expected returnedLogConfig to match passed value.", "returnedLogConfig", returnedLogConfig)
		} else if jsonLogConfig.StringValue != "aStringValue" || jsonLogConfig.IntValue != 1234 || !jsonLogConfig.BoolValue ||
			jsonLogConfig.ObjectValue.AnotherStringValue != "AnotherStringValue" {
			t.Error("Did not expect logconfig to have incorrect value.", "jsonLogConfig", jsonLogConfig)
		}
	})
}

type SimpleLogConfig struct {
	name string
}

func (logConfig *SimpleLogConfig) GetLogDriver() LogDriver {
	return &NoOpsLogDriver{}
}
func (logConfig *SimpleLogConfig) String() string {
	return logConfig.name
}

func TestConfigSource_PopulateConfigFromSource(t *testing.T) {
	t.Run("DefaultObjectWithNilSource", func(t *testing.T) {
		defaultConfigSource := &ConfigSource{}
		err := defaultConfigSource.PopulateConfigFromSource(nil)
		if err == nil {
			t.Error("Expected the load from nil source to fail")
		} else if _, isMappedError := ConfigPopulateResolveLocationErrorCode.AsError(err); !isMappedError {
			t.Error("Expected ConfigPopulateResolveLocationErrorCode error.", "err", err)
		}
	})
	t.Run("DefaultObjectWithNoOpsLogger", func(t *testing.T) {
		defaultLogConfigSource := &ConfigSource{}
		testNoOpsLogConfig := &SimpleLogConfig{
			name: "TestNoOpsLogConfig",
		}
		if err := defaultLogConfigSource.PopulateConfigFromSource(testNoOpsLogConfig); err == nil {
			t.Error("Expected the load using default object to SimpleLogConfig config to fail")
		} else if _, isMappedError := ConfigPopulateResolveLocationErrorCode.AsError(err); !isMappedError {
			t.Error("Expected ConfigPopulateResolveLocationErrorCode error since the logConfigSource is not initialized.", "err", err)
		}
	})
	t.Run("DefaultObjectWithNoOpsLogger", func(t *testing.T) {
		defaultLogConfigSource := &ConfigSource{}
		testNoOpsLogConfig := &SimpleLogConfig{
			name: "TestNoOpsLogConfig",
		}
		if err := defaultLogConfigSource.PopulateConfigFromSource(testNoOpsLogConfig); err == nil {
			t.Error("Expected the load using default object to SimpleLogConfig config to fail")
		} else if _, isMappedError := ConfigPopulateResolveLocationErrorCode.AsError(err); !isMappedError {
			t.Error("Expected ConfigPopulateResolveLocationErrorCode error since the logConfigSource is not initialized.", "err", err)
		}
	})
	defaultLogConfigSource := &ConfigSource{
		LocationType: LocationValue,
		Location:     sampleJSONConfig,
		Format:       FormatJSON,
	}
	t.Run("InitializedObjectWithJSONLogConfig", func(t *testing.T) {
		jsonLogConfig := &JSONLogConfig{}
		err := defaultLogConfigSource.PopulateConfigFromSource(jsonLogConfig)
		if err != nil {
			t.Error("Expected the load using configured object to JSONLogConfig config to succeed", "error", err)
		} else if jsonLogConfig.StringValue != "aStringValue" || jsonLogConfig.IntValue != 1234 || !jsonLogConfig.BoolValue ||
			jsonLogConfig.ObjectValue.AnotherStringValue != "AnotherStringValue" {
			t.Error("Did not expect logconfig to have incorrect value.", "jsonLogConfig", jsonLogConfig)
		}
	})
	t.Run("InitializedObjectWithNullConfig", func(t *testing.T) {
		err := defaultLogConfigSource.PopulateConfigFromSource(nil)
		if err == nil {
			t.Error("Expected the load using configured object to nil config to fail")
		} else if _, isMappedError := ConfigPopulateParseConfigErrorCode.AsError(err); !isMappedError {
			t.Error("Expected ConfigPopulateParseConfigErrorCode error since the logConfigSource is nil.", "err", err)
		}
	})
	t.Run("CallOnNullObject", func(t *testing.T) {
		var nullLogConfigSource *ConfigSource = nil
		err := nullLogConfigSource.PopulateConfigFromSource(nil)
		if err == nil {
			t.Error("Expected the load on nil object to fail")
		} else if _, isMappedError := ConfigPopulateMissingSourceErrorCode.AsError(err); !isMappedError {
			t.Error("Expected ConfigPopulateMissingSourceErrorCode error since the logConfigSource is nil.", "err", err)
		}
	})
	t.Run("ConfigUsingFile", func(t *testing.T) {
		configWdFileName := "myconfig.deleteMe"
		workDirLocation, _ := os.Getwd()
		workDirConfigLocation := path.Join(workDirLocation, configWdFileName)
		if tempWorkFile, createFileErr := os.Create(workDirConfigLocation); createFileErr != nil {
			t.Error("Test failed since the config file can not be created.", createFileErr)
		} else if _, writeFileErr := tempWorkFile.WriteString(sampleJSONConfig); writeFileErr != nil {
			t.Error("Failed to write to config file.", writeFileErr)
		} else {
			_ = tempWorkFile.Close()
			defer func() {
				t.Log("Trying to delete local file", "workDirConfigLocation", workDirConfigLocation)
				err := os.Remove(workDirConfigLocation)
				if err != nil {
					t.Log("Failed to delete local file", "workDirConfigLocation", workDirConfigLocation, "err", err)
				}
			}()
		}
		var initConfigSource *ConfigSource = &ConfigSource{
			LocationType: LocationFile,
			Location:     configWdFileName,
			Format:       FormatJSON,
		}
		jsonLogConfig := &JSONLogConfig{}
		err := initConfigSource.PopulateConfigFromSource(jsonLogConfig)
		if err != nil {
			t.Error("Expected the load to work.")
		} else if jsonLogConfig.StringValue != "aStringValue" || jsonLogConfig.IntValue != 1234 || !jsonLogConfig.BoolValue ||
			jsonLogConfig.ObjectValue.AnotherStringValue != "AnotherStringValue" {
			t.Error("Did not expect logconfig to have incorrect value.", "jsonLogConfig", jsonLogConfig)
		}
	})
}
