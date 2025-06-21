// file: user_management_service/messaging/consumer.go
package messaging

import (
	"encoding/json"
	"log"
	"user_management_service/internal/model"
	"user_management_service/pkg/jwt"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthRequest struct {
	Type     string `json:"type"` // "register" ή "login"
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Status  string `json:"status"`            // "ok" ή "error"
	Message string `json:"message,omitempty"` // λόγος σφάλματος
	Token   string `json:"token,omitempty"`
	Role    string `json:"role,omitempty"`
	UserID  string `json:"userId,omitempty"`
}

func ConsumeAuthQueue(db *gorm.DB) {
	msgs, err := Channel.Consume(
		"auth.request", "", false, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("Consume auth.request: %v", err)
	}

	go func() {
		for d := range msgs {
			var req AuthRequest
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Println("Invalid auth request:", err)
				d.Nack(false, false)
				continue
			}

			switch req.Type {
			case "register":
				handleRegister(db, req)
			case "login":
				handleLogin(db, req)
			default:
				log.Println("Unknown auth type:", req.Type)
			}
			d.Ack(false)
		}
	}()
}

func handleRegister(db *gorm.DB, req AuthRequest) {
	var existing model.User
	if err := db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		PublishEvent("auth.register.failure", AuthResponse{
			Status:  "error",
			Message: "Email already registered",
		})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := model.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         "student",
	}
	if err := db.Create(&user).Error; err != nil {
		PublishEvent("auth.register.failure", AuthResponse{
			Status:  "error",
			Message: "Failed to create user",
		})
		return
	}

	PublishEvent("auth.register.success", AuthResponse{
		Status: "ok",
		UserID: user.ID,
	})
}

func handleLogin(db *gorm.DB, req AuthRequest) {
	var user model.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		PublishEvent("auth.login.failure", AuthResponse{
			Status:  "error",
			Message: "Invalid credentials",
		})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		PublishEvent("auth.login.failure", AuthResponse{
			Status:  "error",
			Message: "Invalid credentials",
		})
		return
	}
	token, err := jwt.GenerateToken(user.ID, user.Email, user.Role, user.StudentID)
	if err != nil {
		PublishEvent("auth.login.failure", AuthResponse{
			Status:  "error",
			Message: "Token generation failed",
		})
		return
	}
	PublishEvent("auth.login.success", AuthResponse{
		Status: "ok",
		Token:  token,
		Role:   user.Role,
		UserID: user.ID,
	})
}
