package seeders

import "github/ahmedhamed993/go-auth/models"

var Roles = []models.Role{
	{Name: "superadmin", Description: "Full system access"},
	{Name: "admin", Description: "Admin level access"},
	{Name: "user", Description: "Normal user"},
}

var RolePermissionsMap = map[string][]string{
	"superadmin": {
		"users:create", "users:read", "users:update", "users:delete",
		"blogs:create", "blogs:read:all", "blogs:update:all", "blogs:delete:all",
	},
	"admin": {
		"users:read", "users:update",
		"blogs:create", "blogs:read:all", "blogs:update:all",
	},
	"user": {
		"users:read:own", "users:update:own",
		"blogs:create", "blogs:read:own", "blogs:update:own", "blogs:delete:own",
	},
}
