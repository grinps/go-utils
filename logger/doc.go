// Package logger provides extensions for Go's standard log/slog package.
//
// # Overview
//
// This package enhances log/slog with a hierarchical "Marker" system for
// context-aware method tracing. It is designed to be lightweight and
// fully compatible with standard slog.Logger.
//
// # Markers
//
// Markers allow you to track the execution path (e.g., "Service.DB.Query(id=1)").
// They implement slog.LogValuer, so they can be added directly as attributes.
//
//	m := logger.NewMarker("MyService")
//	slog.Info("Started", "marker", m)
//
// # Method Tracing & Context Integration
//
// The package provides helpers to automatically manage markers and loggers within
// the context, reducing boilerplate in function signatures.
//
//	func MyFunc(ctx context.Context) {
//	    // Entering automatically:
//	    // 1. Updates the context with a new marker (Parent.MyFunc)
//	    // 2. Retrieves or creates a logger scoped with that marker
//	    // 3. Stores both back in the returned context
//	    ctx = logger.Entering(ctx, "MyFunc", "id", 123)
//	    defer logger.Exiting(ctx)
//
//	    // Retrieve the scoped logger to log with the current marker
//	    log := logger.LoggerFromContext(ctx)
//	    log.Info("Doing work...") // Log will have "marker"="Parent.MyFunc(id 123)"
//	}
package logger
