// internal/api/middleware/auth.go
package middleware

import (
	"context"
	"fmt"
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
			response.Error(w, response.ErrUnauthorizedResponse)
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(w, response.ErrUnauthorizedResponse, "Invalid authorization format")
			return
		}

		token := parts[1]

		fmt.Println(token)
		
		// For development, just use a dummy user ID
		// In production, you'd validate the token
		userID := "dev-user-id"
		
		// Add the user ID to the context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		
		// Call the next handler with the updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}