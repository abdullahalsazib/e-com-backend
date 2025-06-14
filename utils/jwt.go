package utils

import (
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var JWT_SECRET = []byte("your_secret_key")

func GenerateToken(userID uint, email string, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})
	return token.SignedString(JWT_SECRET)
}

func VerifyToken(tokeString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokeString, func(t *jwt.Token) (interface{}, error) {
		return JWT_SECRET, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
