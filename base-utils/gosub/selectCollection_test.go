package gosub

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/grinps/go-utils/errext"
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestNewSelectCollection(t *testing.T) {
	t.Run("ValidTicker", func(t *testing.T) {
		collectBuffer := &bytes.Buffer{}
		var counter int = 0
		collection, err := NewSelectCollectionE(WithTick(50*time.Millisecond, func(selectEvent SelectEvent, collection SelectCollection) (continueSelecting bool) {
			collectBuffer.Write([]byte(fmt.Sprintf("%t-%s", selectEvent.ReceiveStatus, selectEvent.Source)))
			if counter < 1 {
				counter = counter + 1
				return true
			} else {
				return false
			}
		}))
		if err != nil {
			t.Errorf("Expected no err during creating new selection, Actual %#v", err)
		} else {
			initErr := collection.Initialize()
			if initErr != nil {
				t.Errorf("Expected no err during initializing new selection, Actual %#v", initErr)
			} else {
				collection.Select()
				shutErr := collection.Shutdown()
				if shutErr != nil {
					t.Errorf("Expected no err during shutdown of selection, Actual %#v", shutErr)
				}
				expectedValue := "true-Timer-50ms-1true-Timer-50ms-1"
				if collectBuffer.String() != expectedValue {
					t.Errorf("Expected output %s, Actual %s", expectedValue, collectBuffer.String())
				}

			}
		}
	})
	t.Run("Empty", func(t *testing.T) {
		collection, err := NewSelectCollectionE()
		if err != nil {
			t.Errorf("Expected no error while creating collection %#v", err)
		} else if collection == nil {
			t.Errorf("Expected collection to be not nil actual nil")
		} else if err := collection.Initialize(); err != nil {
			t.Errorf("Expected initialization to be success, actual %#v", err)
		} else {
			go func() {
				time.Sleep(100 * time.Millisecond)
				shutErr := collection.Shutdown()
				if shutErr != nil {
					t.Errorf("Expected initialization to be success, actual %#v", err)
				}
			}()
			collection.Select()
		}
	})
	t.Run("MultipleTickers", func(t *testing.T) {
		collector := bytes.Buffer{}
		counter := 0
		var aFunc = func(event SelectEvent, collection SelectCollection) bool {
			returnValue := true
			if counter%2 == 0 {
				collector.Write([]byte(fmt.Sprintf("SourceEven-%s", event.Source)))
			} else {
				collector.Write([]byte(fmt.Sprintf("SourceOdd-%s", event.Source)))
			}
			counter++
			if counter >= 4 {
				returnValue = false
			}
			return returnValue
		}
		if collection, err := NewSelectCollectionE(WithTick(50*time.Millisecond, aFunc),
			WithTick(20*time.Millisecond, aFunc)); err != nil {
			t.Errorf("Expected no error, actual %#v", err)
		} else if collection == nil {
			t.Errorf("Expected valid collection")
		} else if err = collection.Initialize(); err != nil {
			t.Errorf("No error expected during initialize %#v", err)
		} else {
			go func() {
				time.Sleep(120 * time.Millisecond)
				shutErr := collection.Shutdown()
				if shutErr != nil {
					t.Errorf("Expected initialization to be success, actual %#v", err)
				}
			}()
			collection.Select()
			output := collector.String()
			expectedOutput := "SourceEven-Timer-20ms-2SourceOdd-Timer-20ms-2SourceEven-Timer-50ms-1SourceOdd-Timer-20ms-2"
			if output != expectedOutput {
				t.Errorf("Expected output %s, Actual %s", expectedOutput, output)
			}
		}
	})
}

