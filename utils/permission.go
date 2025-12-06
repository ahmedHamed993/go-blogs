package utils

import (
	"github/ahmedhamed993/go-auth/models"

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

func GetRolePermissions(db *gorm.DB, roleID uint) ([]string, error) {
	var role models.Role

	err := db.
		Preload("Permissions").
		First(&role, roleID).Error

	if err != nil {
		return nil, err
	}

	// Convert to string slice
	perms := make([]string, 0)
	for _, p := range role.Permissions {
		perms = append(perms, p.Name)
	}

	return perms, nil
}
