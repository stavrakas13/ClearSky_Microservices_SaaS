package handler

import (
	"net/http"
	"user_management_service/internal/model"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Profile(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		var u model.User
		if err := db.First(&u, "id = ?", userID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"user_id":    u.ID,
			"role":       u.Role,
			"student_id": u.StudentID,
		})
	}
}
