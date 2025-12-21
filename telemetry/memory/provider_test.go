package memory

import (
	"context"
	"testing"
	"time"

	"github.com/grinps/go-utils/telemetry"
)

func TestNewProvider(t *testing.T) {
	p := NewProvider()
	if p == nil {
		t.Fatal("Expected non-nil provider")
	}
	if p.shutdown {
		t.Error("Expected provider to not be shutdown")
	}
}

func TestProvider_Tracer(t *testing.T) {
	p := NewProvider()
	tracer, err := p.Tracer("test-tracer")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if tracer == nil {
		t.Fatal("Expected non-nil tracer")
	}

	// Same name should return same tracer
	tracer2, err := p.Tracer("test-tracer")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if tracer != tracer2 {
		t.Error("Expected same tracer instance for same name")
	}
}

func TestProvider_Meter(t *testing.T) {
	p := NewProvider()
	meter, err := p.Meter("test-meter")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if meter == nil {
		t.Fatal("Expected non-nil meter")
	}

	// Same name should return same meter
	meter2, err := p.Meter("test-meter")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if meter != meter2 {
		t.Error("Expected same meter instance for same name")
	}
}

func TestProvider_Shutdown(t *testing.T) {
	p := NewProvider()

	err := p.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("Expected no error on first shutdown, got: %v", err)
	}

	if !p.IsShutdown() {
		t.Error("Expected provider to be shutdown")
	}

	// Second shutdown should return error
	err = p.Shutdown(context.Background())
	if err == nil {
		t.Error("Expected error on second shutdown")
	}
}

func TestProvider_TracerAfterShutdown(t *testing.T) {
	p := NewProvider()
	_ = p.Shutdown(context.Background())

	tracer, err := p.Tracer("test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should return noop tracer
	ctx, span := tracer.Start(context.Background(), "test-span")
	if ctx == nil {
		t.Error("Expected non-nil context")
	}
	if span == nil {
		t.Error("Expected non-nil span")
	}
	if span.IsRecording() {
		t.Error("Noop span should not be recording")
	}
}

func TestProvider_MeterAfterShutdown(t *testing.T) {
	p := NewProvider()
	_ = p.Shutdown(context.Background())

	meter, err := p.Meter("test")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if meter == nil {
		t.Error("Expected non-nil meter")
	}
}

func TestTracer_Start(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	ctx, span := tracer.Start(context.Background(), "test-operation")
	if ctx == nil {
		t.Fatal("Expected non-nil context")
	}
	if span == nil {
		t.Fatal("Expected non-nil span")
	}
	if !span.IsRecording() {
		t.Error("Expected span to be recording")
	}

	span.End()

	if span.IsRecording() {
		t.Error("Expected span to not be recording after End")
	}
}

func TestSpan_SetAttributes(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test-operation")
	defer span.End()

	// Set attributes using local Attribute type
	span.SetAttributes(String("key", "value"))
}

func TestSpan_AddEvent(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test-operation")
	defer span.End()

	span.AddEvent("test-event")
}

func TestSpan_RecordError(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test-operation")
	defer span.End()

	err := ErrMemoryProviderOperation.New(ErrReasonMemoryProviderShutdown)
	span.RecordError(err)
}

func TestSpan_SetStatus(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test-operation")
	defer span.End()

	span.SetStatus(int(StatusError), "something went wrong")
}

func TestSpan_SetName(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "original-name")
	defer span.End()

	span.SetName("new-name")
}

func TestSpan_TracerProvider(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test-operation")
	defer span.End()

	tp := span.TracerProvider()
	if tp == nil {
		t.Error("Expected non-nil TracerProvider")
	}
}

func TestProvider_RecordedSpans(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	// Initially no spans
	spans := p.RecordedSpans()
	if len(spans) != 0 {
		t.Errorf("Expected 0 spans initially, got %d", len(spans))
	}

	// Create and end a span
	_, span := tracer.Start(context.Background(), "test-operation")
	span.End()

	// Should have 1 recorded span
	spans = p.RecordedSpans()
	if len(spans) != 1 {
		t.Errorf("Expected 1 span, got %d", len(spans))
	}
	if spans[0].Name != "test-operation" {
		t.Errorf("Expected span name 'test-operation', got '%s'", spans[0].Name)
	}
}

