// Package logger provides structured logging using Go's slog package.
// It supports JSON and text output formats, multiple log levels, and automatic
// trace ID injection for request correlation.
//
// Features:
//   - Structured logging with key-value pairs
//   - JSON and text output formats
//   - Multiple log levels (debug, info, warn, error)
//   - Context-aware logging with trace ID support
//   - Automatic source file and line number tracking
//   - Backward compatible with legacy logger API
//
// Basic Usage:
//
//	logger.Init(logger.Config{
//	    Level:  "info",
//	    Format: "json",
//	})
//	logger.Info("Server started", "port", 8080)
//	logger.Error("Failed to connect", "error", err)
//
// Context-aware Usage (with trace ID):
//
//	logger.InfoContext(ctx, "Processing request", "user_id", 123)
//	logger.ErrorContext(ctx, "Request failed", "error", err)
//
// The Context-aware functions automatically include trace ID if present in context.
package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey string

const traceIDKey contextKey = "trace_id"

var (
	// defaultLogger is the package-level logger instance
	defaultLogger *slog.Logger
)

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
}

// Init initializes the logger with configuration
func Init(cfg Config) {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: true, // Add source file and line number
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)

	defaultLogger.Info("Logger initialized", "level", cfg.Level, "format", cfg.Format)
}

// WithTraceID adds trace ID to context for logging
func WithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// FromContext creates a logger from context with trace ID if available
func FromContext(ctx context.Context) *slog.Logger {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		return defaultLogger.With("trace_id", traceID)
	}
	return defaultLogger
}

// Default returns the default logger instance
func Default() *slog.Logger {
	if defaultLogger == nil {
		// Fallback to default if Init was not called
		defaultLogger = slog.Default()
	}
	return defaultLogger
}

// Compatibility layer for existing code

// LogInfo logs an informational message (legacy compatibility)
func LogInfo(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// LogWarning logs a warning message (legacy compatibility)
func LogWarning(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// LogError logs an error message (legacy compatibility)
func LogError(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// Structured logging helpers

// Info logs an info message with structured attributes
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// InfoContext logs an info message with context
func InfoContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Info(msg, args...)
}

// Warn logs a warning message with structured attributes
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// WarnContext logs a warning message with context
func WarnContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Warn(msg, args...)
}

// Error logs an error message with structured attributes
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// ErrorContext logs an error message with context
func ErrorContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Error(msg, args...)
}

// Debug logs a debug message with structured attributes
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// DebugContext logs a debug message with context
func DebugContext(ctx context.Context, msg string, args ...any) {
	FromContext(ctx).Debug(msg, args...)
}
