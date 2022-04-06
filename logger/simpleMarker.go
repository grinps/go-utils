package logger

import (
	"fmt"
	"strings"
)

const (
	SimpleMarkerFormat              = "%s.%s"
	SimpleMarkerMethodFormat        = "%s.%s(%s)"
	SimpleMarkerLoggerNameDelimiter = "."
)

var DefaultSimpleMarker = NewSimpleMarker("Default")

type simpleMarker struct {
	marker    string
	aspects   []string
	subMarker Marker
}

func (marker *simpleMarker) String() string {
	return marker.marker
}

func (marker *simpleMarker) AddMethod(callerMarker Marker, methodName string, values ...interface{}) Marker {
	var returnMarker = marker.Copy()
	returnMarker.subMarker = callerMarker
	var valuesAsString = ""
	if values != nil {
		valuesAsString = fmt.Sprintln(values...)
		valuesAsString = strings.TrimRight(valuesAsString, "\r\n")
	}
	returnMarker.marker = fmt.Sprintf(SimpleMarkerMethodFormat, marker.marker, methodName, valuesAsString)
	return returnMarker
}

func (marker *simpleMarker) Append(values ...string) Marker {
	var valuesAsString = ""
	if values != nil {
		valuesAsString = strings.Join(values, SimpleMarkerLoggerNameDelimiter)
	}
	marker.aspects = append(marker.aspects, values...)
	if valuesAsString != "" {
		marker.marker = fmt.Sprintf(SimpleMarkerFormat, marker.marker, valuesAsString)
	}
	return marker
}

func (marker *simpleMarker) Add(aspects ...string) Marker {
	var returnMarker = marker.Copy()
	returnMarker.Append(aspects...)
	return returnMarker
}

func (marker *simpleMarker) GetAspects() []string {
	return marker.aspects
}

func (marker *simpleMarker) Push(subMarker Marker) Marker {
	var returnMarker = marker.Copy()
	returnMarker.subMarker = subMarker
	return returnMarker
}

func (marker *simpleMarker) Copy() *simpleMarker {
	return &simpleMarker{
		marker:    marker.marker,
		subMarker: marker.subMarker,
		aspects:   marker.aspects,
	}
}

func (marker *simpleMarker) Pop() Marker {
	return marker.subMarker
}

func NewSimpleMarker(marker string) *simpleMarker {
	applicableMarkerName := marker
	aspects := strings.Split(marker, SimpleMarkerLoggerNameDelimiter)
	return &simpleMarker{marker: applicableMarkerName, aspects: aspects}
}
