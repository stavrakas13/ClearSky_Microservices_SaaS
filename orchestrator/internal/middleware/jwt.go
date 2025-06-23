package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWT secret key from env
var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type Claims struct {
	UserID    string `json:"user_id"`
	Username  string `json:"username,omitempty"`
	Role      string `json:"role"`
	StudentID string `json:"student_id,omitempty"` // Add student_id field
	jwt.RegisteredClaims
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header missing"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid Authorization header format"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username) // Add username to context
		c.Set("role", claims.Role)
		c.Set("student_id", claims.StudentID) // Set student_id in context

		c.Next()
	}
}

// Helper functions for other services to use
func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return userID.(string)
	}
	return ""
}

func GetUsername(c *gin.Context) string {
	if Username, exists := c.Get("username"); exists {
		return Username.(string)
	}
	return ""
}

func GetRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists {
		return role.(string)
	}
	return ""
}

func GetStudentID(c *gin.Context) string {
	if studentID, exists := c.Get("student_id"); exists && studentID != nil {
		return studentID.(string)
	}
	return ""
}

func IsStudent(c *gin.Context) bool {
	return GetRole(c) == "student"
}

func RequireStudentID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsStudent(c) && GetStudentID(c) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID is required for this operation"})
			c.Abort()
			return
		}
		c.Next()
	}
}