func TestProvider_RecordedSpansByName(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	// Create multiple spans
	_, span1 := tracer.Start(context.Background(), "operation-a")
	span1.End()

	_, span2 := tracer.Start(context.Background(), "operation-b")
	span2.End()

	_, span3 := tracer.Start(context.Background(), "operation-a")
	span3.End()

	// Filter by name
	spansA := p.RecordedSpansByName("operation-a")
	if len(spansA) != 2 {
		t.Errorf("Expected 2 spans named 'operation-a', got %d", len(spansA))
	}

	spansB := p.RecordedSpansByName("operation-b")
	if len(spansB) != 1 {
		t.Errorf("Expected 1 span named 'operation-b', got %d", len(spansB))
	}
}

func TestProvider_Reset(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	// Create and end a span
	_, span := tracer.Start(context.Background(), "test-operation")
	span.End()

	// Verify span recorded
	if len(p.RecordedSpans()) != 1 {
		t.Fatal("Expected 1 span before reset")
	}

	// Reset
	p.Reset()

	// Verify spans cleared
	if len(p.RecordedSpans()) != 0 {
		t.Error("Expected 0 spans after reset")
	}
}

func TestRecordedSpan_HasAttribute(t *testing.T) {
	span := &RecordedSpan{
		Attributes: []Attribute{
			String("key1", "value1"),
			Int64("key2", 42),
		},
	}

	if !span.HasAttribute("key1") {
		t.Error("Expected HasAttribute to return true for 'key1'")
	}
	if !span.HasAttribute("key2") {
		t.Error("Expected HasAttribute to return true for 'key2'")
	}
	if span.HasAttribute("key3") {
		t.Error("Expected HasAttribute to return false for 'key3'")
	}
}

func TestRecordedSpan_GetAttribute(t *testing.T) {
	span := &RecordedSpan{
		Attributes: []Attribute{
			String("key1", "value1"),
		},
	}

	val := span.GetAttribute("key1")
	if val != "value1" {
		t.Errorf("Expected 'value1', got '%v'", val)
	}

	val = span.GetAttribute("nonexistent")
	if val != nil {
		t.Errorf("Expected nil for nonexistent key, got '%v'", val)
	}
}

func TestRecordedSpan_HasEvent(t *testing.T) {
	span := &RecordedSpan{
		Events: []Event{
			{Name: "event1"},
			{Name: "event2"},
		},
	}

	if !span.HasEvent("event1") {
		t.Error("Expected HasEvent to return true for 'event1'")
	}
	if span.HasEvent("event3") {
		t.Error("Expected HasEvent to return false for 'event3'")
	}
}

func TestRecordedSpan_Duration(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test-operation")
	span.End()

	spans := p.RecordedSpans()
	if len(spans) != 1 {
		t.Fatal("Expected 1 span")
	}

	duration := spans[0].Duration()
	if duration <= 0 {
		t.Error("Expected positive duration")
	}
}

func TestContextTelemetry(t *testing.T) {
	p := NewProvider()

	// Store provider in context
	ctx := telemetry.ContextWithTelemetry(context.Background(), p)

	// Retrieve from context
	retrieved := telemetry.ContextTelemetry(ctx, true)
	if retrieved != p {
		t.Error("Expected same provider from context")
	}
}

func TestContextTracer(t *testing.T) {
	p := NewProvider()
	ctx := telemetry.ContextWithTelemetry(context.Background(), p)

	tracer := telemetry.ContextTracer(ctx, true)
	if tracer == nil {
		t.Error("Expected non-nil tracer")
	}
}

func TestContextMeter(t *testing.T) {
	p := NewProvider()
	ctx := telemetry.ContextWithTelemetry(context.Background(), p)

	meter := telemetry.ContextMeter(ctx, true)
	if meter == nil {
		t.Error("Expected non-nil meter")
	}
}

func TestNewProvider_WithOptions(t *testing.T) {
	// Test that NewProvider accepts options
	opt := func(p *Provider) {
		// Custom option
	}
	p := NewProvider(opt)
	if p == nil {
		t.Fatal("Expected non-nil provider")
	}
}

