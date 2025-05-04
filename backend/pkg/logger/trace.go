// pkg/logger/trace.go
package logger

import (
	"context"
	"net/http"
	"runtime"
	"time"
)

// FunctionTracer traces function execution with timing
func FunctionTracer(ctx context.Context, funcName string) (context.Context, func()) {
	logger := FromContext(ctx)
	layerLogger := logger.WithLayer("function")
	
	// Get caller information
	pc, file, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc)
	if fn != nil {
		funcName = fn.Name()
	}
	
	// Start timer
	timer := layerLogger.Timer(funcName)
	
	// Log entry
	layerLogger.Debugf("ENTER: %s (%s:%d)", funcName, file, line)
	
	// Return cleanup function
	return ctx, func() {
		duration := timer.Stop()
		layerLogger.Debugf("EXIT: %s (%s:%d) took %s", funcName, file, line, duration)
	}
}

// Trace traces the execution of a function with the given name
func Trace(ctx context.Context, name string) func() {
	logger := FromContext(ctx)
	layerLogger := logger.WithLayer("trace")
	
	// Start timer
	start := time.Now()
	
	// Log entry
	layerLogger.Debugf("ENTER: %s", name)
	
	// Return cleanup function
	return func() {
		duration := time.Since(start)
		layerLogger.Debugf("EXIT: %s took %s", name, duration)
	}
}

// TraceHTTPHandler wraps an HTTP handler with tracing
func TraceHTTPHandler(handler http.HandlerFunc, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := FromContext(r.Context())
		layerLogger := logger.WithLayer("http")
		
		// Start timer
		timer := layerLogger.Timer(name)
		
		// Log entry
		layerLogger.Debugf("ENTER: %s", name)
		
		// Call the handler
		handler(w, r)
		
		// Log exit
		duration := timer.Stop()
		layerLogger.Debugf("EXIT: %s took %s", name, duration)
	}
}