// internal/api/middleware/auth.go
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/0xsj/alya.io/backend/pkg/logger"
	"github.com/0xsj/alya.io/backend/pkg/response"
)

type AuthMiddleware struct {
	logger logger.Logger
}

func NewAuthMiddleware(logger logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		logger: logger.WithLayer("middleware.auth"),
	}
}

// Authenticate wraps an HTTP handler with authentication
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.logger.Warn("Missing Authorization header")
			response.Error(w, response.ErrUnauthorizedResponse)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			m.logger.Warn("Invalid authorization format:", authHeader)
			response.Error(w, response.ErrUnauthorizedResponse, "Invalid authorization format")
			return
		}

		token := parts[1]
		
		// For development/testing purposes, accept any token with special handling for test tokens
		var userID string
		
		// For testing, you can use "Bearer test-user-1", "Bearer test-user-2", etc.
		if strings.HasPrefix(token, "test-user-") {
			userID = token
			m.logger.Debug("Using test user ID:", userID)
		} else if token == "dev-token" {
			// Default development token
			userID = "dev-user-id"
			m.logger.Debug("Using default development user ID")
		} else {
			// In a real implementation, you'd validate the token here
			// For now, just accept any token for testing
			userID = "user-" + token[:8] // Use first 8 chars of token as user ID
			m.logger.Debug("Created user ID from token:", userID)
		}
		
		// Add the user ID to the context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		
		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}