func TestTracer_Start_WithOptions(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	// Test with span kind
	ctx, span := tracer.Start(context.Background(), "operation", SpanKindServer)
	if span == nil {
		t.Fatal("Expected non-nil span")
	}
	span.End()

	spans := p.RecordedSpans()
	if spans[0].Kind != SpanKindServer {
		t.Error("Expected SpanKindServer")
	}

	// Test with attributes
	_, span2 := tracer.Start(ctx, "operation2", String("key", "value"))
	span2.End()

	spans = p.RecordedSpans()
	if !spans[1].HasAttribute("key") {
		t.Error("Expected attribute 'key'")
	}

	// Test with links
	link := Link{SpanContext: spans[0].SpanContext}
	_, span3 := tracer.Start(ctx, "operation3", link)
	span3.End()

	spans = p.RecordedSpans()
	if len(spans[2].Links) != 1 {
		t.Errorf("Expected 1 link, got %d", len(spans[2].Links))
	}
}

func TestTracer_Start_ChildSpan(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	// Parent span
	ctx, parentSpan := tracer.Start(context.Background(), "parent")

	// Child span should inherit trace ID
	_, childSpan := tracer.Start(ctx, "child")
	childSpan.End()
	parentSpan.End()

	spans := p.RecordedSpans()
	if len(spans) != 2 {
		t.Fatalf("Expected 2 spans, got %d", len(spans))
	}

	// Find child span (ended first)
	var child, parent *RecordedSpan
	for _, s := range spans {
		if s.Name == "child" {
			child = s
		} else {
			parent = s
		}
	}

	// Child should have parent's span ID as parent
	if child.ParentSpanID != parent.SpanContext.SpanID() {
		t.Error("Child should have parent's span ID")
	}

	// Both should have same trace ID
	if child.SpanContext.TraceID() != parent.SpanContext.TraceID() {
		t.Error("Child and parent should have same trace ID")
	}
}

func TestSpan_EndTwice(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test")
	span.End()
	span.End() // Should be no-op

	// Only 1 span should be recorded
	if len(p.RecordedSpans()) != 1 {
		t.Errorf("Expected 1 span, got %d", len(p.RecordedSpans()))
	}
}

func TestSpan_OperationsAfterEnd(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test")
	span.End()

	// Operations after end should be no-op
	span.SetAttributes(String("key", "value"))
	span.AddEvent("event")
	span.RecordError(ErrMemoryProviderOperation.New(ErrReasonMemoryProviderShutdown))
	span.SetStatus(int(StatusError), "error")
	span.SetName("new-name")

	// Verify no changes were made
	spans := p.RecordedSpans()
	if len(spans[0].Attributes) != 0 {
		t.Error("Expected no attributes after end")
	}
	if len(spans[0].Events) != 0 {
		t.Error("Expected no events after end")
	}
}

func TestSpan_RecordNilError(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test")
	span.RecordError(nil) // Should be no-op
	span.End()

	spans := p.RecordedSpans()
	if len(spans[0].Events) != 0 {
		t.Error("Expected no events for nil error")
	}
}

func TestSpan_SetStatusPriority(t *testing.T) {
	p := NewProvider()
	tracer, _ := p.Tracer("test")

	_, span := tracer.Start(context.Background(), "test")
	span.SetStatus(int(StatusOK), "ok")
	span.SetStatus(int(StatusError), "error") // Higher priority
	span.SetStatus(int(StatusOK), "ok again") // Lower priority, ignored
	span.End()

	spans := p.RecordedSpans()
	if spans[0].Status.Code != StatusError {
		t.Errorf("Expected StatusError, got %v", spans[0].Status.Code)
	}
}

func TestSpanContext_Methods(t *testing.T) {
	traceID := generateTraceID()
	spanID := generateSpanID()
	sc := NewSpanContext(traceID, spanID, TraceFlagsSampled, true)

	if sc.TraceID() != traceID {
		t.Error("TraceID mismatch")
	}
	if sc.SpanID() != spanID {
		t.Error("SpanID mismatch")
	}
	if sc.TraceFlags() != TraceFlagsSampled {
		t.Error("TraceFlags mismatch")
	}
	if !sc.IsRemote() {
		t.Error("Expected IsRemote to be true")
	}
	if !sc.IsValid() {
		t.Error("Expected IsValid to be true")
	}
	if !sc.TraceFlags().IsSampled() {
		t.Error("Expected IsSampled to be true")
	}
}

