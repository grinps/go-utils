package base_utils

// CompareResult defines the results for a compare operation (see Comparable)
type CompareResult int

const (
	Less          CompareResult = -1
	Equals        CompareResult = 0
	Greater       CompareResult = 1
	NotApplicable CompareResult = 10
)

// Equality interface provides ability to compare two object of same type.
// The implementation should try to observe the following
//
// 1. if both source and target objects are nil return true
//
// 2. if both are not nil, and source.Equals(target.(sourcetype)) return true
//
// 3. avoid comparing pointers since that may result in panics
//TODO: Update to support generics when available
type Equality interface {
	Equals(targetObject Equality) bool
}

// Comparable interface defines Compare function to compare source implementation with target. It returns Less, Equals,
// and Greater based on comparison result, NotApplicable if the comparison can not be performed.
//TODO: Update to support generics when available
type Comparable interface {
	Compare(targetObject Comparable) CompareResult
}
