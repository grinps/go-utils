package logger

import (
	"fmt"
	utils "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
	"io"
	golog "log"
	"os"
)

const (
	GoLogConfigName                = "GoLogConfig"
	GoLogDriverName                = "GoLogDriver"
	GoLogDriverInitErrorCode       = "GoLogDriverInitError"
	GoLogConfigDefaultPrefixFormat = "[%s]:[%s]:%s %s"
)

type GoLogConfig struct {
	Prefix              string `json:"prefix"`
	OutputFile          string `json:"outputFile"`
	outputFileReference io.Writer
	Flags               GoLogFlags `json:"flags"`
	flags               int
	PrefixFormat        string `json:"prefixFormat"`
	populated           bool
	logConfigName       string
}

type GoLogFlags struct {
	Date                bool `json:"date"`
	Time                bool `json:"time"`
	TimeInMicrosecond   bool `json:"timeInMicrosecond"`
	TimeInUTC           bool `json:"timeInUTC"`
	LongFile            bool `json:"longFile"`
	ShortFile           bool `json:"shortFile"`
	PrefixAtStartOfLine bool `json:"prefixAtStartOfLine"`
}

func (logConfig *GoLogConfig) GetLogDriver() LogDriver {
	return goLogDriverInstance
}

func (logConfig *GoLogConfig) String() string {
	if logConfig == nil || logConfig.logConfigName == "" {
		return GoLogConfigName
	} else {
		return logConfig.logConfigName
	}
}

func (logConfig *GoLogConfig) GetLevel() Level {
	return Info
}

func (logConfig *GoLogConfig) getWriter() io.Writer {
	if logConfig == nil {
		return nil
	}
	//TODO Add mutex to avoid multiple open calls.
	if logConfig.outputFileReference == nil {
		if logConfig.OutputFile != "" {
			if openedLogFile, openFileErr := os.OpenFile(logConfig.OutputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644); openFileErr == nil {
				logConfig.outputFileReference = openedLogFile
			} else {
				utils.Log("Failed to open output file in create/append with writeonly mode and 0644 permission. Using Stderr as Writer", "error", openFileErr, "logConfig", logConfig)
				logConfig.outputFileReference = os.Stderr
			}
		} else {
			utils.Log("Using Stderr as Writer", "logConfig", logConfig)
			logConfig.outputFileReference = os.Stderr
		}
	}
	return logConfig.outputFileReference
}

func (logConfig *GoLogConfig) getFlag() int {
	if logConfig != nil && logConfig.flags == 0 {
		if logConfig.Flags.Date {
			logConfig.flags = logConfig.flags | golog.Ldate
		}
		if logConfig.Flags.Time {
			logConfig.flags = logConfig.flags | golog.Ltime
		}
		if logConfig.Flags.Time && logConfig.Flags.TimeInMicrosecond {
			logConfig.flags = logConfig.flags | golog.Lmicroseconds
		}
		if logConfig.Flags.LongFile && !logConfig.Flags.ShortFile {
			logConfig.flags = logConfig.flags | golog.Llongfile
		}
		if logConfig.Flags.ShortFile && !logConfig.Flags.LongFile {
			logConfig.flags = logConfig.flags | golog.Lshortfile
		}
		if logConfig.Flags.PrefixAtStartOfLine {
			logConfig.flags = logConfig.flags | golog.Lmsgprefix
		}
		if logConfig.Flags.TimeInUTC {
			logConfig.flags = logConfig.flags | golog.LUTC
		}
		if logConfig.flags == 0 {
			logConfig.flags = golog.LstdFlags
			utils.Log("Go Log config does not have any flags enabled. Enabling default flags", "logConfig", logConfig)
		}
	}
	if logConfig != nil {
		return logConfig.flags
	} else {
		return golog.LstdFlags
	}
}

var goLogDriverInstance = &goLogDriver{}

type goLogDriver struct {
}

var GoLogDriverInitializationFailedErrorCode = errext.NewErrorCodeOfType(1, GoLogDriverInitErrorCode)
var GoLogDriverInitializationMissingConfigErrorCode = errext.NewErrorCodeOfType(2, GoLogDriverInitErrorCode)
var GoLogDriverInitializationInvalidLogConfigTypeErrorCode = errext.NewErrorCodeOfType(3, GoLogDriverInitErrorCode)
var GoLogDriverInitializationNotPopulatedErrorCode = errext.NewErrorCodeOfType(4, GoLogDriverInitErrorCode)

func (logDriver *goLogDriver) Initialize(loggerName string, config LogConfig) (Logger, error) {
	utils.Log("Entering goLogDriver.Initialize", "loggerName", loggerName, "config", config)
	var returnError error = GoLogDriverInitializationFailedErrorCode.New("Unknown error occurred.")
	var returnLogger Logger = nil
	if config == nil {
		returnError = GoLogDriverInitializationMissingConfigErrorCode.New("Configuration was passed as nil.")
		return returnLogger, returnError
	}
	if goLogConfig, ok := config.(*GoLogConfig); !ok {
		returnError = GoLogDriverInitializationInvalidLogConfigTypeErrorCode.New(fmt.Sprintf("Invalid Log config passed. Expected type GoLogConfig for object %v", config))
	} else if !goLogConfig.populated {
		returnError = GoLogDriverInitializationNotPopulatedErrorCode.NewF("Log config passed is not populated.", "config", config)
	} else {
		var goLogger *GoLogger = NewGoLogger(loggerName, goLogConfig)
		returnLogger = NewSimpleLogWrapper(loggerName, nil, goLogConfig, goLogger)
		returnError = nil
	}
	utils.Log("Exiting goLogDriver.Initialize", "returnLogger", returnLogger, "returnError", returnError)
	return returnLogger, returnError
}

func (logDriver *goLogDriver) GetName() string {
	return GoLogDriverName
}

type GoLogger struct {
	config *GoLogConfig
	logger *golog.Logger
}

func (goLogger *GoLogger) Log(marker Marker, level Level, message string, v ...interface{}) {
	err := goLogger.logger.Output(4, fmt.Sprintf(goLogger.config.PrefixFormat, level, marker, message, fmt.Sprintln(v...)))
	if err != nil {
		utils.Log("Failed to log error to go logger", "err", err)
	}
}

func NewGoLogger(loggerName string, config *GoLogConfig) *GoLogger {
	var returnGoLogger *GoLogger = &GoLogger{
		config: defaultLogConfig,
		logger: golog.Default(),
	}
	if config != nil && config.populated {
		returnGoLogger.logger = golog.New(config.getWriter(), config.Prefix, config.getFlag())
		returnGoLogger.config = config
		if config.PrefixFormat == "" {
			config.PrefixFormat = GoLogConfigDefaultPrefixFormat
		}
	}
	return returnGoLogger
}
