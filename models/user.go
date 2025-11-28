package models

import (
	"time"
)

type User struct {
	ID         uint      `gorm:"primaryKey"`
	RoleID     uint      `gorm:"not null"`          // Foreign key to roles
	Username   string    `gorm:"size:100;unique"`   // Unique username
	Password   string    `gorm:"size:255;not null"` // Hashed password
	IsVerified bool      `gorm:"default:false"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime"`
}
