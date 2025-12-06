package models

type Role struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"unique;not null"`
	Description string
	Permissions []Permission `gorm:"many2many:role_permissions;"`
}
