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

	amqp "github.com/rabbitmq/amqp091-go"
)

type AuthRequest struct {
	Type        string `json:"type"` // "register" ή "login"
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	Role        string `json:"role,omitempty"`
	StudentID   string `json:"student_id,omitempty"` // Add student_id field
	OldPassword string `json:"old_password,omitempty"`
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
			// register
			if req.Type == "register" {
				if req.Username == "" {
					resp = AuthResponse{Status: "error", Message: "Username required"}
					goto send
				}

				// Check if student_id is required for student role
				role := req.Role
				if role == "" {
					role = "student"
				}

				if role == "student" && req.StudentID == "" {
					resp = AuthResponse{Status: "error", Message: "Student ID required for student registration"}
					goto send
				}

				var existing model.User
				if err := db.Where("username = ?", req.Username).First(&existing).Error; err == nil {
					resp = AuthResponse{Status: "error", Message: "Username already registered"}
				} else {
					// Check if student_id already exists (if provided)
					if req.StudentID != "" {
						var existingStudent model.User
						if err := db.Where("student_id = ?", req.StudentID).First(&existingStudent).Error; err == nil {
							resp = AuthResponse{Status: "error", Message: "Student ID already registered"}
							goto send
						}
					}

					hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
					user := model.User{
						ID:           uuid.NewString(),
						Username:     req.Username,
						PasswordHash: string(hash),
						Role:         role,
						StudentID:    req.StudentID, // Set student_id
					}
					if err := db.Create(&user).Error; err != nil {
						resp = AuthResponse{Status: "error", Message: "Failed to create user"}
					} else {
						resp = AuthResponse{Status: "ok", UserID: user.ID, Role: user.Role}
					}
				}
				// login
			} else if req.Type == "login" {
				log.Println("[AuthConsumer] Received login request for username:", req.Username)
				var user model.User
				if req.Username != "" {
					if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
						resp = AuthResponse{Status: "error", Message: "Invalid credentials"}
						goto send
					}
				} else {
					resp = AuthResponse{Status: "error", Message: "Username required"}
					goto send
				}
				if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
					resp = AuthResponse{Status: "error", Message: "Invalid credentials"}
				} else {
					token, err := jwt.GenerateToken(user.ID, user.Username, user.Role, user.StudentID)
					if err != nil {
						resp = AuthResponse{Status: "error", Message: "Token generation failed"}
					} else {
						resp = AuthResponse{Status: "ok", Token: token, Role: user.Role, UserID: user.ID}
					}
				}
				// change_password
			} else if req.Type == "change_password" {
				var user model.User
				if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
					resp = AuthResponse{Status: "error", Message: "Invalid credentials"}
				} else if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)) != nil {
					resp = AuthResponse{Status: "error", Message: "Invalid credentials"}
				} else {
					newHash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
					if err := db.Model(&user).Update("password_hash", string(newHash)).Error; err != nil {
						resp = AuthResponse{Status: "error", Message: "Failed to update password"}
					} else {
						resp = AuthResponse{Status: "ok"}
					}
				}
			} else {
				resp = AuthResponse{Status: "error", Message: "Unknown request type"}
			}
		send:
			// send RPC reply
			body, _ := json.Marshal(resp)
			if d.ReplyTo != "" {
				Channel.Publish("", d.ReplyTo, false, false, amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          body,
				})
			}
			d.Ack(false)
		}
	}()
}
