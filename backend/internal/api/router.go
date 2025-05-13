// internal/api/router.go
package api

import (
	"net/http"
	"strings"

	"github.com/0xsj/alya.io/backend/internal/api/handler"
	"github.com/0xsj/alya.io/backend/internal/api/middleware"
	"github.com/0xsj/alya.io/backend/pkg/logger"
)

// Router handles HTTP requests
type Router struct {
	videoHandler  *handler.VideoHandler
	authMiddleware *middleware.AuthMiddleware
	logger        logger.Logger
}

// NewRouter creates a new HTTP router
func NewRouter(
	videoHandler *handler.VideoHandler,
	authMiddleware *middleware.AuthMiddleware,
	logger logger.Logger,
) *Router {
	return &Router{
		videoHandler:  videoHandler,
		authMiddleware: authMiddleware,
		logger:        logger.WithLayer("router"),
	}
}

// ServeHTTP implements the http.Handler interface
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add request ID to the response headers
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}
	w.Header().Set("X-Request-ID", requestID)

	// Log the request
	rt.logger.WithFields(map[string]any{
		"method":      r.Method,
		"path":        r.URL.Path,
		"remote_addr": r.RemoteAddr,
		"request_id":  requestID,
	}).Info("Request received")

	// Route the request based on the path
	path := r.URL.Path

	// Health check endpoint (public)
	if path == "/health" && r.Method == http.MethodGet {
		w.Write([]byte("OK"))
		return
	}

	// API routes (protected by auth middleware)
	if strings.HasPrefix(path, "/api/v1/") {
		// Use auth middleware
		rt.authMiddleware.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rt.routeAPIRequest(w, r)
		})).ServeHTTP(w, r)
		return
	}

	// Default: 404 Not Found
	http.NotFound(w, r)
}

// routeAPIRequest routes API requests to the appropriate handler
func (rt *Router) routeAPIRequest(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// Video routes
	if strings.HasPrefix(path, "/api/v1/videos") {
		rt.videoHandler.ServeHTTP(w, r)
		return
	}

	// Default: 404 Not Found
	http.NotFound(w, r)
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	// In a real implementation, use a proper ID generation method
	return "req-" + randomString(8)
}

// randomString generates a random string of the specified length
func randomString(length int) string {
	// In a real implementation, use crypto/rand
	return "random123"
}