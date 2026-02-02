package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	pkgerrors "github.com/linskybing/platform-go/pkg/errors"
)

// PanicRecovery recovers from panics and logs them with stack traces
func PanicRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Log the panic with full stack trace
				slog.Error("panic recovered in handler",
					"panic", r,
					"path", c.Request.URL.Path,
					"method", c.Request.Method,
					"remote_addr", c.ClientIP(),
					"stack", string(debug.Stack()))

				// Return error response
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"message": "An unexpected error occurred",
				})
				c.Abort()
			}
		}()

		c.Next()
	}
}

// ErrorHandler handles errors from handlers and converts them to appropriate HTTP responses
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) == 0 {
			return
		}

		// Get the last error
		err := c.Errors.Last().Err

		// Determine status code and message based on error type
		statusCode := pkgerrors.GetHTTPStatus(err)
		message := err.Error()

		// Log the error with context
		logError(c, err, statusCode)

		// Return JSON error response
		c.JSON(statusCode, gin.H{
			"error":   http.StatusText(statusCode),
			"message": message,
		})
	}
}

func logError(c *gin.Context, err error, statusCode int) {
	fields := []interface{}{
		"error", err,
		"status", statusCode,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
		"remote_addr", c.ClientIP(),
	}

	// Add user info if available
	if username, exists := c.Get("username"); exists {
		fields = append(fields, "username", username)
	}

	// Log based on error severity
	switch {
	case statusCode >= 500:
		slog.Error("internal server error", fields...)
	case statusCode >= 400:
		slog.Warn("client error", fields...)
	default:
		slog.Info("request error", fields...)
	}
}

// SafeGoroutine wraps a goroutine with panic recovery
func SafeGoroutine(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in goroutine",
					"goroutine", name,
					"panic", r,
					"stack", string(debug.Stack()))
			}
		}()
		fn()
	}()
}

// SafeGoroutineWithContext wraps a goroutine with panic recovery and context
func SafeGoroutineWithContext(name string, fn func() error) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic in goroutine",
					"goroutine", name,
					"panic", r,
					"stack", string(debug.Stack()))
			}
		}()

		if err := fn(); err != nil {
			slog.Error("goroutine error",
				"goroutine", name,
				"error", err)
		}
	}()
}

// WrapError wraps an error with context information
func WrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}
