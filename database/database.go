package database

import (
	"fmt"
	"github/ahmedhamed993/go-auth/models"
	"github/ahmedhamed993/go-auth/seeders"

	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := "host=localhost user=postgres password=postgres dbname=app_db port=5433 sslmode=disable"

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}

	fmt.Println("📌 Connected to PostgreSQL successfully!")
	DB = db

	// Auto migrate tables
	err = DB.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
	)
	if err != nil {
		log.Fatalf("❌ DB migration failed: %v", err)
	}

	seeders.SeedRBAC(DB)         // roles + permissions
	seeders.SeedDefaultUsers(DB) // defau

	fmt.Println("📌 Database migrated successfully!")
}
