package telemetry

import (
	"context"
)

// Tracer is responsible for creating Spans.
// It is obtained from a Provider and should be named to identify
// the instrumentation library or application component.
//
// Implementations must be safe for concurrent use.
type Tracer interface {
	// Start creates and starts a new Span.
	// The returned context contains the new span and should be used for
	// any downstream operations that should be children of this span.
	//
	// The span name should describe the operation being performed.
	// It should be a low-cardinality string (e.g., "HTTP GET /users", not "HTTP GET /users/123").
	Start(ctx context.Context, name string, opts ...any) (context.Context, Span)
}
