package logger

import (
	"log/slog"
	"os"
)

var (
	// Default logger instance
	defaultLogger *slog.Logger
)

func init() {
	InitLogger()
}

// InitLogger initializes the structured logger
func InitLogger() {
	opts := &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	defaultLogger = slog.New(handler)
}

// Info logs informational messages
func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

// Warn logs warning messages
func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

// Error logs error messages
func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

// Debug logs debug messages
func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

// With returns a logger with added attributes
func With(args ...any) *slog.Logger {
	return defaultLogger.With(args...)
}

// Default returns the default logger
func Default() *slog.Logger {
	return defaultLogger
}