func TestTraceID_IsValid(t *testing.T) {
	var zeroID TraceID
	if zeroID.IsValid() {
		t.Error("Zero TraceID should be invalid")
	}

	validID := generateTraceID()
	if !validID.IsValid() {
		t.Error("Generated TraceID should be valid")
	}
}

func TestSpanID_IsValid(t *testing.T) {
	var zeroID SpanID
	if zeroID.IsValid() {
		t.Error("Zero SpanID should be invalid")
	}

	validID := generateSpanID()
	if !validID.IsValid() {
		t.Error("Generated SpanID should be valid")
	}
}

func TestAttribute_Helpers(t *testing.T) {
	strAttr := String("str", "value")
	if strAttr.Key != "str" || strAttr.Value != "value" {
		t.Error("String attribute mismatch")
	}

	intAttr := Int64("int", 42)
	if intAttr.Key != "int" || intAttr.Value != int64(42) {
		t.Error("Int64 attribute mismatch")
	}

	floatAttr := Float64("float", 3.14)
	if floatAttr.Key != "float" || floatAttr.Value != 3.14 {
		t.Error("Float64 attribute mismatch")
	}

	boolAttr := Bool("bool", true)
	if boolAttr.Key != "bool" || boolAttr.Value != true {
		t.Error("Bool attribute mismatch")
	}
}

func TestNewEvent(t *testing.T) {
	event := NewEvent("test-event", String("key", "value"))
	if event.Name != "test-event" {
		t.Error("Event name mismatch")
	}
	if len(event.Attributes) != 1 {
		t.Error("Expected 1 attribute")
	}
	if event.Timestamp.IsZero() {
		t.Error("Expected non-zero timestamp")
	}
}

func TestSpanFromContext(t *testing.T) {
	// Nil context
	span := SpanFromContext(nil)
	if span != nil {
		t.Error("Expected nil span from nil context")
	}

	// Empty context
	span = SpanFromContext(context.Background())
	if span != nil {
		t.Error("Expected nil span from empty context")
	}

	// Context with span
	p := NewProvider()
	tracer, _ := p.Tracer("test")
	ctx, expectedSpan := tracer.Start(context.Background(), "test")

	span = SpanFromContext(ctx)
	if span != expectedSpan {
		t.Error("Expected same span from context")
	}
}

func TestRecordedSpan_Duration_NotEnded(t *testing.T) {
	span := &RecordedSpan{}
	if span.Duration() != 0 {
		t.Error("Expected zero duration for unended span")
	}
}

func TestInstrument_DescriptionAndUnit(t *testing.T) {
	p := NewProvider()
	meter, _ := p.Meter("test")

	inst, _ := meter.NewInstrument("test",
		telemetry.InstrumentTypeCounter,
		telemetry.CounterTypeMonotonic,
		"Test description",
		"requests",
	)

	// Verify instrument was created (description/unit are stored internally)
	if inst == nil {
		t.Error("Expected non-nil instrument")
	}
	counter, ok := inst.(*Counter[int64])
	if !ok {
		t.Fatal("Expected Counter[int64]")
	}
	if counter.Precision() != telemetry.PrecisionInt64 {
		t.Errorf("Expected PrecisionInt64, got %s", counter.Precision())
	}
}

func TestParseTracerOptions(t *testing.T) {
	// With version string
	cfg := parseTracerOptions("1.0.0")
	if cfg.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", cfg.version)
	}

	// With TracerConfig
	tracerCfg := TracerConfig{version: "2.0.0"}
	cfg = parseTracerOptions(tracerCfg)
	if cfg.version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", cfg.version)
	}

	// With *TracerConfig
	cfg = parseTracerOptions(&TracerConfig{version: "3.0.0"})
	if cfg.version != "3.0.0" {
		t.Errorf("Expected version '3.0.0', got '%s'", cfg.version)
	}

	// Nil pointer should be ignored
	var nilCfg *TracerConfig
	cfg = parseTracerOptions(nilCfg, "4.0.0")
	if cfg.version != "4.0.0" {
		t.Errorf("Expected version '4.0.0', got '%s'", cfg.version)
	}
}

