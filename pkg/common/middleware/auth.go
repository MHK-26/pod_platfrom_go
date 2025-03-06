// pkg/common/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/your-username/podcast-platform/pkg/auth/usecase"
	"github.com/your-username/podcast-platform/pkg/common/utils"
)

// AuthMiddleware is a middleware for authenticating requests
func AuthMiddleware(authUsecase usecase.Usecase) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondWithError(c, http.StatusUnauthorized, "Authorization header is required")
			c.Abort()
			return
		}

		// Check if the Authorization header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.RespondWithError(c, http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
			c.Abort()
			return
		}

		// Extract the token
		tokenString := parts[1]

		// Verify the token
		payload, err := authUsecase.VerifyToken(c.Request.Context(), tokenString)
		if err != nil {
			utils.RespondWithError(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user data in context
		c.Set("user_id", payload.UserID.String())
		c.Set("email", payload.Email)
		c.Set("user_type", payload.UserType)

		c.Next()
	}
}

// RoleMiddleware checks if the user has the required role
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user type from context (set by AuthMiddleware)
		userType, exists := c.Get("user_type")
		if !exists {
			utils.RespondWithError(c, http.StatusUnauthorized, "Unauthorized")
			c.Abort()
			return
		}

		// Check if user has the required role
		hasRole := false
		for _, role := range roles {
			if userType.(string) == role {
				hasRole = true
				break
			}
		}

		if !hasRole {
			utils.RespondWithError(c, http.StatusForbidden, "Insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}