func TestSelectCollectionImpl_Register(t *testing.T) {
	t.Run("NilSelectConfig", func(t *testing.T) {
		var selectCollection *selectCollectionImpl = &selectCollectionImpl{}
		selectCollection = nil
		identifier, registerErr := selectCollection.Register(nil)
		if registerErr == nil {
			t.Errorf("Expected error, Actual no error")
		} else if identifier != SelectorIdentifierInvalid {
			t.Errorf("Expected %s identifier actual %s", SelectorIdentifierInvalid, identifier)
		}
		if _, isErr := ErrRegistrationFailed.AsError(registerErr); !isErr {
			t.Errorf("Expected ErrRegistrationFailed, Actual %#v", registerErr)
		}
	})
	t.Run("NilObject", func(t *testing.T) {
		var selectCollection *selectCollectionImpl = nil
		identifier, registerErr := selectCollection.Register(&timeSelection{})
		if registerErr == nil {
			t.Errorf("Expected error, Actual no error")
		} else if identifier != SelectorIdentifierInvalid {
			t.Errorf("Expected %s identifier actual %s", SelectorIdentifierInvalid, identifier)
		}
		if _, isErr := ErrRegistrationFailed.AsError(registerErr); !isErr {
			t.Errorf("Expected ErrRegistrationFailed, Actual %#v", registerErr)
		}
	})
}

type dummySelectConfig struct {
	stringValue    string
	selectId       SelectorIdentifier
	source         interface{}
	onSelect       OnSelect
	monitorChannel InitiateChannelMonitor
	connect        func(selectId SelectorIdentifier, channel chan interface{})
	execute        func(selectId SelectorIdentifier, selectorEventType SelectorEvent, parameters ...interface{}) error
	connectError   error
	executeErr     error
}

func (config *dummySelectConfig) String() string {
	return config.stringValue
}

func (config *dummySelectConfig) GetSource() interface{} {
	return config.source
}
func (config *dummySelectConfig) GetOnSelect() OnSelect {
	return config.onSelect
}
func (config *dummySelectConfig) ConnectToChannel(channel chan interface{}) (InitiateChannelMonitor, error) {
	if config.connect != nil {
		config.connect(config.selectId, channel)
	}
	return config.monitorChannel, config.connectError
}
func (config *dummySelectConfig) Execute(selectorEventType SelectorEvent, parameters ...interface{}) error {
	if config.execute != nil {
		return config.execute(config.selectId, selectorEventType, parameters...)
	}
	return config.executeErr
}

func TestSelectCollectionImpl_Initialize(t *testing.T) {
	var nilCol *selectCollectionImpl = nil
	runTest(t, []testCaseDef{
		{
			name: "NilObject", collection: nilCol,
			initialize: true, initializeSuccessExpected: false, initializeErrCode: ErrInitializationFailed,
		},
		{
			name: "NilCollectionDummyDefaultSelectConfig", collection: nilCol,
			register: true, registerValue: &dummySelectConfig{}, registerSuccessExpected: false, registerErrCode: ErrRegistrationFailed,
			initialize: true, initializeSuccessExpected: false, initializeErrCode: ErrInitializationFailed,
		},
		{
			name: "DummyDefaultSelectConfig", collection: NewSelectCollection(),
			register: true, registerValue: &dummySelectConfig{}, registerSuccessExpected: true, registerErrCode: nil,
			initialize: true, initializeSuccessExpected: false, initializeErrCode: ErrInitializationFailed,
		},
		{
			name: "DummyDefaultSelectConfigMonitorChannelError", collection: NewSelectCollection(),
			register: true, registerValue: &dummySelectConfig{
				stringValue:    "DummyDefaultSelectConfigMonitorChannelError",
				monitorChannel: func(identifier SelectorIdentifier) error { return errors.New("Ageneric error") },
			}, registerSuccessExpected: true, registerErrCode: nil,
			initialize: true, initializeSuccessExpected: false, initializeErrCode: ErrInitializationFailed,
		},
		{
			name: "DummyDefaultSelectConfigMonitorChannelBeginError", collection: NewSelectCollection(),
			register: true, registerValue: &dummySelectConfig{
				stringValue:    "DummyDefaultSelectConfigMonitorChannelBeginError",
				monitorChannel: func(identifier SelectorIdentifier) error { return nil },
				connectError:   errors.New("connect error"),
			}, registerSuccessExpected: true, registerErrCode: nil,
			initialize: true, initializeSuccessExpected: false, initializeErrCode: ErrInitializationFailed,
			selectOnCol:                 false,
			shutdownBeforeSelectInGoSub: false, shutdownWaitInGoSub: 0, shutdown: false, shutdownSuccessExpected: false,
			shutdownErrCode: nil,
		},
	})
}

