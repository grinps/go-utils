package telemetry

import "testing"

func TestErrorParameterConstants(t *testing.T) {
	// Verify error parameter constants are defined
	params := map[string]string{
		"ErrParamName":                ErrParamName,
		"ErrParamReason":              ErrParamReason,
		"ErrParamType":                ErrParamType,
		"ErrParamOperation":           ErrParamOperation,
		"ErrParamValue":               ErrParamValue,
		"ErrParamKey":                 ErrParamKey,
		"ErrParamInstrumentType":      ErrParamInstrumentType,
		"ErrParamCounterType":         ErrParamCounterType,
		"ErrParamAggregationStrategy": ErrParamAggregationStrategy,
		"ErrParamOption":              ErrParamOption,
	}

	for name, value := range params {
		if value == "" {
			t.Errorf("%s should not be empty", name)
		}
	}
}

func TestErrorReasonConstants(t *testing.T) {
	// Verify error reason constants are defined
	reasons := map[string]string{
		"ErrReasonNilProvider":                ErrReasonNilProvider,
		"ErrReasonNilTracer":                  ErrReasonNilTracer,
		"ErrReasonNilMeter":                   ErrReasonNilMeter,
		"ErrReasonNilSpan":                    ErrReasonNilSpan,
		"ErrReasonNilContext":                 ErrReasonNilContext,
		"ErrReasonEmptyName":                  ErrReasonEmptyName,
		"ErrReasonInvalidValue":               ErrReasonInvalidValue,
		"ErrReasonAlreadyShutdown":            ErrReasonAlreadyShutdown,
		"ErrReasonNotInitialized":             ErrReasonNotInitialized,
		"ErrReasonInstrumentExists":           ErrReasonInstrumentExists,
		"ErrReasonInvalidInstrumentType":      ErrReasonInvalidInstrumentType,
		"ErrReasonInvalidCounterType":         ErrReasonInvalidCounterType,
		"ErrReasonInvalidAggregationStrategy": ErrReasonInvalidAggregationStrategy,
		"ErrReasonInvalidOption":              ErrReasonInvalidOption,
	}

	for name, value := range reasons {
		if value == "" {
			t.Errorf("%s should not be empty", name)
		}
	}
}

func TestErrorCodes(t *testing.T) {
	// Verify error codes are not nil
	codes := []struct {
		name string
		code interface{}
	}{
		{"ErrProviderCreation", ErrProviderCreation},
		{"ErrProviderShutdown", ErrProviderShutdown},
		{"ErrTracerCreation", ErrTracerCreation},
		{"ErrSpanCreation", ErrSpanCreation},
		{"ErrSpanOperation", ErrSpanOperation},
		{"ErrMeterCreation", ErrMeterCreation},
		{"ErrInstrumentCreation", ErrInstrumentCreation},
		{"ErrInstrumentOperation", ErrInstrumentOperation},
		{"ErrInvalidAttribute", ErrInvalidAttribute},
		{"ErrContextPropagation", ErrContextPropagation},
	}

	for _, tc := range codes {
		if tc.code == nil {
			t.Errorf("%s should not be nil", tc.name)
		}
	}
}

func TestErrTypePrefix(t *testing.T) {
	if ErrTypePrefix == "" {
		t.Error("ErrTypePrefix should not be empty")
	}
	expected := "github.com/grinps/go-utils/telemetry"
	if ErrTypePrefix != expected {
		t.Errorf("ErrTypePrefix = %q, want %q", ErrTypePrefix, expected)
	}
}

func TestErrorCodeCreation(t *testing.T) {
	// Test that error codes can create errors
	err := ErrProviderCreation.New(ErrReasonNilProvider)
	if err == nil {
		t.Error("ErrProviderCreation.New() returned nil")
	}

	err = ErrTracerCreation.New(ErrReasonNilTracer, ErrParamName, "test")
	if err == nil {
		t.Error("ErrTracerCreation.New() with params returned nil")
	}
}
