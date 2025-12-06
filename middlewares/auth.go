package middlewares

import (
	"github/ahmedhamed993/go-auth/database"
	"github/ahmedhamed993/go-auth/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func setPublic(c *gin.Context) {
	c.Set(ContextKeyUserID, nil)
	c.Set(ContextKeyRoleID, nil)
	c.Set(ContextKeyPermissions, []string{})
	c.Set(ContextKeyScope, "public")
}

func AuthMiddleware(allowPublic bool) gin.HandlerFunc {
	return func(c *gin.Context) {

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			if allowPublic {
				setPublic(c)
				c.Next()
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		// Expect: "Bearer token"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			if allowPublic {
				setPublic(c)
				c.Next()
				return
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization format",
			})
			return
		}

		tokenString := parts[1]

		claims, err := utils.ParseToken(tokenString)
		if err != nil {

			// Public route → treat as guest
			if allowPublic {
				setPublic(c)
				c.Next()
				return
			}

			// Private route → reject
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		permissions, err := utils.GetRolePermissions(database.DB, claims.RoleID)
		if err != nil {
			permissions = []string{}
		}

		// Set authenticated user
		c.Set(ContextKeyUserID, claims.UserID)
		c.Set(ContextKeyRoleID, claims.RoleID)
		c.Set(ContextKeyPermissions, permissions)
		c.Set(ContextKeyScope, "auth")

		c.Next()
	}
}
