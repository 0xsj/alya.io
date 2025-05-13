// internal/api/handler/video_handler.go
package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/0xsj/alya.io/backend/internal/domain"
	"github.com/0xsj/alya.io/backend/pkg/logger"
	"github.com/0xsj/alya.io/backend/pkg/response"
)

type VideoHandler struct {
	service domain.VideoService
	logger  logger.Logger
}

func NewVideoHandler(service domain.VideoService, logger logger.Logger) *VideoHandler {
	return &VideoHandler{
		service: service,
		logger:  logger.WithLayer("handler.video"),
	}
}

// ServeHTTP implements the http.Handler interface
func (h *VideoHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Parse the path
	path := r.URL.Path
	
	// Handle different routes based on path and method
	switch {
	case r.Method == http.MethodPost && path == "/api/v1/videos":
		h.ProcessVideo(w, r)
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v1/videos/"):
		h.GetVideo(w, r)
	default:
		// Return 404 for unknown routes
		http.NotFound(w, r)
	}
}

// ProcessVideo handles the request to process a new YouTube video
func (h *VideoHandler) ProcessVideo(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, response.ErrBadRequestResponse, "Invalid request body")
		return
	}

	// Get user ID from context (added by auth middleware)
	userID := r.Context().Value("user_id").(string)

	// Process the video
	video, err := h.service.ProcessVideo(req.URL, userID)
	if err != nil {
		h.logger.Error("Failed to process video:", err)
		response.HandleError(w, err, h.logger)
		return
	}

	// Return the video details
	response.Created(w, video, "Video processing started")
}

// GetVideo handles the request to get a video by ID
func (h *VideoHandler) GetVideo(w http.ResponseWriter, r *http.Request) {
	// Get video ID from URL path
	// The path is expected to be /api/v1/videos/{id}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		response.Error(w, response.ErrBadRequestResponse, "Video ID is required")
		return
	}
	
	videoID := parts[len(parts)-1]
	if videoID == "" {
		response.Error(w, response.ErrBadRequestResponse, "Video ID is required")
		return
	}

	// Get user ID from context (added by auth middleware)
	userID := r.Context().Value("user_id").(string)

	// Get the video
	video, err := h.service.GetVideoDetails(videoID, userID)
	if err != nil {
		h.logger.Error("Failed to get video:", err)
		response.HandleError(w, err, h.logger)
		return
	}

	// Return the video details
	response.Success(w, video, "")
}