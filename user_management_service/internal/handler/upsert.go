package handler

import (
	"net/http"

	"user_management_service/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UpsertRequest struct {
	Username  string `json:"username" binding:"required"`
	Role      string `json:"role" binding:"required,oneof=student instructor institution_representative"`
	StudentID string `json:"student_id,omitempty"`
}

func UpsertUser(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req UpsertRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var u model.User
		if err := db.Where("username = ?", req.Username).First(&u).Error; err != nil {
			// create new
			u = model.User{
				ID:        uuid.NewString(),
				Username:  req.Username,
				Role:      req.Role,
				StudentID: req.StudentID,
			}
			db.Create(&u)
		} else {
			// update role if changed
			if u.Role != req.Role {
				u.Role = req.Role
			}
			// set StudentID once supplied
			if req.StudentID != "" && u.StudentID != req.StudentID {
				u.StudentID = req.StudentID
			}
			db.Save(&u)
		}

		c.JSON(http.StatusOK, gin.H{"user_id": u.ID, "role": u.Role})
	}
}
