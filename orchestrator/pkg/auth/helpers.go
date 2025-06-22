package auth

import (
	"github.com/gin-gonic/gin"
)

// AuthInfo contains all user authentication information
type AuthInfo struct {
	UserID    string
	Username  string
	Role      string
	StudentID string
}

// GetAuthInfo extracts all authentication info from gin context
func GetAuthInfo(c *gin.Context) AuthInfo {
	return AuthInfo{
		UserID:    getStringFromContext(c, "user_id"),
		Username:  getStringFromContext(c, "username"),
		Role:      getStringFromContext(c, "role"),
		StudentID: getStringFromContext(c, "student_id"),
	}
}

// GetStudentID returns the student ID if user is a student
func GetStudentID(c *gin.Context) string {
	if getStringFromContext(c, "role") == "student" {
		return getStringFromContext(c, "student_id")
	}
	return ""
}

// IsStudent checks if the current user is a student
func IsStudent(c *gin.Context) bool {
	return getStringFromContext(c, "role") == "student"
}

// RequireStudent middleware that ensures only students can access
func RequireStudent() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsStudent(c) {
			c.JSON(403, gin.H{"error": "Access restricted to students only"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireStudentWithID middleware that ensures student has a valid student_id
func RequireStudentWithID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !IsStudent(c) {
			c.JSON(403, gin.H{"error": "Access restricted to students only"})
			c.Abort()
			return
		}
		if GetStudentID(c) == "" {
			c.JSON(400, gin.H{"error": "Student ID is required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func getStringFromContext(c *gin.Context, key string) string {
	if value, exists := c.Get(key); exists && value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}
