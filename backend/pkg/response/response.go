// pkg/response/response.go
package response

import (
	"encoding/json"
	stdErrors "errors"
	"fmt"
	"net/http"

	"github.com/0xsj/alya.io/backend/pkg/errors"
	"github.com/0xsj/alya.io/backend/pkg/logger"
)

// Response is a standard API response structure
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    any `json:"data,omitempty"`
	Meta    any `json:"meta,omitempty"`
}

// ErrorResponse is a standard API error response structure
type ErrorResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details any			`json:"details,omitempty"`
}

// PaginationMeta holds pagination metadata
type PaginationMeta struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	PerPage      int `json:"per_page"`
	TotalRecords int `json:"total_records"`
}

// Predefined error responses
var (
	ErrBadRequestResponse = ErrorResponse{
		Code:    "BAD_REQUEST",
		Message: "The request was invalid",
	}
	ErrUnauthorizedResponse = ErrorResponse{
		Code:    "UNAUTHORIZED",
		Message: "Authentication is required",
	}
	ErrForbiddenResponse = ErrorResponse{
		Code:    "FORBIDDEN",
		Message: "You don't have permission to access this resource",
	}
	ErrNotFoundResponse = ErrorResponse{
		Code:    "NOT_FOUND",
		Message: "The requested resource was not found",
	}
	ErrConflictResponse = ErrorResponse{
		Code:    "CONFLICT",
		Message: "The resource already exists",
	}
	ErrInternalServerResponse = ErrorResponse{
		Code:    "INTERNAL_SERVER_ERROR",
		Message: "An unexpected error occurred",
	}
	ErrServiceUnavailableResponse = ErrorResponse{
		Code:    "SERVICE_UNAVAILABLE",
		Message: "The service is currently unavailable",
	}
	ErrRateLimitedResponse = ErrorResponse{
		Code:    "RATE_LIMITED",
		Message: "You have exceeded the rate limit",
	}
	ErrYouTubeAPIResponse = ErrorResponse{
		Code:    "YOUTUBE_API_ERROR",
		Message: "Error communicating with YouTube API",
	}
	ErrTranscriptionResponse = ErrorResponse{
		Code:    "TRANSCRIPTION_ERROR",
		Message: "Error processing video transcription",
	}
	ErrAIProcessingResponse = ErrorResponse{
		Code:    "AI_PROCESSING_ERROR",
		Message: "Error processing content with AI",
	}
	ErrInvalidURLResponse = ErrorResponse{
		Code:    "INVALID_URL",
		Message: "The provided URL is invalid or not supported",
	}
	ErrVideoUnavailableResponse = ErrorResponse{
		Code:    "VIDEO_UNAVAILABLE",
		Message: "The requested video is unavailable",
	}
)

// JSON writes a JSON response to the provided ResponseWriter
func JSON(w http.ResponseWriter, statusCode int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	return json.NewEncoder(w).Encode(data)
}

// Success sends a successful response
func Success(w http.ResponseWriter, data any, message string, statusCode ...int) error {
	resp := Response{
		Success: true,
		Data:    data,
	}
	
	if message != "" {
		resp.Message = message
	}
	
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	
	return JSON(w, code, resp)
}

// Created sends a 201 Created response
func Created(w http.ResponseWriter, data any, message string) error {
	return Success(w, data, message, http.StatusCreated)
}

// Accepted sends a 202 Accepted response for async processing
func Accepted(w http.ResponseWriter, data any, message string) error {
	return Success(w, data, message, http.StatusAccepted)
}

// WithPagination sends a paginated response
func WithPagination(w http.ResponseWriter, data any, meta PaginationMeta) error {
	resp := Response{
		Success: true,
		Data:    data,
		Meta:    meta,
	}
	
	return JSON(w, http.StatusOK, resp)
}