func TestSelectCollectionImpl_Select(t *testing.T) {
	t.Run("NilObject", func(t *testing.T) {
		defer func() {
			if recover() != nil {
				t.Errorf("Expected no error, actual %#v", recover())
			}
		}()
		var selectCollection *selectCollectionImpl = nil
		selectCollection.Select()
	})
	SelectAfterDeleteColl := NewSelectCollection()
	runTest(t, []testCaseDef{
		{
			name: "SelectValid", collection: NewSelectCollection(), register: true,
			registerValue: &dummySelectConfig{
				stringValue: "SelectValid", source: struct{}{},
				onSelect: OnSelectOnce, monitorChannel: PrintChannelIdOnMonitor, connectError: nil, executeErr: nil,
			},
			registerSuccessExpected: true, registerErrCode: nil, initialize: true, initializeSuccessExpected: true,
			initializeErrCode: nil, selectOnCol: true, shutdownBeforeSelectInGoSub: true, shutdownWaitInGoSub: 100 * time.Millisecond,
			shutdown: false, shutdownSuccessExpected: true, shutdownErrCode: nil,
		},
		{
			name: "SelectConnectWithNilValue", collection: NewSelectCollection(), register: true,
			registerValue: &dummySelectConfig{
				stringValue: "SelectConnectWithNilValue",
				source:      struct{}{}, onSelect: ExitOnSelect, monitorChannel: PrintChannelIdOnMonitor,
				connectError: nil, executeErr: nil, execute: nil, connect: PushValuesInGoSub(nil, nil),
			},
			registerSuccessExpected: true, registerErrCode: nil, initialize: true, initializeSuccessExpected: true,
			initializeErrCode: nil, selectOnCol: true, shutdownBeforeSelectInGoSub: true, shutdownWaitInGoSub: 100 * time.Millisecond,
			shutdown: false, shutdownSuccessExpected: true, shutdownErrCode: nil,
		},
		{
			name: "SelectOnUnInitialized", collection: NewSelectCollection(), register: true,
			registerValue: &dummySelectConfig{
				stringValue: "SelectOnUnInitialized",
				source:      struct{}{}, onSelect: OnSelectOnce,
				monitorChannel: PrintChannelIdOnMonitor,
				connectError:   nil, executeErr: nil, execute: nil, connect: PushValuesInGoSub(nil, nil),
			},
			registerSuccessExpected: true, registerErrCode: nil, initialize: false, initializeSuccessExpected: true,
			initializeErrCode: nil, selectOnCol: true, shutdownBeforeSelectInGoSub: true, shutdownWaitInGoSub: 1 * time.Second,
			shutdown: false, shutdownSuccessExpected: true, shutdownErrCode: nil,
		},
		{
			name: "SelectAfterDelete", collection: SelectAfterDeleteColl, register: true,
			registerValue: &dummySelectConfig{
				stringValue: "SelectAfterDelete",
				source:      struct{}{}, onSelect: ExitOnSelect,
				monitorChannel: func(identifier SelectorIdentifier) error {
					fmt.Println("Deleting identifier", identifier)
					delete(SelectAfterDeleteColl.(*selectCollectionImpl).collection, identifier)
					return nil
				},
				connectError: nil, executeErr: nil, execute: nil, connect: PushValuesInGoSub(SelectEvent{ReceiveStatus: true}, SelectEvent{ReceiveStatus: true}),
			},
			registerSuccessExpected: true, registerErrCode: nil, initialize: true, initializeSuccessExpected: true,
			initializeErrCode: nil, selectOnCol: true, shutdownBeforeSelectInGoSub: true, shutdownWaitInGoSub: 1 * time.Second,
			shutdown: false, shutdownSuccessExpected: true, shutdownErrCode: nil,
		},
	})
}

