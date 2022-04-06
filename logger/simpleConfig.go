package logger

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	utils "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

const (
	ConfigReaderResolveLocationErrorCode = "ConfigReaderResolveLocationErrorCode"
	ConfigPopulateErrorCode              = "ConfigPopulateErrorCode"
	ConfigReaderParseErrorCode           = "ConfigReaderParseErrorCode"
	EnvironmentLogConfigLocation         = "GO_UTL_LOG_CFG_DIR"
)

type Format int
type LocationType int

const (
	FormatUnknown Format = iota
	FormatXML
	FormatJSON
	FormatYAML
	FormatUnsupported
	LocationTypeUnknown LocationType = iota
	LocationURL
	LocationFile
	LocationValue
	LocationUnsupported
)

type ConfigSource struct {
	LocationType LocationType
	Location     string
	Format       Format
}

var ConfigPopulateMissingSourceErrorCode = errext.NewErrorCodeOfType(1, ConfigPopulateErrorCode)
var ConfigPopulateResolveLocationErrorCode = errext.NewErrorCodeOfType(2, ConfigPopulateErrorCode)
var ConfigPopulateReadConfigErrorCode = errext.NewErrorCodeOfType(3, ConfigPopulateErrorCode)
var ConfigPopulateParseConfigErrorCode = errext.NewErrorCodeOfType(4, ConfigPopulateErrorCode)

func (source *ConfigSource) PopulateConfigFromSource(logConfig LogConfig) error {
	utils.Log("Entering PopulateConfigFromSource", "logConfig", logConfig)
	var returnError error = nil
	if source != nil {
		if reader, err := ResolveLocation(source.LocationType, source.Location); err == nil {
			if configData, configReadError := io.ReadAll(reader); configReadError == nil && configData != nil {
				if _, implementsCloser := reader.(io.Closer); implementsCloser {
					closeErr := reader.(io.Closer).Close()
					if closeErr != nil {
						utils.Log("Ignoring close error associated with reading config", "closeErr", closeErr)
					}
				}
				if _, parsingError := ParseConfig(source.Format, configData, logConfig); parsingError == nil {
					utils.Log("Config data populated.", "config", logConfig)
					returnError = nil
				} else {
					returnError = ConfigPopulateParseConfigErrorCode.NewWithErrorF(parsingError, "Failed to parse configuration", "source", source, "config", string(configData), "logConfig", logConfig)
				}
			} else {
				returnError = ConfigPopulateReadConfigErrorCode.NewWithErrorF(configReadError, "Failed to read content from reader", "reader", reader, "source", source, "configData", configData)
			}
		} else {
			returnError = ConfigPopulateResolveLocationErrorCode.NewWithErrorF(err, "Failed to resolve location", "source", source)
		}
	} else {
		returnError = ConfigPopulateMissingSourceErrorCode.NewF("PopulateConfigFromSource called on nil object.")
	}
	utils.Log("Exiting PopulateConfigFromSource", "returnError", returnError)
	return returnError
}

var ResolveLocationReadErrorCode = errext.NewErrorCodeOfType(1, ConfigReaderResolveLocationErrorCode)
var ResolveLocationHTTPGETFailedErrorCode = errext.NewErrorCodeOfType(2, ConfigReaderResolveLocationErrorCode)
var ResolveLocationInvalidSchemeErrorCode = errext.NewErrorCodeOfType(3, ConfigReaderResolveLocationErrorCode)
var ResolveLocationNotSupportedErrorCode = errext.NewErrorCodeOfType(4, ConfigReaderResolveLocationErrorCode)
var ResolveLocationValueNotCreatedErrorCode = errext.NewErrorCodeOfType(5, ConfigReaderResolveLocationErrorCode)

