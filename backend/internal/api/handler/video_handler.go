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
	case r.Method == http.MethodGet && strings.HasPrefix(path, "/api/v1/videos/") && !strings.Contains(path[14:], "/"):
		h.GetVideo(w, r)
	case r.Method == http.MethodGet && path == "/api/v1/videos":
		h.ListVideos(w, r)
	case r.Method == http.MethodGet && path == "/api/v1/videos/search":
		h.SearchVideos(w, r)
	case r.Method == http.MethodDelete && strings.HasPrefix(path, "/api/v1/videos/"):
		h.DeleteVideo(w, r)
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
		h.logger.Error("Failed to decode request body:", err)
		response.Error(w, response.ErrBadRequestResponse, "Invalid request body: "+err.Error())
		return
	}

	if req.URL == "" {
		response.Error(w, response.ErrBadRequestResponse, "URL is required")
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

// ListVideos handles the request to list videos
func (h *VideoHandler) ListVideos(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	q := r.URL.Query()
	
	page := 1
	pageSize := 20
	
	// Parse pagination params
	if pageStr := q.Get("page"); pageStr != "" {
		if n, err := parsePositiveInt(pageStr); err == nil {
			page = n
		}
	}
	
	if pageSizeStr := q.Get("page_size"); pageSizeStr != "" {
		if n, err := parsePositiveInt(pageSizeStr); err == nil && n <= 100 {
			pageSize = n
		}
	}
	
	// Get user ID from context (added by auth middleware)
	userID := r.Context().Value("user_id").(string)
	
	// Build filters
	filters := map[string]any{}
	
	// Only show videos created by this user or public videos
	filters["created_by"] = userID

	// Execute search with service
	videos, total, err := h.service.SearchVideos("", page, pageSize, userID)
	if err != nil {
		h.logger.Error("Failed to list videos:", err)
		response.HandleError(w, err, h.logger)
		return
	}
	
	// Return paginated response
	meta := response.PaginationMeta{
		CurrentPage:  page,
		PerPage:      pageSize,
		TotalRecords: total,
		TotalPages:   (total + pageSize - 1) / pageSize,
	}
	
	response.WithPagination(w, videos, meta)
}

// SearchVideos handles the request to search videos
func (h *VideoHandler) SearchVideos(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	q := r.URL.Query()
	
	searchQuery := q.Get("q")
	if searchQuery == "" {
		response.Error(w, response.ErrBadRequestResponse, "Search query is required")
		return
	}
	
	page := 1
	pageSize := 20
	
	// Parse pagination params
	if pageStr := q.Get("page"); pageStr != "" {
		if n, err := parsePositiveInt(pageStr); err == nil {
			page = n
		}
	}
	
	if pageSizeStr := q.Get("page_size"); pageSizeStr != "" {
		if n, err := parsePositiveInt(pageSizeStr); err == nil && n <= 100 {
			pageSize = n
		}
	}
	
	// Get user ID from context (added by auth middleware)
	userID := r.Context().Value("user_id").(string)
	
	// Execute search with service
	videos, total, err := h.service.SearchVideos(searchQuery, page, pageSize, userID)
	if err != nil {
		h.logger.Error("Failed to search videos:", err)
		response.HandleError(w, err, h.logger)
		return
	}
	
	// Return paginated response
	meta := response.PaginationMeta{
		CurrentPage:  page,
		PerPage:      pageSize,
		TotalRecords: total,
		TotalPages:   (total + pageSize - 1) / pageSize,
	}
	
	response.WithPagination(w, videos, meta)
}

// DeleteVideo handles the request to delete a video
func (h *VideoHandler) DeleteVideo(w http.ResponseWriter, r *http.Request) {
	// Get video ID from URL path
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

	// Delete the video
	err := h.service.DeleteVideo(videoID, userID)
	if err != nil {
		h.logger.Error("Failed to delete video:", err)
		response.HandleError(w, err, h.logger)
		return
	}

	// Return success response
	response.NoContent(w)
}

// Helper function to parse positive integers
func parsePositiveInt(s string) (int, error) {
	n, err := json.Number(s).Int64()
	if err != nil || n < 1 {
		return 0, err
	}
	return int(n), nil
}