func TestSelectCollectionImpl_Shutdown(t *testing.T) {
	runTest(t, []testCaseDef{
		{
			name: "ShutdownCloseError", collection: NewSelectCollection(), register: true,
			registerValue: &dummySelectConfig{
				stringValue: "SelectValid", source: struct{}{},
				onSelect: ExitOnSelect, monitorChannel: PrintChannelIdOnMonitor, connectError: nil, executeErr: errors.New("GenericCloseError"),
			},
			registerSuccessExpected: true, registerErrCode: nil, initialize: true, initializeSuccessExpected: true,
			initializeErrCode: nil, selectOnCol: true, shutdownBeforeSelectInGoSub: true, shutdownWaitInGoSub: 100 * time.Millisecond,
			shutdown: false, shutdownSuccessExpected: true, shutdownErrCode: nil,
		},
	})
}
func TestNewSelectCollectionE(t *testing.T) {
	newColl, err := NewSelectCollectionE(SelectorFunction(func(collection SelectCollection) error { return errors.New("GenericErr") }))
	if newColl == nil {
		t.Errorf("Expected not nil collection, actual nil")
	}
	if err == nil {
		t.Errorf("Expected not nil error, actual nil")
	}
}

func TestNewSelectCollectionP(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		newColl := NewSelectCollectionP()
		if newColl == nil {
			t.Errorf("Expected not nil collection, actual nil")
		}
	})
	t.Run("Error", func(t *testing.T) {
		defer func() {
			if recover() == nil {
				t.Errorf("Expected panic, actual no issues.")
			}
		}()
		newColl := NewSelectCollectionP(SelectorFunction(func(collection SelectCollection) error { return errors.New("GenericErr") }))
		if newColl == nil {
			t.Errorf("Expected not nil collection, actual nil")
		}
	})
}

func TestSelectorIdentifier_Equals(t *testing.T) {
	selColl := NewSelectCollection()
	selMonitor := SelectorIdentifierInvalid
	selOnSelect := SelectorIdentifierInvalid
	var applicableSelectIdVal SelectorIdentifier = SelectorIdentifierInvalid
	var applicableSelectId *SelectorIdentifier = &applicableSelectIdVal
	GetLatestSelectId := func() SelectorIdentifier {
		return *applicableSelectId
	}
	selectId1, regErr := selColl.Register(&dummySelectConfig{
		stringValue: "SelectID1",
		connect: func(selectId SelectorIdentifier, channel chan interface{}) {
			go func() {
				channel <- SelectEvent{
					Source:        GetLatestSelectId(),
					ReceiveStatus: true,
					Received:      struct{}{},
				}
			}()
		},
		monitorChannel: func(identifier SelectorIdentifier) error {
			selMonitor = identifier
			return nil
		},
		onSelect: func(event SelectEvent, collection SelectCollection) (continueSelecting bool) {
			selOnSelect = event.Source
			ExitOnSelect(event, collection)
			return false
		},
	})
	*applicableSelectId = selectId1
	if regErr != nil {
		t.Errorf("Expected registration error nil actual %#v", regErr)
	}
	sameObject := &dummySelectConfig{stringValue: "SelectID2", monitorChannel: PrintChannelIdOnMonitor}
	sel2, _ := selColl.Register(sameObject)
	sel3, _ := selColl.Register(sameObject)
	initErr := selColl.Initialize()
	if initErr != nil {
		t.Errorf("Expected init error nil actual %#v", initErr)
	}
	go func() {
		time.Sleep(1000 * time.Millisecond)
		selColl.Shutdown()
	}()
	selColl.Select()
	if selectId1 == SelectorIdentifierInvalid {
		t.Errorf("Expected registration valid actual Invalid ")
	}
	if !selectId1.Equals(selMonitor) {
		t.Errorf("Expected match between monitoring begin %#v & received identifier %#v", selMonitor, selectId1)
	}
	if !selectId1.Equals(selOnSelect) {
		t.Errorf("Expected match between onSelect %#v & received identifier %#v", selOnSelect, selectId1)
	}
	if sel3.Equals(sel2) {
		t.Errorf("Expected no match between two registration of same object.")
	}
}