func ResolveLocation(locationType LocationType, location string) (io.Reader, error) {
	utils.Log("Entering ResolveLocation", "location", location)
	var returnedLocationReader io.Reader = nil
	var returnedError error = ResolveLocationReadErrorCode.New("Unknown error occurred.")
	switch locationType {
	case LocationTypeUnknown:
		if parsedURL, parsingError := url.Parse(location); parsingError == nil && parsedURL.Scheme != "" {
			returnedLocationReader, returnedError = ResolveLocation(LocationURL, location)
		} else if fileReader, fileReaderErr := ResolveLocation(LocationFile, location); fileReaderErr == nil {
			returnedLocationReader = fileReader
			returnedError = nil
		} else {
			returnedError = ResolveLocationNotSupportedErrorCode.NewF("Could not resolve the location as either URL, file or content.", "ErrorForURL", parsingError, "ErrorForFile", fileReaderErr)
		}
		break
	default:
		returnedError = ResolveLocationNotSupportedErrorCode.New(fmt.Sprintf("Location %v is not currently supported.", locationType))
		break
	case LocationValue:
		returnedLocationReader = strings.NewReader(location)
		if returnedLocationReader != nil {
			returnedError = nil
		} else {
			returnedError = ResolveLocationValueNotCreatedErrorCode.New(fmt.Sprintf("Could not create a reader using content %s. No error was identified", location))
		}
		break
	case LocationURL:
		utils.Log("Checking if the location is URL.")
		if parsedURL, parsingError := url.Parse(location); parsingError == nil && parsedURL.Scheme != "" {
			utils.Log("Trying to resolve location as a URL", "parsedURL", parsedURL)
			urlScheme := strings.ToLower(parsedURL.Scheme)
			utils.Log("URL scheme identified", "urlScheme", urlScheme)
			switch urlScheme {
			case "http", "https":
				{
					utils.Log("Trying to perform HTTP(s) GET", "parsedURL", parsedURL)
					response, httpError := http.Get(location)
					utils.Log("HTTP(s) GET completed", "response", response, "httpError", httpError)
					if httpError != nil {
						returnedError = ResolveLocationHTTPGETFailedErrorCode.NewWithError(fmt.Sprintf("Failed to perform HTTP(s) GET on URL %s", location), httpError)
					} else {
						returnedLocationReader = response.Body
						returnedError = nil
					}
					break
				}
			default:
				returnedError = ResolveLocationInvalidSchemeErrorCode.New(fmt.Sprintf("Can not resolve location %s. URL scheme %s not supported", location, urlScheme))
			}
		} else {
			returnedError = ResolveLocationInvalidSchemeErrorCode.NewWithErrorF(parsingError, "Can not resolve location", "location", location, "")
		}
		break
	case LocationFile:
		utils.Log("Trying to resolve as local file location")
		if fileReader, fileExists := FileExistsMultiLocation(location,
			func() string { return os.Getenv(EnvironmentLogConfigLocation) },
			func() string { workDir, _ := os.Getwd(); return workDir },
			func() string {
				if exeDir, extracted := ExecutableDirectory(); extracted {
					return exeDir
				} else {
					return ""
				}
			},
			func() string {
				if exeDir, extracted := ExecutableDirectory(); extracted {
					return exeDir + "../config"
				} else {
					return ""
				}
			},
		); fileExists {
			returnedLocationReader = fileReader
			returnedError = nil
		} else {
			returnedError = ResolveLocationReadErrorCode.New(fmt.Sprintf("Could not resolve location %s as file.", location))
		}
		break
	}
	//TODO debug why returning nil does not work for readerObject == nil
	utils.Log("Leaving ResolveLocation", "reader", returnedLocationReader, "error", returnedError)
	return returnedLocationReader, returnedError
}

func FileExistsMultiLocation(location string, prefixFunctions ...func() string) (io.Reader, bool) {
	utils.Log("Entering FileExistsMultiLocation", "prefixFunctions", prefixFunctions, "location", location)
	var fileReader io.Reader = nil
	var fileExists = false
	if prefixFunctions != nil {
		for index, prefixFunction := range prefixFunctions {
			if prefixFunction != nil {
				prefixValue := prefixFunction()
				if prefixValue != "" {
					internalLocation := filepath.Join(prefixValue, location)
					if localFileReader, localFileExists := FileExists(internalLocation); localFileExists {
						utils.Log("Located file", "index", index, "internalLocation", internalLocation)
						fileReader = localFileReader
						fileExists = true
						break
					} else {
						utils.Log("File was not located", "index", index, "internalLocation", internalLocation)
					}
				} else {
					utils.Log("Prefix value returned as empty", "index", index)
				}
			} else {
				utils.Log("Ignoring nil prefix function", "index", index)
			}
		}
	}
	if !fileExists {
		fileReader, fileExists = FileExists(location)
	}
	utils.Log("Leaving FileExistsMultiLocation", "fileReader", fileReader, "fileExists", fileExists)
	return fileReader, fileExists
}

