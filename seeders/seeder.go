package seeders

import (
	"github/ahmedhamed993/go-auth/models"
	"github/ahmedhamed993/go-auth/utils"

	"log"

	"gorm.io/gorm"
)

func SeedRBAC(db *gorm.DB) {
	// 1. Insert permissions
	for _, p := range Permissions {
		db.FirstOrCreate(&p, models.Permission{Name: p.Name})
	}

	// 2. Insert roles
	for _, r := range Roles {
		db.FirstOrCreate(&r, models.Role{Name: r.Name})
	}

	// 3. Assign permissions to roles
	for roleName, permNames := range RolePermissionsMap {
		var role models.Role
		db.Where("name = ?", roleName).First(&role)

		for _, permName := range permNames {
			var perm models.Permission
			db.Where("name = ?", permName).First(&perm)

			rp := models.RolePermission{
				RoleID:       role.ID,
				PermissionID: perm.ID,
			}

			db.FirstOrCreate(&rp, rp)
		}
	}

	log.Println("RBAC seeded successfully")
}

// SeedDefaultUsers inserts a default user, admin, and superadmin
func SeedDefaultUsers(db *gorm.DB) {
	var roles []models.Role
	db.Find(&roles)

	roleMap := make(map[string]models.Role)

	var users = []models.User{
		{
			Username:   "superadmin",
			Password:   utils.HashPassword("superadmin123"),
			RoleID:     roleMap["superadmin"].ID,
			IsVerified: true,
		},
		{
			Username:   "admin",
			Password:   utils.HashPassword("admin123"),
			RoleID:     roleMap["admin"].ID,
			IsVerified: true,
		},
		{
			Username:   "user",
			Password:   utils.HashPassword("user123"),
			RoleID:     roleMap["user"].ID,
			IsVerified: true,
		},
	}

	for _, r := range roles {
		roleMap[r.Name] = r
	}

	for _, u := range users {
		db.FirstOrCreate(&u, models.User{Username: u.Username})
	}

	log.Println("✅ Default users seeded successfully")
}
