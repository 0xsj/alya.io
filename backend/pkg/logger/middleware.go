// pkg/logger/middleware.go
package logger

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type ctxKey int

const (
	loggerKey ctxKey = iota
	requestIDKey
)

// HTTPMiddleware is a middleware that logs HTTP requests
func HTTPMiddleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			
			start := time.Now()
			
			// Create request-specific logger
			reqLogger := logger.With("request_id", requestID).
				With("method", r.Method).
				With("path", r.URL.Path).
				With("remote_addr", r.RemoteAddr).
				With("user_agent", r.UserAgent())
			
			// Store logger in context
			ctx := context.WithValue(r.Context(), loggerKey, reqLogger)
			ctx = context.WithValue(ctx, requestIDKey, requestID)
			r = r.WithContext(ctx)
			
			// Add request ID to response headers
			w.Header().Set("X-Request-ID", requestID)
			
			// Create response wrapper to capture status code
			ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
			
			reqLogger.Infof("Request started: %s %s", r.Method, r.URL.Path)
			
			// Execute the handler
			next.ServeHTTP(ww, r)
			
			// Calculate duration
			duration := time.Since(start)
			
			// Log request completion
			reqLogger.WithFields(map[string]interface{}{
				"status":       ww.statusCode,
				"duration_ms":  duration.Milliseconds(),
				"content_type": w.Header().Get("Content-Type"),
			}).Infof("Request completed: %s %s", r.Method, r.URL.Path)
		})
	}
}

// responseWriter is a wrapper around http.ResponseWriter to capture the status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write captures the status code if not already set
func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

// FromContext retrieves the logger from the context
func FromContext(ctx context.Context) Logger {
	logger, ok := ctx.Value(loggerKey).(Logger)
	if !ok {
		return Default()
	}
	return logger
}

// RequestIDFromContext retrieves the request ID from the context
func RequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}
	return requestID
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// Simple implementation - in production, use a more robust method
	return fmt.Sprintf("%d", time.Now().UnixNano())
}