func FileExists(location string) (io.Reader, bool) {
	utils.Log("Entering FileExists", "location", location)
	var fileReader io.Reader = nil
	var fileExists = false
	fileDetails, fileStatError := os.Stat(location)
	utils.Log("Completed stat location", "fileDetails", fileDetails, "fileStatError", fileStatError)
	if fileStatError == nil {
		if fileDetails.IsDir() {
			utils.Log("File exists but it is a directory")
		} else {
			utils.Log("Trying to open file")
			file, fileOpenError := os.Open(location)
			utils.Log("Completed open file", "file", file, "fileOpenError", fileOpenError)
			if fileOpenError == nil {
				fileReader = file
				fileExists = true
			}
		}
	} else if os.IsNotExist(fileStatError) {
		utils.Log("File does not exist")
	} else {
		utils.Log("File existence could not be determined due to error in stat", "fileStatError", fileStatError)
	}
	utils.Log("Leaving FileExists", "fileReader", fileReader, "fileExists", fileExists)
	return fileReader, fileExists
}

func ExecutableDirectory() (string, bool) {
	utils.Log("Entering ExecutableDirectory")
	var executableLocation = ""
	var executableDirResolved = false
	executable, executableReadError := os.Executable()
	utils.Log("Executable located", "executable", executable, "executableReadError", executableReadError)
	if executableReadError == nil {
		executableFileDir, executableFileDirErr := filepath.Abs(filepath.Dir(executable))
		if executableFileDirErr == nil {
			executableLocation = executableFileDir
			executableDirResolved = true
		} else {
			utils.Log("Resolving directory failed", "executableFileDirErr", executableFileDirErr)
		}
	} else {
		utils.Log("Executable location could not be located", "executableReadError", executableReadError)
	}
	utils.Log("Leaving ExecutableDirectory", "executableLocation", executableLocation, "executableDirResolved", executableDirResolved)
	return executableLocation, executableDirResolved
}

var ConfigReaderParseInvalidFormatErrorCode = errext.NewErrorCodeOfType(1, ConfigReaderParseErrorCode)
var ConfigReaderParseParsingErrorCode = errext.NewErrorCodeOfType(2, ConfigReaderParseErrorCode)

func ParseConfig(format Format, configData []byte, logConfig LogConfig) (LogConfig, error) {
	var returnLogConfig = logConfig
	var returnError = ConfigReaderParseParsingErrorCode.New("Unknown error occurred.")
	switch format {
	case FormatUnknown:
		for currentFormat := FormatUnknown + 1; currentFormat < FormatUnsupported; currentFormat++ {
			if _, parsingError := ParseConfig(currentFormat, configData, logConfig); parsingError == nil {
				returnError = nil
				break
			}
		}
	case FormatJSON:
		jasonParseError := json.Unmarshal(configData, logConfig)
		if jasonParseError != nil {
			utils.Log("Failed to parse config as JSON", "jasonParseError", jasonParseError)
			returnError = ConfigReaderParseParsingErrorCode.NewWithErrorF(jasonParseError, "Failed to parse configuration as JSON")
		} else {
			returnError = nil
		}
		break
	case FormatXML:
		xmlParseError := xml.Unmarshal(configData, logConfig)
		if xmlParseError != nil {
			utils.Log("Failed to parse config as XML", "jasonParseError", xmlParseError)
			returnError = ConfigReaderParseParsingErrorCode.NewWithErrorF(xmlParseError, "Failed to parse configuration as XML")
		} else {
			returnError = nil
		}
		break
	default:
		returnError = ConfigReaderParseInvalidFormatErrorCode.NewF("Unsupported format", "format", format)
		break
	}
	return returnLogConfig, returnError
}
