package utils

import (
	"gorm.io/gorm"
)

// LoadUserPermissions loads all permissions for a user based on their role
func LoadUserPermissions(db *gorm.DB, roleID uint) ([]string, error) {
	var permissionNames []string

	err := db.
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

