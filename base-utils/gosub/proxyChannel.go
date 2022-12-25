package gosub

import (
	logger "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
)

func NewProxyChannel[T any](sender <-chan T, receiver chan interface{}) *ProxyChannel[T] {
	return &ProxyChannel[T]{sender: sender, receiver: receiver}
}

func NewProxyChannelWithSource[T any](sender <-chan T, receiver chan interface{}, source SelectorIdentifier) *ProxyChannel[T] {
	return &ProxyChannel[T]{sender: sender, receiver: receiver, source: source}
}

type ProxyChannel[T any] struct {
	sender   <-chan T
	receiver chan interface{}
	source   SelectorIdentifier
}

var ErrProxyChannelBeginError = errext.NewErrorCodeWithOptions(errext.WithTemplate("Failed to begin proxy channel with sender",
	"["+ErrProxyChannelParamSender+"]", "and receiver", "["+ErrProxyChannelParamReceiver+"]", "due to", "["+ErrProxyChannelParamReason+"]"))

const (
	ErrProxyChannelParamSender           = "Sender"
	ErrProxyChannelParamReceiver         = "Receiver"
	ErrProxyChannelParamReason           = "reason"
	ErrProxyChannelParamReasonNilChannel = "proxy channel is nil"
)

func (channel *ProxyChannel[T]) Begin(identifier SelectorIdentifier) error {
	if channel == nil {
		return ErrProxyChannelBeginError.NewF(ErrProxyChannelParamSender, nil, ErrProxyChannelParamReceiver, nil, ErrProxyChannelParamReason, ErrProxyChannelParamReasonNilChannel)
	}
	applicableIdentifier := channel.source
	if identifier != SelectorIdentifierInvalid {
		applicableIdentifier = identifier
	}
	go func() {
		defer func() {
			if recoverPanic := recover(); recoverPanic != nil {
				logger.Log("Proxy channel go routine panic'ed for sender", channel.sender, "with identifier", channel.source, "error", recoverPanic)
				panic(recoverPanic)
			}
		}()
		logger.Log("Starting proxy channel go routine from", channel.sender, "to", channel.receiver, "for source", applicableIdentifier)
		breakout := false
		for {
			select {
			case itemReceived, ok := <-channel.sender:
				if ok {
					logger.Log("Received item", itemReceived, "for source", applicableIdentifier)
					channel.receiver <- SelectEvent{
						Source:        applicableIdentifier,
						ReceiveStatus: ok,
						Received:      itemReceived,
					}
				} else {
					logger.Log("Received ok as false so breaking out", applicableIdentifier)
					breakout = true
				}
			}
			if breakout {
				logger.Log("Exiting for-select loop for source", applicableIdentifier)
				break
			}
		}
	}()
	logger.Log("Completed setup proxy channel from", channel.sender, "to", channel.receiver)
	return nil
}
