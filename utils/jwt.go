package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTClaims struct {
	UserID uint `json:"user_id"`
	RoleID uint `json:"role_id"`
	jwt.RegisteredClaims
}

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

func ParseToken(tokenString string) (*JWTClaims, error) {
	// Parse and validate token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return SecretKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
