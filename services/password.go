package services

import (
	"log"

	"golang.org/x/crypto/bcrypt"
)

// hashPassword generates bcrypt hash
func HashPassword(password string) string {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}
	return string(hashed)
}
