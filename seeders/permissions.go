package seeders

import "github/ahmedhamed993/go-auth/models"

var Permissions = []models.Permission{
	// User permissions
	{Name: "users:create", Description: "Create a user"},
	{Name: "users:read", Description: "Read any user"},
	{Name: "users:read:own", Description: "Read own user profile"},
	{Name: "users:update", Description: "Update any user"},
	{Name: "users:update:own", Description: "Update own user profile"},
	{Name: "users:delete", Description: "Delete any user"},

	// Blog permissions (scoped)
	{Name: "blogs:create", Description: "Create blogs"},
	{Name: "blogs:read:all", Description: "Read all blogs"},
	{Name: "blogs:read:own", Description: "Read own blogs"},
	{Name: "blogs:update:all", Description: "Update any blog"},
	{Name: "blogs:update:own", Description: "Update own blog"},
	{Name: "blogs:delete:all", Description: "Delete any blog"},
	{Name: "blogs:delete:own", Description: "Delete own blog"},
}
