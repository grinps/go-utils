package logger

import (
	"fmt"
	utils "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
)

const (
	NoLevelValue  int = -1
	AllLevelValue int = 1000
)

var NoLevel = &simpleLevel{
	level:     NoLevelValue,
	set:       true,
	levelName: "NoLevel",
} // NoLevel must be initialized explicitly since the checks will fail in NewLevel()

var Trace = newLevelNoError(100, "Trace")
var Debug = newLevelNoError(200, "Debug")
var Info = newLevelNoError(300, "Info")
var Warn = newLevelNoError(400, "Warn")
var ERROR = newLevelNoError(500, "Error")
var Fatal = newLevelNoError(600, "Fatal")
var ALL = &simpleLevel{
	level:     AllLevelValue,
	set:       true,
	levelName: "All",
} // All Level must be initialized explicitly since the checks will fail in NewLevel()

type simpleLevel struct {
	level     int
	set       bool
	levelName string
}

//TODO: Mutex
var levels = map[int]simpleLevel{}
var InvalidInputSimpleLevelErrorCode = errext.NewErrorCodeOfType(1, "SimpleLevelErrorCodes")
var DuplicateSimpleLevelErrorCode = errext.NewErrorCodeOfType(2, "SimpleLevelErrorCodes")

func (level *simpleLevel) String() string {
	if level.set {
		return level.levelName
	} else {
		return "NotSet"
	}
}

func (level *simpleLevel) Level() int {
	return level.level
}

func (level *simpleLevel) Compare(level2 Level) CompareResult {
	var returnValue CompareResult = NotApplicable
	var levelValue = -1
	if !level.set {
		utils.Log("Can not compare given Level to uninitialized simpleObject")
		return NotApplicable
	}
	if passedLevel, ok := level2.(interface {
		Level() int
	}); ok {
		if passedSimpleLevel, ok := passedLevel.(*simpleLevel); ok {
			if passedSimpleLevel.set {
				levelValue = passedSimpleLevel.level
			} else {
				utils.Log("Can not compare simpleLevel with simpleLevel if value is not set on given object", "this", level.levelName, "comparedTo", level2)
				return NotApplicable
			}
		} else {
			levelValue = passedLevel.Level()
		}
	} else {
		utils.Log("Can not compare Level with simpleLevel if it does not contain Level() int method", "this", level.levelName, "comparedTo", level2)
		return NotApplicable
	}
	if level.level < levelValue {
		returnValue = Less
	} else if level.level > levelValue {
		returnValue = Greater
	} else {
		returnValue = Equal
	}
	return returnValue
}

func newLevelNoError(level int, levelName string) *simpleLevel {
	newLevel, err := NewLevel(level, levelName)
	if err != nil {
		utils.Log("Failed to create new simple level", "level", level, "levelName", levelName, "err", err)
		return nil
	}
	return newLevel
}

func NewLevel(level int, levelName string) (*simpleLevel, error) {
	if existingLevel, ok := levels[level]; ok {
		if existingLevel.levelName == levelName {
			return &existingLevel, nil
		}
		return &existingLevel, DuplicateSimpleLevelErrorCode.New(fmt.Sprintf("Level at %d already exists with name %s. Can not create new level with name %s", level, existingLevel.levelName, levelName))
	}
	if level <= NoLevelValue {
		return nil, InvalidInputSimpleLevelErrorCode.New(fmt.Sprintf("Level can not be less than NoLevel level (%d)", NoLevelValue))
	}
	if level >= AllLevelValue {
		return nil, InvalidInputSimpleLevelErrorCode.New(fmt.Sprintf("Level can not be greater than All level (%d)", AllLevelValue))
	}
	newLevel := simpleLevel{
		level:     level,
		levelName: levelName,
		set:       true,
	}
	levels[level] = newLevel
	//utils.Log("New level created", "newLevel", newLevel)
	return &newLevel, nil
}
