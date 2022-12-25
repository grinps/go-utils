package gosub

import (
	"fmt"
	logger "github.com/grinps/go-utils/base-utils/logs"
	"time"
)

type timeSelection struct {
	source           *time.Ticker
	duration         time.Duration
	onSelectFunction OnSelect
	stringRep        string
}

func (selection *timeSelection) String() string         { return selection.stringRep }
func (selection *timeSelection) GetSource() interface{} { return selection.source }
func (selection *timeSelection) GetOnSelect() OnSelect  { return selection.onSelectFunction }
func (selection *timeSelection) ConnectToChannel(channel chan interface{}) (InitiateChannelMonitor, error) {
	proxyChannel := NewProxyChannel(selection.source.C, channel)
	return func(identifier SelectorIdentifier) error {
		return proxyChannel.Begin(identifier)
	}, nil
}

func (selection *timeSelection) Execute(selectorEventType SelectorEvent, parameters ...interface{}) (returnErr error) {
	switch selectorEventType {
	case SelectorEventStop:
		logger.Log("Stopping ticker", selection)
		selection.source.Stop()
	case SelectorEventReset:
		logger.Log("Resetting ticker", selection)
		selection.source.Reset(selection.duration)
	default:
		returnErr = ErrSelectionConfigExecute.NewF(ErrSelectionConfigExecuteParamEventType, selectorEventType,
			ErrSelectionConfigExecuteParamEventParams, parameters,
			ErrSelectionConfigExecuteParamReason, ErrSelectionConfigExecuteReasonNotSupported)
	}
	return
}

func WithTick(timeDuration time.Duration, onSelect OnSelect) Selectable {
	return SelectorFunction(func(collection SelectCollection) error {
		var returnError error = nil
		_, returnError = collection.Register(&timeSelection{
			source:           time.NewTicker(timeDuration),
			duration:         timeDuration,
			onSelectFunction: onSelect,
			stringRep:        fmt.Sprintf("Timer-%s", timeDuration),
		})
		return returnError
	})
}
