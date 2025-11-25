package logger

import (
	"context"
	"log/slog"
)

const (
	// LevelTrace is a log level for detailed tracing, lower than Debug.
	// slog.LevelDebug is -4, so we set Trace to -8.
	LevelTrace = slog.Level(-8)
)

type contextKey struct{}
type loggerKey struct{}

var (
	markerContextKey = contextKey{}
	loggerContextKey = loggerKey{}
)

// FromContext retrieves the Marker from the context.
// If no marker is found, it returns an empty Marker.
func FromContext(ctx context.Context) Marker {
	if m, ok := ctx.Value(markerContextKey).(Marker); ok {
		return m
	}
	return Marker{}
}

// WithMarker returns a new context with the given Marker.
func WithMarker(ctx context.Context, m Marker) context.Context {
	return context.WithValue(ctx, markerContextKey, m)
}

// LoggerFromContext retrieves the logger from the context.
// If no logger is found, it returns slog.Default().
func LoggerFromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerContextKey).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

// WithLogger returns a new context with the given Logger.
func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, l)
}

// Entering aids in method tracing. It adds the method name (and optional args)
// to the current Marker in the context, logs an "Entering" message at LevelTrace,
// and returns the updated context.
//
// The returned context contains both the new Marker and the Logger annotated
// with that marker.
//
// Usage:
//
//	func MyFunc(ctx context.Context) {
//	    ctx = logger.Entering(ctx, "MyFunc", "arg", 1)
//	    defer logger.Exiting(ctx)
//
//	    log := logger.LoggerFromContext(ctx)
//	    log.Info("Doing work...")
//	}
func Entering(ctx context.Context, method string, args ...any) context.Context {
	ctx, _ = EnteringWithLogger(ctx, method, args...)
	return ctx
}

func EnteringWithLogger(ctx context.Context, method string, args ...any) (context.Context, *slog.Logger) {
	parentMarker := FromContext(ctx)
	newMarker := parentMarker.AddMethod(method, args...)

	// Get the current logger from context (or default)
	parentLogger := LoggerFromContext(ctx)

	// Create a new logger with the updated marker
	childLogger := parentLogger.With("marker", newMarker)

	// Update context with BOTH the new marker and the new logger
	newCtx := WithMarker(ctx, newMarker)
	newCtx = WithLogger(newCtx, childLogger)

	// Log "Entering" using the CHILD logger (so it has the marker)
	if childLogger.Enabled(newCtx, LevelTrace) {
		childLogger.Log(newCtx, LevelTrace, "Entering "+method)
	}

	return newCtx, childLogger
}

// Exiting logs an "Exiting" message at LevelTrace using the logger found in the context.
// It is intended to be used with defer.
func Exiting(ctx context.Context) {
	l := LoggerFromContext(ctx)
	if l.Enabled(ctx, LevelTrace) {
		l.Log(ctx, LevelTrace, "Exiting")
	}
}