func TestParseMeterOptions(t *testing.T) {
	// With version string
	cfg := parseMeterOptions("1.0.0")
	if cfg.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", cfg.version)
	}

	// With MeterConfig
	meterCfg := MeterConfig{version: "2.0.0"}
	cfg = parseMeterOptions(meterCfg)
	if cfg.version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", cfg.version)
	}

	// With *MeterConfig
	cfg = parseMeterOptions(&MeterConfig{version: "3.0.0"})
	if cfg.version != "3.0.0" {
		t.Errorf("Expected version '3.0.0', got '%s'", cfg.version)
	}

	// Nil pointer should be ignored
	var nilCfg *MeterConfig
	cfg = parseMeterOptions(nilCfg, "4.0.0")
	if cfg.version != "4.0.0" {
		t.Errorf("Expected version '4.0.0', got '%s'", cfg.version)
	}
}

func TestParseSpanOptions(t *testing.T) {
	// With SpanConfig
	spanCfg := SpanConfig{kind: SpanKindClient}
	cfg := parseSpanOptions(spanCfg)
	if cfg.kind != SpanKindClient {
		t.Errorf("Expected SpanKindClient, got %v", cfg.kind)
	}

	// With *SpanConfig
	cfg = parseSpanOptions(&SpanConfig{kind: SpanKindProducer})
	if cfg.kind != SpanKindProducer {
		t.Errorf("Expected SpanKindProducer, got %v", cfg.kind)
	}

	// With attribute slice
	cfg = parseSpanOptions([]Attribute{String("k", "v")})
	if len(cfg.attributes) != 1 {
		t.Error("Expected 1 attribute")
	}

	// With links slice
	cfg = parseSpanOptions([]Link{{SpanContext: SpanContext{}}})
	if len(cfg.links) != 1 {
		t.Error("Expected 1 link")
	}

	// Nil pointer should be ignored
	var nilCfg *SpanConfig
	cfg = parseSpanOptions(nilCfg)
	if cfg.kind != SpanKindInternal {
		t.Error("Expected default SpanKindInternal")
	}
}

func TestParseSpanEndOptions(t *testing.T) {
	// With SpanEndConfig
	endCfg := SpanEndConfig{}
	cfg := parseSpanEndOptions(endCfg)
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	// With *SpanEndConfig
	cfg = parseSpanEndOptions(&SpanEndConfig{})
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	// Nil pointer should be ignored
	var nilCfg *SpanEndConfig
	cfg = parseSpanEndOptions(nilCfg)
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}
}

func TestParseEventOptions(t *testing.T) {
	// With EventConfig
	eventCfg := EventConfig{}
	cfg := parseEventOptions(eventCfg)
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	// With *EventConfig
	cfg = parseEventOptions(&EventConfig{})
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}

	// With attributes slice
	cfg = parseEventOptions([]Attribute{String("k", "v")})
	if len(cfg.attributes) != 1 {
		t.Error("Expected 1 attribute")
	}

	// Nil pointer should be ignored
	var nilCfg *EventConfig
	cfg = parseEventOptions(nilCfg)
	if cfg == nil {
		t.Fatal("Expected non-nil config")
	}
}

func TestParseInstrumentOptions(t *testing.T) {
	// With InstrumentConfig
	instCfg := InstrumentConfig{description: "desc"}
	cfg := parseInstrumentOptions(instCfg)
	if cfg.description != "desc" {
		t.Errorf("Expected 'desc', got '%s'", cfg.description)
	}

	// With *InstrumentConfig
	cfg = parseInstrumentOptions(&InstrumentConfig{description: "desc2"})
	if cfg.description != "desc2" {
		t.Errorf("Expected 'desc2', got '%s'", cfg.description)
	}

	// With buckets
	cfg = parseInstrumentOptions([]float64{1.0, 2.0, 5.0})
	if len(cfg.buckets) != 3 {
		t.Error("Expected 3 buckets")
	}

	// Nil pointer should be ignored
	var nilCfg *InstrumentConfig
	cfg = parseInstrumentOptions(nilCfg, "description", "unit")
	if cfg.description != "description" || cfg.unit != "unit" {
		t.Error("Expected description and unit to be set")
	}
}

