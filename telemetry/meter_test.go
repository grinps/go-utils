package telemetry

import "testing"

func TestAggregationStrategy_Name(t *testing.T) {
	tests := []struct {
		strategy AggregationStrategy
		want     string
	}{
		{AggregationStrategyUnknown, "unknown"},
		{AggregationStrategyNone, "none"},
		{AggregationStrategyHistogram, "histogram"},
	}

	for _, tt := range tests {
		t.Run(string(tt.strategy), func(t *testing.T) {
			if got := tt.strategy.Name(); got != tt.want {
				t.Errorf("AggregationStrategy.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInstrumentType_Constants(t *testing.T) {
	if InstrumentTypeUnknown != "" {
		t.Errorf("InstrumentTypeUnknown = %q, want empty", InstrumentTypeUnknown)
	}
	if InstrumentTypeCounter != "counter" {
		t.Errorf("InstrumentTypeCounter = %q, want counter", InstrumentTypeCounter)
	}
	if InstrumentTypeRecorder != "recorder" {
		t.Errorf("InstrumentTypeRecorder = %q, want recorder", InstrumentTypeRecorder)
	}
}

func TestCounterType_Constants(t *testing.T) {
	if CounterTypeUnknown != "" {
		t.Errorf("CounterTypeUnknown = %q, want empty", CounterTypeUnknown)
	}
	if CounterTypeMonotonic != "monotonic" {
		t.Errorf("CounterTypeMonotonic = %q, want monotonic", CounterTypeMonotonic)
	}
	if CounterTypeUpDown != "updown" {
		t.Errorf("CounterTypeUpDown = %q, want updown", CounterTypeUpDown)
	}
}

func TestAggregationStrategy_Constants(t *testing.T) {
	if AggregationStrategyUnknown != "unknown" {
		t.Errorf("AggregationStrategyUnknown = %q, want unknown", AggregationStrategyUnknown)
	}
	if AggregationStrategyNone != "none" {
		t.Errorf("AggregationStrategyNone = %q, want none", AggregationStrategyNone)
	}
	if AggregationStrategyHistogram != "histogram" {
		t.Errorf("AggregationStrategyHistogram = %q, want histogram", AggregationStrategyHistogram)
	}
}
