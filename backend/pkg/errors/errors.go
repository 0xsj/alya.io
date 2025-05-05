// pkg/errors/errors.go
package errors

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"time"

	"maps"

	"github.com/0xsj/alya.io/backend/pkg/logger"
)

// Standard error types
var (
	ErrInvalidInput     = errors.New("invalid input")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrNotFound         = errors.New("resource not found")
	ErrInternalServer   = errors.New("internal server error")
	ErrDuplicateEntry   = errors.New("duplicate entry")
	ErrValidationFailed = errors.New("validation failed")
	ErrDatabase         = errors.New("database error")
	ErrExternalService  = errors.New("external service error")
	
	// YouTube summary specific errors
	ErrYouTubeAPI       = errors.New("youtube API error")
	ErrTranscription    = errors.New("transcription error")
	ErrAIProcessing     = errors.New("AI processing error")
	ErrRateLimited      = errors.New("rate limited")
	ErrInvalidURL       = errors.New("invalid URL format")
	ErrVideoUnavailable = errors.New("video unavailable")
	ErrProcessingFailed = errors.New("processing failed")
)

// AppError represents an application error with additional context
type AppError struct {
	Err        error                  // Original error
	Message    string                 // Human-readable error message
	Code       string                 // Error code for API responses
	Status     int                    // HTTP status code
	LogLevel   int                    // Log level for this error
	StackTrace string                 // Stack trace when the error occurred
	Fields     map[string]any		  // Additional context fields
	Timestamp  time.Time              // Time when the error occurred
	Operation  string                 // Operation that failed (function name, API endpoint, etc.)
}

// Error returns the error message
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks if this error matches the target error
func (e *AppError) Is(target error) bool {
	if e.Err == nil {
		return false
	}
	return errors.Is(e.Err, target)
}

// WithField adds a context field to the error
func (e *AppError) WithField(key string, value any) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	e.Fields[key] = value
	return e
}

// WithFields adds multiple context fields to the error
func (e *AppError) WithFields(fields map[string]any) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]any)
	}
	maps.Copy(e.Fields, fields)
	return e
}

// WithOperation adds the operation name to the error
func (e *AppError) WithOperation(operation string) *AppError {
	e.Operation = operation
	return e
}

// Log logs the error using the provided logger
func (e *AppError) Log(log logger.Logger) {
	// Create contextual logger with error fields
	contextLogger := log
	if e.Fields != nil {
		contextLogger = log.WithFields(e.Fields)
	}

	if e.Operation != "" {
		contextLogger = contextLogger.With("operation", e.Operation)
	}

	// Create error message
	errMsg := fmt.Sprintf("Error: %s (Code: %s, Status: %d)",
		e.Message, e.Code, e.Status)

	if e.Err != nil {
		errMsg = fmt.Sprintf("%s, Cause: %v", errMsg, e.Err)
	}

	// Log with appropriate level
	switch e.LogLevel {
	case logger.DebugLevel:
		if e.StackTrace != "" {
			contextLogger.WithStackTrace().Debug(errMsg)
		} else {
			contextLogger.Debug(errMsg)
		}
	case logger.InfoLevel:
		contextLogger.Info(errMsg)
	case logger.WarnLevel:
		contextLogger.Warn(errMsg)
	case logger.ErrorLevel:
		if e.StackTrace != "" {
			contextLogger.WithStackTrace().Error(errMsg)
		} else {
			contextLogger.Error(errMsg)
		}
	case logger.FatalLevel:
		if e.StackTrace != "" {
			contextLogger.WithStackTrace().Fatal(errMsg)
		} else {
			contextLogger.Fatal(errMsg)
		}
	default:
		if e.StackTrace != "" {
			contextLogger.WithStackTrace().Error(errMsg)
		} else {
			contextLogger.Error(errMsg)
		}
	}
}

