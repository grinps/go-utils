package gosub

import (
	"fmt"
	"github.com/grinps/go-utils/errext"
)

type SelectorIdentifier string

const (
	SelectorIdentifierInvalid SelectorIdentifier = "SelectorIdentifierInvalid"
)

func (id SelectorIdentifier) Equals(inputId SelectorIdentifier) bool {
	return string(id) == string(inputId)
}

type SelectEvent struct {
	Source        SelectorIdentifier
	ReceiveStatus bool
	Received      interface{}
}

type OnSelect func(event SelectEvent, collection SelectCollection) (continueSelecting bool)

func ExitOnSelect(event SelectEvent, collection SelectCollection) (continueSelecting bool) {
	return false
}

type SelectorEvent int

func (eventType SelectorEvent) Equals(givenEventType SelectorEvent) bool {
	return int(eventType) == int(givenEventType)
}

const (
	SelectorEventUnknown SelectorEvent = iota
	SelectorEventStop                  = SelectorEvent(1)
	SelectorEventReset                 = SelectorEvent(2)
)

var ErrSelectionConfigConnectToChannel = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to connect to channel",
	"["+ErrSelectionConfigConnectToChannelParamChannel+"]", "due to error",
	"["+ErrSelectionConfigConnectToChannelParamReason+"]"))

const (
	ErrSelectionConfigConnectToChannelParamChannel         = "channel"
	ErrSelectionConfigConnectToChannelParamReason          = "reason"
	ErrSelectionConfigConnectToChannelReasonInvalidChannel = "given channel is invalid"
)

var ErrSelectionConfigExecute = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to execute event type",
	"["+ErrSelectionConfigExecuteParamEventType+"]",
	"with parameters", "["+ErrSelectionConfigExecuteParamEventParams+"]",
	"due to error", "["+ErrSelectionConfigExecuteParamReason+"]"))

const (
	ErrSelectionConfigExecuteParamEventType          = "eventType"
	ErrSelectionConfigExecuteParamEventParams        = "eventParameters"
	ErrSelectionConfigExecuteParamReason             = "reason"
	ErrSelectionConfigExecuteReasonNotSupported      = "given event is not supported"
	ErrSelectionConfigExecuteReasonInvalidParameters = "given parameters are invalid"
	ErrSelectionConfigExecuteReasonExecutionFailed   = "failed to execute. See error"
)

type SelectionConfig interface {
	fmt.Stringer
	GetSource() interface{}
	GetOnSelect() OnSelect
	ConnectToChannel(channel chan interface{}) (InitiateChannelMonitor, error)
	Execute(selectorEventType SelectorEvent, parameters ...interface{}) error
}
type InitiateChannelMonitor func(identifier SelectorIdentifier) error

var ErrRegistrationFailed = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to register selector",
	"["+ErrRegistrationFailedParamSelectable+"]", "due to error", "["+ErrSelectCollectionParamReason+"]"))

const (
	ErrSelectCollectionParamCollection       = "collection"
	ErrRegistrationFailedParamSelectable     = "selectable"
	ErrSelectCollectionParamReason           = "reason"
	ErrRegistrationFailedReasonNilSelectable = "selectable specified is nil"
	ErrRegistrationFailedReasonNilCollection = "calling method on nil selection"
)

var ErrInitializationFailed = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to initialize select collection",
	"["+ErrSelectCollectionParamCollection+"]", "due to error", "["+ErrSelectCollectionParamReason+"]"))

type SelectCollection interface {
	Register(selectable SelectionConfig) (SelectorIdentifier, error)
	Initialize() error
	GetSelector(identifier SelectorIdentifier) (selectable SelectionConfig, returnErr error)
	Select()
	Shutdown() error
}

type Selectable interface {
	Selector(collection SelectCollection) error
}

type SelectorFunction func(collection SelectCollection) error

func (function SelectorFunction) Selector(collection SelectCollection) error {
	return function(collection)
}
