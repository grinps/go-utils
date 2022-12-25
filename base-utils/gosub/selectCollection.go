package gosub

import (
	logger "github.com/grinps/go-utils/base-utils/logs"
	"github.com/grinps/go-utils/errext"
	"reflect"
	"strconv"
	"sync"
)

type selectCollectionImpl struct {
	collection       map[SelectorIdentifier]SelectionConfig
	collectorChannel chan interface{}
	lastAssigned     int8
	initialized      bool
	collectionMutex  *sync.Mutex
	shutdownChannel  chan interface{}
}

func (selectCollection *selectCollectionImpl) Register(selectionConfig SelectionConfig) (SelectorIdentifier, error) {
	if selectionConfig == nil {
		return SelectorIdentifierInvalid, ErrRegistrationFailed.NewF(ErrRegistrationFailedParamSelectable, selectionConfig, ErrSelectCollectionParamReason, ErrRegistrationFailedReasonNilSelectable)
	}
	if selectCollection == nil {
		return SelectorIdentifierInvalid, ErrRegistrationFailed.NewF(ErrRegistrationFailedParamSelectable, selectionConfig, ErrSelectCollectionParamReason, ErrRegistrationFailedReasonNilCollection)
	}
	selectCollection.collectionMutex.Lock()
	defer selectCollection.collectionMutex.Unlock()
	selectCollection.lastAssigned++
	returnIdentifier := SelectorIdentifier(selectionConfig.String() + "-" + strconv.Itoa(int(selectCollection.lastAssigned)))
	selectCollection.collection[returnIdentifier] = selectionConfig
	logger.Log("Registered", selectionConfig, "with identifier", returnIdentifier, "to", selectCollection)
	return returnIdentifier, nil
}

const (
	ErrInitializationFailedParamSelectorID             = "SelectorIdentifier"
	ErrInitializationFailedReasonMissingChannelMonitor = "select config returned nil channel monitor"
	ErrInitializationFailedReasonChannelConnectFailed  = "failed to connect receiver with sender channel from source config"
	ErrInitializationFailedReasonChannelStartFailed    = "failed to start sender-receiver connection for source config"
)

func (selectCollection *selectCollectionImpl) Initialize() (returnErr error) {
	logger.Log("Initializing", selectCollection)
	if selectCollection == nil {
		returnErr = ErrInitializationFailed.NewF(ErrSelectCollectionParamCollection, selectCollection,
			ErrSelectCollectionParamReason, ErrRegistrationFailedReasonNilCollection)
		return
	}
	selectCollection.collectionMutex.Lock()
	defer selectCollection.collectionMutex.Unlock()
	selectCollection.collectorChannel = make(chan interface{})
	for selectConfigIndex, selectConfig := range selectCollection.collection {
		logger.Log("Starting connection to channel for", selectConfigIndex)
		if channelMonitor, errChannelMonitor := selectConfig.ConnectToChannel(selectCollection.collectorChannel); errChannelMonitor != nil {
			// TODO: support multi-error
			returnErr = ErrInitializationFailed.NewWithErrorF(errChannelMonitor,
				ErrSelectCollectionParamCollection, selectCollection,
				ErrSelectCollectionParamReason, ErrInitializationFailedReasonChannelConnectFailed,
				errext.NewField(ErrInitializationFailedParamSelectorID, selectConfigIndex))
			break
		} else if channelMonitor == nil {
			returnErr = ErrInitializationFailed.NewF(ErrSelectCollectionParamCollection, selectCollection,
				ErrSelectCollectionParamReason, ErrInitializationFailedReasonMissingChannelMonitor,
				errext.NewField(ErrInitializationFailedParamSelectorID, selectConfigIndex))
			break
		} else if errStartingChannelMonitor := channelMonitor(selectConfigIndex); errStartingChannelMonitor != nil {
			// TODO: support multi-error
			returnErr = ErrInitializationFailed.NewWithErrorF(errStartingChannelMonitor, ErrSelectCollectionParamCollection, selectCollection,
				ErrSelectCollectionParamReason, ErrInitializationFailedReasonChannelStartFailed,
				errext.NewField(ErrInitializationFailedParamSelectorID, selectConfigIndex))
			break
		} else {
			logger.Log("Started channel monitor for", selectConfigIndex)
		}
	}
	selectCollection.initialized = true
	logger.Log("Initialized", selectCollection, "with error", returnErr)
	return
}

