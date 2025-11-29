package services

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var SecretKey = []byte("SUPER_SECRET_KEY")

func GenerateToken(userId uint, roleId uint) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"role_id": roleId,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(SecretKey)
}
