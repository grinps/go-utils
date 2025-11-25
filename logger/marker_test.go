package logger

import (
"context"
"log/slog"
"testing"
)

func TestMarker(t *testing.T) {
	m := NewMarker("Service")
	if m.String() != "Service" {
		t.Errorf("Expected Service, got %s", m.String())
	}

	m2 := m.Add("Database")
	if m2.String() != "Service.Database" {
		t.Errorf("Expected Service.Database, got %s", m2.String())
	}

	m3 := m2.AddMethod("Query", "id", 1)
	if m3.String() != "Service.Database.Query(id 1)" {
		t.Errorf("Expected Service.Database.Query(id 1), got %s", m3.String())
	}
}

func TestTrace(t *testing.T) {
	// 1. Setup initial context with a logger (optional, but good for testing isolation)
	ctx := context.Background()
	
	// 2. Call Entering
	ctx = Entering(ctx, "TestFunc")
	
	// 3. Verify Marker in Context
	m := FromContext(ctx)
	if m.String() != "TestFunc()" {
		t.Errorf("Expected TestFunc(), got %s", m.String())
	}

	// 4. Verify Logger in Context
	log := LoggerFromContext(ctx)
	if log == nil {
		t.Error("Expected logger in context, got nil")
	}
	// Ideally we'd check if the logger has the marker handler, but slog internals are private.
// We rely on the fact that Entering calls .With()

// 5. Call Exiting (should not panic)
Exiting(ctx)
}

func TestWithLogger(t *testing.T) {
ctx := context.Background()
l := slog.Default()
ctx = WithLogger(ctx, l)

if LoggerFromContext(ctx) != l {
t.Error("LoggerFromContext did not return the logger set by WithLogger")
}
}
