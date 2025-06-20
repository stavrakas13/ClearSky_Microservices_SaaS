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
	Type        string `json:"type"` // "register", "login", "delete", "change_password"
	Email       string `json:"email"`
	Password    string `json:"password,omitempty"` // for login/register
	Role        string `json:"role,omitempty"`
	OldPassword string `json:"old_password,omitempty"` // for change_password
	NewPassword string `json:"new_password,omitempty"`
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

			var resp AuthResponse
			var corrID = d.CorrelationId
			var replyTo = d.ReplyTo

			switch req.Type {
			case "register":
				resp = handleRegister(db, req)
			case "login":
				resp = handleLogin(db, req)
			case "delete":
				resp = handleDelete(db, req)
			case "change_password":
				resp = handleChangePassword(db, req)
			default:
				log.Println("Unknown auth type:", req.Type)
				resp = AuthResponse{Status: "error", Message: "Unknown auth type"}
			}

			if replyTo != "" && corrID != "" {
				SendResponse(Channel, replyTo, corrID, resp)
			}
			d.Ack(false)
		}
	}()
}

func handleRegister(db *gorm.DB, req AuthRequest) AuthResponse {
	var existing model.User
	if err := db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return AuthResponse{
			Status:  "error",
			Message: "Email already registered",
		}
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	role := req.Role
	if role == "" {
		role = "student"
	}
	// Optionally: validate role value here
	user := model.User{
		ID:           uuid.NewString(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         role,
	}
	if err := db.Create(&user).Error; err != nil {
		return AuthResponse{
			Status:  "error",
			Message: "Failed to create user",
		}
	}

	return AuthResponse{
		Status: "ok",
		UserID: user.ID,
	}
}

func handleLogin(db *gorm.DB, req AuthRequest) AuthResponse {
	var user model.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return AuthResponse{
			Status:  "error",
			Message: "Invalid credentials",
		}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return AuthResponse{
			Status:  "error",
			Message: "Invalid credentials",
		}
	}
	token, err := jwt.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		return AuthResponse{
			Status:  "error",
			Message: "Token generation failed",
		}
	}
	return AuthResponse{
		Status: "ok",
		Token:  token,
		Role:   user.Role,
		UserID: user.ID,
	}
}

// Add this function:
func handleDelete(db *gorm.DB, req AuthRequest) AuthResponse {
	if err := db.Where("email = ?", req.Email).Delete(&model.User{}).Error; err != nil {
		return AuthResponse{
			Status:  "error",
			Message: "Failed to delete user",
		}
	}
	return AuthResponse{
		Status: "ok",
	}
}

func handleChangePassword(db *gorm.DB, req AuthRequest) AuthResponse {
	var user model.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return AuthResponse{Status: "error", Message: "User not found"}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return AuthResponse{Status: "error", Message: "Invalid current password"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{Status: "error", Message: "Failed to hash new password"}
	}
	user.PasswordHash = string(hash)
	if err := db.Save(&user).Error; err != nil {
		return AuthResponse{Status: "error", Message: "Failed to update password"}
	}
	return AuthResponse{Status: "ok"}
}
