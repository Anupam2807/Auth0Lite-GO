package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	AccessSecret  = []byte(os.Getenv("secret_key"))
	RefreshSecret = []byte(os.Getenv("secret_key"))
)

func GenerateTokens(userID int, email string) (string, string, error) {

	accessClaims := jwt.MapClaims{
		"userId": userID,
		"email":  email,
		"exp":    time.Now().Add(2 * time.Hour).Unix(),
	}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(AccessSecret)

	if err != nil {
		return "", "", err
	}
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     time.Now().Add(7 * 24 * time.Hour).Unix(),
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(RefreshSecret)
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
