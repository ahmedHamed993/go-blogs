package middlewares

import (
	"github/ahmedhamed993/go-auth/services"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	ContextKeyPermissions      = "permissions"
	ContextKeyUserID           = "user_id"
	ContextKeyRoleID           = "role_id"
	ContextKeyPermissionScope  = "permission_scope"  // "own", "all", or "" (base permission)
	ContextKeyNeedsOwnershipCheck = "needs_ownership_check" // true if user has :own permission
)

// loadUserPermissions loads permissions for the authenticated user and stores them in context
func loadUserPermissions(c *gin.Context) ([]string, error) {
	// Check if permissions are already loaded in context
	if perms, exists := c.Get(ContextKeyPermissions); exists {
		if permissions, ok := perms.([]string); ok {
			return permissions, nil
		}
	}

	// Get role_id from context (set by AuthMiddleware)
	roleID, exists := c.Get(ContextKeyRoleID)
	if !exists {
		return nil, gin.Error{Err: gin.Error{Err: nil}, Meta: "role_id not found in context"}
	}

	roleIDUint, ok := roleID.(uint)
	if !ok {
		return nil, gin.Error{Err: gin.Error{Err: nil}, Meta: "invalid role_id type"}
	}

	// Load permissions from database
	permissions, err := services.LoadUserPermissions(roleIDUint)
	if err != nil {
		return nil, err
	}

	// Store in context for future use
	c.Set(ContextKeyPermissions, permissions)

	return permissions, nil
}

// RequirePermission middleware handles authorization with scoped permissions
// - If permission is public (:public suffix): allows access without authentication
// - If permission is not public: requires authentication and checks permissions
// - Sets scope in context if determinable (own, all, or public)
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if permission is public first
		if services.CheckPublicPermission(permission) {
			// Public permissions don't require authentication
			// Set scope and allow access
			c.Set(ContextKeyPermissionScope, "public")
			c.Set(ContextKeyNeedsOwnershipCheck, false)
			c.Next()
			return
		}

		// For non-public permissions, require authentication
		// Check if user is authenticated (user_id must be in context)
		_, exists := c.Get(ContextKeyUserID)
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication required",
			})
			c.Abort()
			return
		}

		// Load user permissions
		userPermissions, err := loadUserPermissions(c)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to load permissions",
			})
			c.Abort()
			return
		}

		// Check scoped permission
		hasPermission, isPublic, needsOwnershipCheck := services.CheckScopedPermission(userPermissions, permission)
		
		if isPublic {
			// Public permission, allow access and set scope
			c.Set(ContextKeyPermissionScope, "public")
			c.Set(ContextKeyNeedsOwnershipCheck, false)
			c.Next()
			return
		}

		if !hasPermission {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Insufficient permissions",
			})
			c.Abort()
			return
		}

		// Determine the actual scope from user's permissions
		// The scope indicates what the user actually has: "all", "own", or "public"
		var scope string
		parts := strings.Split(permission, ":")
		basePermission := ""
		if len(parts) >= 2 {
			basePermission = parts[0] + ":" + parts[1]
		}

		// Check user's permissions to determine their actual scope
		for _, userPerm := range userPermissions {
			// Check for exact match first
			if userPerm == permission {
				if strings.HasSuffix(userPerm, ":public") {
					scope = "public"
					break
				} else if strings.HasSuffix(userPerm, ":all") {
					scope = "all"
					break
				} else if strings.HasSuffix(userPerm, ":own") {
					scope = "own"
					break
				}
			}
			
			// Check for base permission match (e.g., user has "users:read" when checking "users:read:own")
			if basePermission != "" && userPerm == basePermission {
				scope = "all" // Base permission grants access to all
				break
			}
			
			// Check for :all variant (e.g., user has "users:read:all" when checking "users:read")
			if basePermission != "" && userPerm == basePermission+":all" {
				scope = "all"
				break
			}
			
			// Check for :own variant (e.g., user has "users:read:own" when checking "users:read")
			if basePermission != "" && userPerm == basePermission+":own" {
				scope = "own"
				break
			}
		}

		// If scope can be determined, set it in context
		if scope != "" {
			c.Set(ContextKeyPermissionScope, scope)
			c.Set(ContextKeyNeedsOwnershipCheck, needsOwnershipCheck)
		} else if needsOwnershipCheck {
			// If needsOwnershipCheck is true but scope wasn't found, set "own" as scope
			c.Set(ContextKeyPermissionScope, "own")
			c.Set(ContextKeyNeedsOwnershipCheck, needsOwnershipCheck)
		}
		// If scope cannot be determined, don't set anything and just continue

		c.Next()
	}
}