func TestSelectorEvent_Equals(t *testing.T) {
	if SelectorEventStop.Equals(SelectorEventUnknown) {
		t.Errorf("Expected two const SelectorEventStop, SelectorEventUnknown to be unequal")
	}
	select1 := SelectorEvent(int(SelectorEventStop))
	if !select1.Equals(SelectorEventStop) {
		t.Errorf("Expected two const with same int value to be equal")
	}
}

func TestNewProxyChannelWithSource(t *testing.T) {

	t.Run("NilPChannel", func(t *testing.T) {
		var pChannel *ProxyChannel[chan interface{}] = nil
		beginErr := pChannel.Begin(SelectorIdentifierInvalid)
		if beginErr == nil {
			t.Errorf("Expected begin err not nil actual nil")
		} else if _, isErr := ErrProxyChannelBeginError.AsError(beginErr); !isErr {
			t.Errorf("Expected ErrProxyChannelBeginError, Actual %#v", beginErr)
		}
	})
	t.Run("", func(t *testing.T) {
		sender := make(chan interface{})
		receiver := make(chan interface{})
		source := SelectorIdentifierInvalid
		pChannel := NewProxyChannelWithSource(sender, receiver, source)
		pChannel.Begin(SelectorIdentifierInvalid)
		close(sender)
	})
}

func TestTimeSelection_Execute(t *testing.T) {
	duration := time.Duration(100 * time.Millisecond)
	timer := time.NewTicker(duration)
	timeSelect := &timeSelection{
		source:           timer,
		duration:         duration,
		onSelectFunction: ExitOnSelect,
		stringRep:        fmt.Sprintf("Timer-%s", duration),
	}
	resetErr := timeSelect.Execute(SelectorEventReset)
	if resetErr != nil {
		t.Errorf("Expected reset error nil actual %#v", resetErr)
	}
	randomEventErr := timeSelect.Execute(SelectorEvent(5))
	if randomEventErr == nil {
		t.Errorf("Expected unsupported event error not nil actual nil")
	} else if _, isErr := ErrSelectionConfigExecute.AsError(randomEventErr); !isErr {
		t.Errorf("Expected unsupported event error OF TYPE ErrSelectionConfigExecute actual %#v", randomEventErr)
	}
	if timeSelect.GetSource() != timer {
		t.Errorf("Expected source to match Expected: %#v, Actual %#v", timer, timeSelect.GetSource())
	}
}

func TestWithSignals(t *testing.T) {
	runTest(t, []testCaseDef{
		/*{
			name: "NoSignal", collection: NewSelectCollection(WithSignals(ExitOnSelect)),
			initialize: true, initializeSuccessExpected: true, initializeErrCode: nil,
			selectOnCol: true, shutdownBeforeSelectInGoSub: true, shutdownWaitInGoSub: 100 * time.Millisecond, shutdownSuccessExpected: true,
		}, */
		{
			name: "1Signal", collection: NewSelectCollection(WithSignals(ExitOnSelect, os.Interrupt)),
			initialize: true, initializeSuccessExpected: true, initializeErrCode: nil,
			selectOnCol: true, shutdownBeforeSelectInGoSub: true, shutdownWaitInGoSub: 100 * time.Millisecond, shutdownSuccessExpected: true,
		},
	})
	t.Run("SignalOps", func(t *testing.T) {
		sigChannel := make(chan os.Signal)
		signal.Notify(sigChannel, os.Interrupt)
		signalAsString := fmt.Sprintf("%s", []os.Signal{os.Interrupt})
		signalSelect := &signalSelection{
			source:           sigChannel,
			signals:          signalAsString,
			onSelectFunction: ExitOnSelect,
		}
		if signalSelect.GetSource() != sigChannel {
			t.Errorf("Expected %#v, Actual %#v", sigChannel, signalSelect.GetSource())
		}
		if signalSelect.GetOnSelect() == nil {
			t.Errorf("Expected nil, Actual %#v", signalSelect.GetOnSelect())
		}
		randomEventErr := signalSelect.Execute(SelectorEvent(5))
		if randomEventErr == nil {
			t.Errorf("Expected unsupported event error not nil actual nil")
		} else if _, isErr := ErrSelectionConfigExecute.AsError(randomEventErr); !isErr {
			t.Errorf("Expected unsupported event error OF TYPE ErrSelectionConfigExecute actual %#v", randomEventErr)
		}
		stopErr := signalSelect.Execute(SelectorEventStop)
		if stopErr != nil {
			t.Errorf("Expected no error, actual %#v", stopErr)
		}
	})
}

