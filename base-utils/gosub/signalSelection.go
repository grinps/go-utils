package gosub

import (
	"fmt"
	"github.com/grinps/go-utils/errext"
	"os"
	"os/signal"
)

type signalSelection struct {
	source           chan os.Signal
	signals          string
	onSelectFunction OnSelect
}

func (selection *signalSelection) String() string         { return "Signal" + selection.signals }
func (selection *signalSelection) GetSource() interface{} { return selection.source }
func (selection *signalSelection) GetOnSelect() OnSelect  { return selection.onSelectFunction }
func (selection *signalSelection) ConnectToChannel(channel chan interface{}) (InitiateChannelMonitor, error) {
	proxyChannel := NewProxyChannel(selection.source, channel)
	return func(identifier SelectorIdentifier) error {
		return proxyChannel.Begin(identifier)
	}, nil
}

func (selection *signalSelection) Execute(selectorEventType SelectorEvent, parameters ...interface{}) (returnErr error) {
	switch selectorEventType {
	case SelectorEventStop:
		defer func() {
			recoverValue := recover()
			if recoverValue != nil {
				if asErr, isErr := recoverValue.(error); isErr {
					returnErr = ErrSelectionConfigExecute.NewWithErrorF(asErr, ErrSelectionConfigExecuteParamEventType, selectorEventType,
						ErrSelectionConfigExecuteParamEventParams, parameters,
						ErrSelectionConfigExecuteParamReason, ErrSelectionConfigExecuteReasonExecutionFailed, errext.NewField("recoveredValue", recoverValue))
				} else {
					returnErr = ErrSelectionConfigExecute.NewF(recoverValue, ErrSelectionConfigExecuteParamEventType, selectorEventType,
						ErrSelectionConfigExecuteParamEventParams, parameters,
						ErrSelectionConfigExecuteParamReason, ErrSelectionConfigExecuteReasonNotSupported, errext.NewField("recoveredValue", recoverValue))
				}
			}
		}()
		close(selection.source)
	default:
		returnErr = ErrSelectionConfigExecute.NewF(ErrSelectionConfigExecuteParamEventType, selectorEventType,
			ErrSelectionConfigExecuteParamEventParams, parameters,
			ErrSelectionConfigExecuteParamReason, ErrSelectionConfigExecuteReasonNotSupported)
	}
	return
}

func WithSignals(onSelectFunction OnSelect, signals ...os.Signal) Selectable {
	sigChannel := make(chan os.Signal)
	signal.Notify(sigChannel, signals...)
	signalAsString := fmt.Sprintf("%s", signals)
	return SelectorFunction(func(collection SelectCollection) error {
		var returnError error = nil
		_, returnError = collection.Register(&signalSelection{
			source:           sigChannel,
			signals:          signalAsString,
			onSelectFunction: onSelectFunction,
		})
		return returnError
	})
}
