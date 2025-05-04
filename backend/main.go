// examples/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/0xsj/alya.io/backend/pkg/logger"
)

func main() {
	// Create a custom logger
	log := logger.New(logger.Config{
		Level:        logger.DebugLevel,
		EnableJSON:   false,
		EnableTime:   true,
		EnableCaller: true,
		CallerSkip:   1,
		CallerDepth:  10,
	})
	
	// Basic logging
	log.Info("Starting application")
	
	// Logging with fields
	log.WithFields(map[string]interface{}{
		"version": "1.0.0",
		"env":     "development",
	}).Info("Application initialized")
	
	// Logging with layers
	dbLogger := log.WithLayer("database")
	dbLogger.Info("Database connection established")
	
	// Logging with stack trace
	if err := someFunction(); err != nil {
		log.WithStackTrace().Error("Error occurred:", err)
	}
	
	// Timing a function
	timer := log.Timer("process-data")
	processData()
	timer.Stop()
	
	// Using context
	ctx := context.Background()
	ctx = context.WithValue(ctx, logger.LoggerKey, log)
	
	// Tracing a function
	traceFunc(ctx)
	
	// HTTP server with logging middleware
	http.Handle("/", logger.HTTPMiddleware(log)(http.HandlerFunc(handler)))
	http.ListenAndServe(":8080", nil)
}

func someFunction() error {
	// Simulate an error
	return fmt.Errorf("something went wrong")
}

func processData() {
	// Simulate processing
	time.Sleep(100 * time.Millisecond)
}

func traceFunc(ctx context.Context) {
	ctx, done := logger.FunctionTracer(ctx, "traceFunc")
	defer done()
	
	// Do something
	time.Sleep(50 * time.Millisecond)
	
	// Nested tracing
	nestedFunc(ctx)
}

func nestedFunc(ctx context.Context) {
	defer logger.Trace(ctx, "nestedFunc")()
	
	// Do something else
	time.Sleep(25 * time.Millisecond)
}

func handler(w http.ResponseWriter, r *http.Request) {
	// Get logger from context
	log := logger.FromContext(r.Context())
	log.Info("Handling request")
	
	// Do something
	time.Sleep(200 * time.Millisecond)
	
	// Response
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, world!"))
}