func (selectCollection *selectCollectionImpl) Select() {
	if selectCollection == nil {
		return
	}
	if selectCollection.initialized {
		logger.Log("Begin select on", selectCollection)
		for {
			continueSelecting := true
			logger.Log("Waiting on select", selectCollection)
			select {
			case returnedValue := <-selectCollection.collectorChannel:
				logger.Log("Received from collection channel for", selectCollection, "as", returnedValue)
				if asEvent, isEvent := returnedValue.(SelectEvent); isEvent && asEvent.ReceiveStatus {
					if applicableSelector, getSelectorErr := selectCollection.GetSelector(asEvent.Source); getSelectorErr != nil || applicableSelector == nil {
						logger.Log("Failed to retrieve applicable selector for", asEvent.Source, "from collection", selectCollection, "error", getSelectorErr)
					} else {
						logger.Log("Executing on select")
						continueSelecting = applicableSelector.GetOnSelect()(asEvent, selectCollection)
						logger.Log("Executed on select with result", continueSelecting)
					}
				} else {
					logger.Log("Received from collection channel", selectCollection, "event", returnedValue, "of", reflect.TypeOf(returnedValue), "and not a SelectEvent")
				}
			case <-selectCollection.shutdownChannel:
				logger.Log("Received shutdown event")
				continueSelecting = false
			}
			logger.Log("Completed waiting on select", selectCollection)
			if !continueSelecting {
				break
			}
		}
		logger.Log("Ended select on", selectCollection)
	} else {
		logger.Log("No select on", selectCollection, "since it is not initialized")
	}
}

func (selectCollection *selectCollectionImpl) GetSelector(identifier SelectorIdentifier) (selectable SelectionConfig, returnErr error) {
	return selectCollection.collection[identifier], nil
}

func (selectCollection *selectCollectionImpl) Shutdown() (returnErr error) {
	logger.Log("Shutting down selectCollection", selectCollection)
	for selectConfigIndex, selectConfig := range selectCollection.collection {
		logger.Log("Shutting down selection config", selectConfigIndex)
		closingError := selectConfig.Execute(SelectorEventStop)
		if closingError != nil {
			logger.Warn("Failed to shutdown selection config", selectConfigIndex, "due to error", closingError)
		}
	}
	logger.Log("Sending message to shutdown channel")
	selectCollection.shutdownChannel <- struct{}{}
	logger.Log("Shutdown selectCollection", selectCollection, "with error", returnErr)
	return
}

func NewSelectCollection(selectables ...Selectable) SelectCollection {
	collection, _ := NewSelectCollectionE(selectables...)
	return collection
}

func NewSelectCollectionP(selectables ...Selectable) SelectCollection {
	collection, err := NewSelectCollectionE(selectables...)
	if err != nil {
		panic(err)
	}
	return collection
}

func NewSelectCollectionE(selectables ...Selectable) (SelectCollection, error) {
	returnSelectCollection := &selectCollectionImpl{
		collection:       map[SelectorIdentifier]SelectionConfig{},
		collectorChannel: make(chan interface{}),
		lastAssigned:     0,
		collectionMutex:  &sync.Mutex{},
		shutdownChannel:  make(chan interface{}, 100),
	}
	var returnError error = nil
	for _, selectable := range selectables {
		//TODO: Multi error
		returnError = selectable.Selector(returnSelectCollection)
		if returnError != nil {
			logger.Warn("Failed to add selectable", selectable, "to collection", returnSelectCollection, "due to error", returnError)
			break
		}
	}
	return returnSelectCollection, returnError
}
