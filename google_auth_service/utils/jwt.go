package utils

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username,omitempty"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	StudentID string `json:"student_id,omitempty"` // Add student_id field
	jwt.RegisteredClaims
}

func GenerateJWT(userID, email, role, studentID string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &Claims{
		UserID:    userID,
		Username:  email, // Use email as username for Google users
		Email:     email,
		Role:      role,
		StudentID: studentID, // Include student_id in JWT
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func VerifyJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
