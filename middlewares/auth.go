package middlewares

import (
	"github/ahmedhamed993/go-auth/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthMiddleware makes authentication optional - it doesn't abort if token is missing
// If a valid token is provided, it sets user context (user_id, role_id, permissions)
// If token is missing or invalid, it continues without setting user context (allows public access)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// If no authorization header, continue without authentication
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.Next()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return services.SecretKey, nil
		})

		// If token is invalid, continue without authentication (don't abort)
		// Authorization middleware will handle requiring auth for protected routes
		if err != nil || !token.Valid {
			c.Next()
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		userID := uint(claims["user_id"].(float64))
		roleID := uint(claims["role_id"].(float64))

		c.Set(ContextKeyUserID, userID)
		c.Set(ContextKeyRoleID, roleID)

		// Preload user permissions for efficient access
		permissions, err := services.LoadUserPermissions(roleID)
		if err == nil {
			c.Set(ContextKeyPermissions, permissions)
		}

		c.Next()
	}
}
