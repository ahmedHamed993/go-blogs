package services

import (
	"github/ahmedhamed993/go-auth/database"
	"strings"
)

// LoadUserPermissions loads all permissions for a user based on their role
func LoadUserPermissions(roleID uint) ([]string, error) {
	var permissionNames []string

	err := database.DB.
		Table("role_permissions").
		Select("permissions.name").
		Joins("JOIN permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Pluck("permissions.name", &permissionNames).
		Error

	if err != nil {
		return nil, err
	}

	return permissionNames, nil
}

// HasPermission checks if a user has a specific permission
// Returns true if user has the exact permission or a more general permission
func HasPermission(userPermissions []string, requiredPermission string) bool {
	// Check for exact match
	for _, perm := range userPermissions {
		if perm == requiredPermission {
			return true
		}
	}

	// Check for base permission (e.g., if user has "users:read", they can access "users:read:own")
	// Split the required permission to get base (resource:action)
	parts := strings.Split(requiredPermission, ":")
	if len(parts) >= 2 {
		basePermission := parts[0] + ":" + parts[1]
		for _, perm := range userPermissions {
			if perm == basePermission {
				return true
			}
		}
	}

	return false
}

// CheckScopedPermission checks if a user has permission with scope handling
// Handles :own, :all, and :public scopes
// Returns: (hasPermission, isPublic, needsOwnershipCheck)
func CheckScopedPermission(userPermissions []string, requiredPermission string) (bool, bool, bool) {
	// Check if it's a public permission
	if strings.HasSuffix(requiredPermission, ":public") {
		// Check if user has the public permission or any variant
		for _, perm := range userPermissions {
			if perm == requiredPermission || strings.HasPrefix(perm, strings.TrimSuffix(requiredPermission, ":public")+":") {
				return true, true, false
			}
		}
		// Public permissions can be accessed even without authentication
		// But we still check if user has it for authenticated users
		return false, true, false
	}

	// Check for exact match
	if HasPermission(userPermissions, requiredPermission) {
		// If it's an :all permission, no ownership check needed
		if strings.HasSuffix(requiredPermission, ":all") {
			return true, false, false
		}
		// If it's an :own permission, ownership check is needed
		if strings.HasSuffix(requiredPermission, ":own") {
			return true, false, true
		}
		// Base permission (e.g., "users:read") allows all
		return true, false, false
	}

	// Check if user has :all variant (e.g., has "blogs:read:all" but needs "blogs:read")
	parts := strings.Split(requiredPermission, ":")
	if len(parts) >= 2 {
		basePermission := parts[0] + ":" + parts[1]
		allPermission := basePermission + ":all"
		
		for _, perm := range userPermissions {
			if perm == allPermission {
				return true, false, false
			}
		}
	}

	// Check if user has :own variant and we're checking for base permission
	// (e.g., user has "users:read:own" and we're checking "users:read" on their own resource)
	if len(parts) >= 2 {
		basePermission := parts[0] + ":" + parts[1]
		ownPermission := basePermission + ":own"
		
		for _, perm := range userPermissions {
			if perm == ownPermission {
				// User has :own permission, but we need to check ownership
				return true, false, true
			}
		}
	}

	return false, false, false
}

// CheckPublicPermission checks if a permission is public (ends with :public)
func CheckPublicPermission(permission string) bool {
	return strings.HasSuffix(permission, ":public")
}