func TestParseTracerOptions_KeyValue(t *testing.T) {
	// Test key-value pairs for version and schemaURL
	cfg := parseTracerOptions("version", "1.0.0", "schemaURL", "http://example.com")
	if cfg.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", cfg.version)
	}
	if cfg.schemaURL != "http://example.com" {
		t.Errorf("Expected schemaURL 'http://example.com', got '%s'", cfg.schemaURL)
	}

	// Test just version key-value
	cfg = parseTracerOptions("version", "2.0.0")
	if cfg.version != "2.0.0" {
		t.Errorf("Expected version '2.0.0', got '%s'", cfg.version)
	}

	// Test just schemaURL key-value
	cfg = parseTracerOptions("schemaURL", "http://schema.example.com")
	if cfg.schemaURL != "http://schema.example.com" {
		t.Errorf("Expected schemaURL, got '%s'", cfg.schemaURL)
	}

	// Test attributes from key-value pairs
	cfg = parseTracerOptions("version", "1.0.0", "service.env", "prod", "service.region", "us-east-1")
	if cfg.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", cfg.version)
	}
	if len(cfg.attributes) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(cfg.attributes))
	}
	if cfg.attributes[0].Key != "service.env" || cfg.attributes[0].Value != "prod" {
		t.Errorf("First attribute mismatch: %v", cfg.attributes[0])
	}
	if cfg.attributes[1].Key != "service.region" || cfg.attributes[1].Value != "us-east-1" {
		t.Errorf("Second attribute mismatch: %v", cfg.attributes[1])
	}

	// Test with Attribute type directly
	cfg = parseTracerOptions(String("key1", "val1"), "key2", 42)
	if len(cfg.attributes) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(cfg.attributes))
	}
}

func TestParseMeterOptions_KeyValue(t *testing.T) {
	// Test key-value pairs for version and schemaURL
	cfg := parseMeterOptions("version", "1.0.0", "schemaURL", "http://example.com")
	if cfg.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", cfg.version)
	}
	if cfg.schemaURL != "http://example.com" {
		t.Errorf("Expected schemaURL 'http://example.com', got '%s'", cfg.schemaURL)
	}

	// Test attributes from key-value pairs
	cfg = parseMeterOptions("version", "1.0.0", "service.name", "my-service", "count", 100)
	if cfg.version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", cfg.version)
	}
	if len(cfg.attributes) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(cfg.attributes))
	}
	if cfg.attributes[0].Key != "service.name" || cfg.attributes[0].Value != "my-service" {
		t.Errorf("First attribute mismatch: %v", cfg.attributes[0])
	}
	if cfg.attributes[1].Key != "count" || cfg.attributes[1].Value != 100 {
		t.Errorf("Second attribute mismatch: %v", cfg.attributes[1])
	}

	// Test with Attribute type directly
	cfg = parseMeterOptions(String("key1", "val1"), "key2", 3.14)
	if len(cfg.attributes) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(cfg.attributes))
	}
}

func TestParseEventOptions_KeyValue(t *testing.T) {
	ts := time.Now()

	// Test time.Time as first entry and key-value pairs
	cfg := parseEventOptions(ts, "user.id", "12345", "count", 10)
	if cfg.timestamp != ts {
		t.Error("Expected timestamp to be set")
	}
	if len(cfg.attributes) != 2 {
		t.Fatalf("Expected 2 attributes, got %d", len(cfg.attributes))
	}
	if cfg.attributes[0].Key != "user.id" || cfg.attributes[0].Value != "12345" {
		t.Errorf("First attribute mismatch: %v", cfg.attributes[0])
	}
	if cfg.attributes[1].Key != "count" || cfg.attributes[1].Value != 10 {
		t.Errorf("Second attribute mismatch: %v", cfg.attributes[1])
	}

	// Test key-value pairs without timestamp
	cfg = parseEventOptions("key", "value")
	if len(cfg.attributes) != 1 {
		t.Fatalf("Expected 1 attribute, got %d", len(cfg.attributes))
	}
	if cfg.attributes[0].Key != "key" {
		t.Error("Expected 'key' attribute")
	}
}
