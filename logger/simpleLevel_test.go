package logger

import "testing"

func TestDefaultSimpleLevel(t *testing.T) {
	var simpleLevel1 Level = &simpleLevel{}
	t.Run("DefaultCompareToAll", func(t *testing.T) {
		if simpleLevel1.Compare(ALL) != NotApplicable {
			t.Error("ALL is greater than all levels")
		}
	})
	t.Run("DefaultCompareToAllReverse", func(t *testing.T) {
		if ALL.Compare(simpleLevel1) != NotApplicable {
			t.Error("ALL is greater than all levels")
		}
	})
	t.Run("DefaultCompareToNone", func(t *testing.T) {
		if NoLevel.Compare(simpleLevel1) != NotApplicable {
			t.Error("NoLevel is less than all levels")
		}
	})
	t.Run("DefaultNameIsNotSet", func(t *testing.T) {
		if simpleLevel1.String() != "NotSet" {
			t.Error("The level has not been set.")
		}
	})
	t.Run("DefaultLevelIsZero", func(t *testing.T) {
		if simpleLevel1.(*simpleLevel).Level() != 0 {
			t.Error("The level has not been set.")
		}
	})
}

func TestInValidSimpleLevel(t *testing.T) {
	t.Run("New(AllLevel+1)Level", func(t *testing.T) {
		_, err := NewLevel(AllLevelValue+1, "NewLevel")
		if errorValue, ok := InvalidInputSimpleLevelErrorCode.AsError(err); !ok || errorValue.Error() != "Level can not be greater than All level (1000)" {
			t.Error("Generated an invalid level", "level", AllLevelValue+1, "err", err)
		}
	})
	t.Run("New(AllLevel)Level", func(t *testing.T) {
		_, err := NewLevel(AllLevelValue, "NewLevel")
		if errorValue, ok := InvalidInputSimpleLevelErrorCode.AsError(err); !ok || errorValue.Error() != "Level can not be greater than All level (1000)" {
			t.Error("Generated an invalid level", "level", AllLevelValue, "err", err)
		}
	})
	t.Run("New(NoLevel)Level", func(t *testing.T) {
		_, err := NewLevel(NoLevelValue, "NewLevel")
		if errorValue, ok := InvalidInputSimpleLevelErrorCode.AsError(err); !ok || errorValue.Error() != "Level can not be less than NoLevel level (-1)" {
			t.Error("Generated an invalid level", "level", NoLevelValue, "err", err)
		}
	})
	t.Run("New(NoLevel-1)Level", func(t *testing.T) {
		_, err := NewLevel(NoLevelValue-1, "NewLevel")
		if errorValue, ok := InvalidInputSimpleLevelErrorCode.AsError(err); !ok || errorValue.Error() != "Level can not be less than NoLevel level (-1)" {
			t.Error("Generated an invalid level", "level", NoLevelValue, "err", err)
		}
	})
	t.Run("ExistingLevelWithDifferentString", func(t *testing.T) {
		_, err := NewLevel(Info.Level(), "NewLevel")
		if _, ok := DuplicateSimpleLevelErrorCode.AsError(err); !ok {
			t.Error("Generated a duplicate level of Info with different name", "err", err)
		}
	})
}

func TestValidSimpleLevel(t *testing.T) {
	t.Run("New(AllLevel-1)Level", func(t *testing.T) {
		_, err := NewLevel(AllLevelValue-1, "NewLevel")
		if err != nil {
			t.Error("Failed to generate a valid level", "level", AllLevelValue-1, "err", err)
		}
	})
	t.Run("New(NoLevelValue+1)Level", func(t *testing.T) {
		_, err := NewLevel(NoLevelValue+1, "NewLevel")
		if err != nil {
			t.Error("Failed to generate a valid level", "level", NoLevelValue+1, "err", err)
		}
	})
	t.Run("ExistingLevelWithMatchingString", func(t *testing.T) {
		_, err := NewLevel(Info.Level(), "Info")
		if err != nil {
			t.Error("Failed to generate a valid level", "level", NoLevelValue+1, "err", err)
		}
	})
}

type newLevelType struct {
	level int
}

func (levelType *newLevelType) Level() int {
	return levelType.level
}

func (levelType *newLevelType) String() string {
	return "AlwaysSame"
}

func (levelType *newLevelType) Compare(level Level) CompareResult {
	return NotApplicable
}

type uncomparableLevel struct {
}

func (levelType *uncomparableLevel) String() string {
	return "AlwaysSame"
}

func (levelType *uncomparableLevel) Compare(level Level) CompareResult {
	return NotApplicable
}

func TestLevelCompare(t *testing.T) {
	t.Run("CompareWithUnsetLevel", func(t *testing.T) {
		if Info.Compare(&simpleLevel{}) != NotApplicable {
			t.Error("Comparison with default objects should fail with NotApplicable")
		}
	})
	t.Run("CompareOfUnsetLevel", func(t *testing.T) {
		var uninitializedSimpleLevel simpleLevel = simpleLevel{}
		if uninitializedSimpleLevel.Compare(Info) != NotApplicable {
			t.Error("Comparison of valid Level with uninitialized object should fail with NotApplicable")
		}
	})
	t.Run("CompareWithDifferentImplementationOfLevel", func(t *testing.T) {
		var newValidLevel = &newLevelType{level: Info.level}
		if Info.Compare(newValidLevel) != Equal {
			t.Error("Comparison of alternate implementation with matching level should be successful")
		}
	})
	t.Run("CompareWithALevelWithMissingLevelFunction", func(t *testing.T) {
		var uncomparable Level = &uncomparableLevel{}
		if value := Info.Compare(uncomparable); value != NotApplicable {
			t.Error("Comparison of alternate implementation with no Level() method should return NotApplicable", "returnedValue", value)
		}
	})
	t.Run("CompareWithALevelWithLowerValue", func(t *testing.T) {
		var newLevel Level = newLevelNoError(Info.level-1, "newLevel")
		if value := Info.Compare(newLevel); value != Greater {
			t.Error("Comparison of level with lower level value should return Greater", "returnedValue", value)
		}
	})
	t.Run("CompareWithALevelWithHigherValue", func(t *testing.T) {
		var newLevel Level = newLevelNoError(Info.level+1, "newLevel")
		if value := Info.Compare(newLevel); value != Less {
			t.Error("Comparison of level with lower level value should return Less", "returnedValue", value)
		}
	})
	t.Run("CompareWithALevelWithEqualValue", func(t *testing.T) {
		var newLevel Level = newLevelNoError(Info.level, "Info")
		if value := Info.Compare(newLevel); value != Equal {
			t.Error("Comparison of level with equal level and string should return Equal", "returnedValue", value)
		}
	})
	t.Run("newLevelNoErrorInvalidData", func(t *testing.T) {
		var newLevel *simpleLevel = newLevelNoError(NoLevelValue-1, "LessThanNoLevelValue")
		if newLevel != nil {
			t.Error("No level less than NoLevelValue can be created.", "newLevel", newLevel)
		}
	})
}
