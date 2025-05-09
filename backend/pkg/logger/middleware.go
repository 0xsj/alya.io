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
	LoggerKey ctxKey = iota
	requestIDKey
)

func HTTPMiddleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}
			
			start := time.Now()
			reqLogger := logger.With("request_id", requestID).
				With("method", r.Method).
				With("path", r.URL.Path).
				With("remote_addr", r.RemoteAddr).
				With("user_agent", r.UserAgent())
		
			ctx := context.WithValue(r.Context(), LoggerKey, reqLogger)
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
			
			reqLogger.WithFields(map[string]any{
				"status":       ww.statusCode,
				"duration_ms":  duration.Milliseconds(),
				"content_type": w.Header().Get("Content-Type"),
			}).Infof("Request completed: %s %s", r.Method, r.URL.Path)
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if rw.statusCode == 0 {
		rw.statusCode = http.StatusOK
	}
	return rw.ResponseWriter.Write(b)
}

func FromContext(ctx context.Context) Logger {
	logger, ok := ctx.Value(LoggerKey).(Logger)
	if !ok {
		return Default()
	}
	return logger
}

func RequestIDFromContext(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}
	return requestID
}

func generateRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}