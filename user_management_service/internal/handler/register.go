package handler

import (
	"net/http"
	"user_management_service/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Struct για το request σώμα
type RegisterRequest struct {
	Username string `json:"username" binding:"omitempty"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required,oneof=student instructor institution_representative"`
}

// Handler function
func Register(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Έλεγχος αν το username υπάρχει ήδη
		var existingUser model.User
		if req.Username != "" && db.Where("username = ?", req.Username).First(&existingUser).Error == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username already registered"})
			return
		}

		// Hashάρισμα του password
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}

		// Δημιουργία νέου χρήστη
		user := model.User{
			ID:           uuid.New().String(),
			Username:     req.Username,
			PasswordHash: string(hashedPassword),
			Role:         req.Role,
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
	}
}