// captureStackTrace captures the current stack trace
func captureStackTrace(skip int, depth int) string {
	var pcs [32]uintptr
	n := runtime.Callers(skip+1, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	
	var builder strings.Builder
	i := 0
	for {
		if i >= depth {
			break
		}
		
		frame, more := frames.Next()
		fmt.Fprintf(&builder, "\n    %s\n\t%s:%d", frame.Function, frame.File, frame.Line)
		
		if !more {
			break
		}
		i++
	}
	return builder.String()
}

// Common helper function for creating new errors
func newError(err error, message string, code string, status int, logLevel int) *AppError {
	return &AppError{
		Err:        err,
		Message:    message,
		Code:       code,
		Status:     status,
		LogLevel:   logLevel,
		StackTrace: captureStackTrace(3, 10),
		Fields:     make(map[string]interface{}),
		Timestamp:  time.Now(),
	}
}


// NewBadRequestError creates a new bad request error
func NewBadRequestError(message string, err error) *AppError {
	return newError(err, message, "BAD_REQUEST", http.StatusBadRequest, logger.WarnLevel)
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string, err error) *AppError {
	return newError(err, message, "UNAUTHORIZED", http.StatusUnauthorized, logger.WarnLevel)
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(message string, err error) *AppError {
	return newError(err, message, "FORBIDDEN", http.StatusForbidden, logger.WarnLevel)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(message string, err error) *AppError {
	return newError(err, message, "NOT_FOUND", http.StatusNotFound, logger.InfoLevel)
}

// NewConflictError creates a new conflict error
func NewConflictError(message string, err error) *AppError {
	return newError(err, message, "CONFLICT", http.StatusConflict, logger.WarnLevel)
}

// NewInternalError creates a new internal server error
func NewInternalError(message string, err error) *AppError {
	return newError(err, message, "INTERNAL_SERVER_ERROR", http.StatusInternalServerError, logger.ErrorLevel)
}

// NewValidationError creates a new validation error
func NewValidationError(message string, err error) *AppError {
	return newError(err, message, "VALIDATION_ERROR", http.StatusBadRequest, logger.InfoLevel)
}

// NewDatabaseError creates a new database error
func NewDatabaseError(message string, err error) *AppError {
	return newError(err, message, "DATABASE_ERROR", http.StatusInternalServerError, logger.ErrorLevel)
}

// NewExternalServiceError creates a new external service error
func NewExternalServiceError(message string, err error) *AppError {
	return newError(err, message, "EXTERNAL_SERVICE_ERROR", http.StatusInternalServerError, logger.ErrorLevel)
}

// YouTube summary specific error creators

// NewYouTubeAPIError creates a new YouTube API error
func NewYouTubeAPIError(message string, err error) *AppError {
	return newError(err, message, "YOUTUBE_API_ERROR", http.StatusBadGateway, logger.ErrorLevel)
}

// NewTranscriptionError creates a new transcription error
func NewTranscriptionError(message string, err error) *AppError {
	return newError(err, message, "TRANSCRIPTION_ERROR", http.StatusInternalServerError, logger.ErrorLevel)
}

// NewAIProcessingError creates a new AI processing error
func NewAIProcessingError(message string, err error) *AppError {
	return newError(err, message, "AI_PROCESSING_ERROR", http.StatusInternalServerError, logger.ErrorLevel)
}

// NewRateLimitedError creates a new rate limit error
func NewRateLimitedError(message string, err error) *AppError {
	return newError(err, message, "RATE_LIMITED", http.StatusTooManyRequests, logger.WarnLevel)
}

// NewInvalidURLError creates a new invalid URL error
func NewInvalidURLError(message string, err error) *AppError {
	return newError(err, message, "INVALID_URL", http.StatusBadRequest, logger.InfoLevel)
}

// NewVideoUnavailableError creates a new video unavailable error
func NewVideoUnavailableError(message string, err error) *AppError {
	return newError(err, message, "VIDEO_UNAVAILABLE", http.StatusNotFound, logger.InfoLevel)
}

// NewProcessingFailedError creates a new processing failed error
func NewProcessingFailedError(message string, err error) *AppError {
	return newError(err, message, "PROCESSING_FAILED", http.StatusInternalServerError, logger.ErrorLevel)
}

// Wrap wraps an error with a message
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}

	var appErr *AppError
	if errors.As(err, &appErr) {
		if message != "" {
			appErr.Message = message + ": " + appErr.Message
		}
		return appErr
	}

	return NewInternalError(message, err)
}

// WrapWith wraps an error with a message and error type
func WrapWith(err error, message string, errType *AppError) error {
	if err == nil {
		return nil
	}

	return &AppError{
		Err:        err,
		Message:    message,
		Code:       errType.Code,
		Status:     errType.Status,
		LogLevel:   errType.LogLevel,
		StackTrace: captureStackTrace(2, 10),
		Fields:     make(map[string]interface{}),
		Timestamp:  time.Now(),
	}
}


// IsNotFound checks if an error is a Not Found error
func IsNotFound(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "NOT_FOUND"
}

// IsConflict checks if an error is a Conflict error
func IsConflict(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "CONFLICT"
}

// IsValidationError checks if an error is a Validation error
func IsValidationError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "VALIDATION_ERROR"
}

// IsUnauthorized checks if an error is an Unauthorized error
func IsUnauthorized(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "UNAUTHORIZED"
}

// IsForbidden checks if an error is a Forbidden error
func IsForbidden(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "FORBIDDEN"
}

// IsRateLimited checks if an error is a Rate Limited error
func IsRateLimited(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "RATE_LIMITED"
}

// IsYouTubeAPIError checks if an error is a YouTube API error
func IsYouTubeAPIError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "YOUTUBE_API_ERROR"
}

// IsTranscriptionError checks if an error is a Transcription error
func IsTranscriptionError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "TRANSCRIPTION_ERROR"
}

// IsAIProcessingError checks if an error is an AI Processing error
func IsAIProcessingError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "AI_PROCESSING_ERROR"
}

// IsProcessingFailedError checks if an error is a Processing Failed error
func IsProcessingFailedError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.Code == "PROCESSING_FAILED"
}