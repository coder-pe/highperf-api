package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"
)

type Logger struct {
	*slog.Logger
}

type Config struct {
	Level     string `json:"level" envconfig:"LOG_LEVEL" default:"info"`
	Format    string `json:"format" envconfig:"LOG_FORMAT" default:"json"` // json, text
	AddSource bool   `json:"add_source" envconfig:"LOG_ADD_SOURCE" default:"true"`
}

// New creates a new structured logger
func New(cfg Config) *Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.Format == "text" {
		handler = slog.NewTextHandler(os.Stdout, opts)
	} else {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return &Logger{
		Logger: slog.New(handler),
	}
}

// NewForTesting creates a logger that discards output for testing
func NewForTesting() *Logger {
	handler := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return &Logger{
		Logger: slog.New(handler),
	}
}

// WithRequestID adds request ID to logger context
func (l *Logger) WithRequestID(ctx context.Context, requestID string) *Logger {
	return &Logger{
		Logger: l.Logger.With("request_id", requestID),
	}
}

// WithError adds error to logger context
func (l *Logger) WithError(err error) *Logger {
	return &Logger{
		Logger: l.Logger.With("error", err.Error()),
	}
}

// WithFields adds multiple fields to logger context
func (l *Logger) WithFields(fields map[string]any) *Logger {
	args := make([]any, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{
		Logger: l.Logger.With(args...),
	}
}

// HTTPRequest logs HTTP request details
func (l *Logger) HTTPRequest(method, path, userAgent, clientIP string, statusCode int, duration time.Duration, bodySize int64) {
	l.Info("http_request",
		"method", method,
		"path", path,
		"user_agent", userAgent,
		"client_ip", clientIP,
		"status_code", statusCode,
		"duration_ms", duration.Milliseconds(),
		"body_size_bytes", bodySize,
	)
}

// HTTPError logs HTTP error details
func (l *Logger) HTTPError(method, path string, statusCode int, err error, duration time.Duration) {
	l.Error("http_error",
		"method", method,
		"path", path,
		"status_code", statusCode,
		"error", err.Error(),
		"duration_ms", duration.Milliseconds(),
	)
}

// BusinessEvent logs business logic events
func (l *Logger) BusinessEvent(event string, fields map[string]any) {
	args := []any{"event", event}
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.Info("business_event", args...)
}

// Performance logs performance metrics
func (l *Logger) Performance(operation string, duration time.Duration, fields map[string]any) {
	args := []any{
		"operation", operation,
		"duration_ms", duration.Milliseconds(),
	}
	for k, v := range fields {
		args = append(args, k, v)
	}
	l.Info("performance", args...)
}