// NoContent sends a 204 No Content response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Error sends an error response
func Error(w http.ResponseWriter, err ErrorResponse, details ...any) error {
	if len(details) > 0 {
		err.Details = details[0]
	}
	
	statusCode := http.StatusInternalServerError
	switch err.Code {
	case "BAD_REQUEST", "VALIDATION_ERROR", "INVALID_URL":
		statusCode = http.StatusBadRequest
	case "UNAUTHORIZED":
		statusCode = http.StatusUnauthorized
	case "FORBIDDEN":
		statusCode = http.StatusForbidden
	case "NOT_FOUND", "VIDEO_UNAVAILABLE":
		statusCode = http.StatusNotFound
	case "CONFLICT":
		statusCode = http.StatusConflict
	case "RATE_LIMITED":
		statusCode = http.StatusTooManyRequests
	case "SERVICE_UNAVAILABLE":
		statusCode = http.StatusServiceUnavailable
	case "YOUTUBE_API_ERROR":
		statusCode = http.StatusBadGateway
	}
	
	return JSON(w, statusCode, err)
}

// HandleError processes an error and sends appropriate response
func HandleError(w http.ResponseWriter, err error, log logger.Logger) {
	var appErr *errors.AppError
	if stdErrors.As(err, &appErr) {
		appErr.Log(log)
		JSON(w, appErr.Status, ErrorResponse{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Fields,
		})
		return
	}
	
	// If it's not an AppError, check for standard error types
	switch {
	case stdErrors.Is(err, errors.ErrInvalidInput), stdErrors.Is(err, errors.ErrValidationFailed):
		log.With("error", err.Error()).Warn("Bad request error")
		Error(w, ErrBadRequestResponse, err.Error())
	case stdErrors.Is(err, errors.ErrInvalidURL):
		log.With("error", err.Error()).Info("Invalid URL error")
		Error(w, ErrInvalidURLResponse, err.Error())
	case stdErrors.Is(err, errors.ErrUnauthorized):
		log.With("error", err.Error()).Warn("Unauthorized error")
		Error(w, ErrUnauthorizedResponse)
	case stdErrors.Is(err, errors.ErrForbidden):
		log.With("error", err.Error()).Warn("Forbidden error")
		Error(w, ErrForbiddenResponse)
	case stdErrors.Is(err, errors.ErrNotFound):
		log.With("error", err.Error()).Info("Not found error")
		Error(w, ErrNotFoundResponse)
	case stdErrors.Is(err, errors.ErrVideoUnavailable):
		log.With("error", err.Error()).Info("Video unavailable error")
		Error(w, ErrVideoUnavailableResponse, err.Error())
	case stdErrors.Is(err, errors.ErrDuplicateEntry):
		log.With("error", err.Error()).Warn("Conflict error")
		Error(w, ErrConflictResponse)
	case stdErrors.Is(err, errors.ErrRateLimited):
		log.With("error", err.Error()).Warn("Rate limited error")
		Error(w, ErrRateLimitedResponse)
	case stdErrors.Is(err, errors.ErrYouTubeAPI):
		log.With("error", err.Error()).Error("YouTube API error")
		Error(w, ErrYouTubeAPIResponse, err.Error())
	case stdErrors.Is(err, errors.ErrTranscription):
		log.With("error", err.Error()).Error("Transcription error")
		Error(w, ErrTranscriptionResponse, err.Error())
	case stdErrors.Is(err, errors.ErrAIProcessing):
		log.With("error", err.Error()).Error("AI processing error")
		Error(w, ErrAIProcessingResponse, err.Error())
	case stdErrors.Is(err, errors.ErrDatabase) || stdErrors.Is(err, errors.ErrExternalService):
		log.With("error", err.Error()).Error("Database/external service error")
		Error(w, ErrInternalServerResponse)
	default:
		log.With("error", err.Error()).Error("Unhandled error")
		Error(w, ErrInternalServerResponse)
	}
}

// Stream creates a streaming response for large data
func Stream(w http.ResponseWriter, data []byte, contentType string) error {
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(data)))
	w.WriteHeader(http.StatusOK)
	_, err := w.Write(data)
	return err
}

// Redirect sends a redirect response
func Redirect(w http.ResponseWriter, r *http.Request, url string, statusCode ...int) {
	code := http.StatusFound // 302 by default
	if len(statusCode) > 0 {
		code = statusCode[0]
	}
	http.Redirect(w, r, url, code)
}

// File sends a file response
func File(w http.ResponseWriter, data []byte, filename string, contentType string) error {
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	return Stream(w, data, contentType)
}