func OnSelectOnce(event SelectEvent, collection SelectCollection) (continueSelecting bool) {
	fmt.Println("Selecting once")
	return false
}

func PrintChannelIdOnMonitor(identifier SelectorIdentifier) error {
	fmt.Println("Channel id", identifier)
	return nil
}

func PushValuesInGoSub(parameters ...interface{}) func(identifier SelectorIdentifier, channel chan interface{}) {
	return func(identifier SelectorIdentifier, channel chan interface{}) {
		go func() {
			for _, item := range parameters {
				if asEvent, isEvent := item.(SelectEvent); isEvent {
					asEvent.Source = identifier
				}
				fmt.Println("Sending to channel", item)
				channel <- item
			}
		}()
	}
}

type testCaseDef struct {
	name                        string
	collection                  SelectCollection
	register                    bool
	registerValue               SelectionConfig
	registerSuccessExpected     bool
	registerErrCode             errext.ErrorCode
	initialize                  bool
	initializeSuccessExpected   bool
	initializeErrCode           errext.ErrorCode
	selectOnCol                 bool
	shutdownBeforeSelectInGoSub bool
	shutdownWaitInGoSub         time.Duration
	shutdown                    bool
	shutdownSuccessExpected     bool
	shutdownErrCode             errext.ErrorCode
}

func runTest(t *testing.T, testCases []testCaseDef) {
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			collection := testCase.collection
			selectId := SelectorIdentifierInvalid
			if testCase.register {
				var registerErr error = nil
				selectId, registerErr = collection.Register(testCase.registerValue)
				if testCase.registerSuccessExpected {
					if registerErr != nil {
						t.Errorf("Expected registration success, actual %#v", registerErr)
					}
					if selectId == SelectorIdentifierInvalid {
						t.Errorf("Expected valid select identifier, actual %#v", selectId)
					}
				} else {
					if registerErr == nil {
						t.Errorf("Expected registration failure, actual no error")
					}
					if selectId != SelectorIdentifierInvalid {
						t.Errorf("Expected SelectorIdentifierInvalid, actual %#v", selectId)
					}
				}
				if asDummyConfig, isDummyConfig := testCase.registerValue.(*dummySelectConfig); isDummyConfig {
					asDummyConfig.selectId = selectId
				}
			}
			if testCase.initialize {
				var initErr error = nil
				initErr = collection.Initialize()
				if testCase.initializeSuccessExpected {
					if initErr != nil {
						t.Errorf("Expected initialize success, actual %#v", initErr)
					}
				} else {
					if initErr == nil {
						t.Errorf("Expected initialization failure, actual no error")
					}
				}
			}
			if testCase.shutdownBeforeSelectInGoSub {
				go func(collection SelectCollection) {
					time.Sleep(testCase.shutdownWaitInGoSub)
					shutDownErr := collection.Shutdown()
					if testCase.shutdownSuccessExpected && shutDownErr != nil {
						t.Errorf("Expected successful shutdown, actual %#v", shutDownErr)
					} else if !testCase.shutdownSuccessExpected && shutDownErr == nil {
						t.Errorf("Expected error during shutdown, actual no error")
					}
				}(collection)
			}
			if testCase.selectOnCol {
				collection.Select()
			}
			if !testCase.shutdownBeforeSelectInGoSub && testCase.shutdown {
				shutDownErr := collection.Shutdown()
				if testCase.shutdownSuccessExpected && shutDownErr != nil {
					t.Errorf("Expected successful shutdown, actual %#v", shutDownErr)
				} else if !testCase.shutdownSuccessExpected && shutDownErr == nil {
					t.Errorf("Expected error during shutdown, actual no error")
				}
			}
		})
	}
}
