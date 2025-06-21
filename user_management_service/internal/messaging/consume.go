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
	Type     string `json:"type"` // "register" ή "login"
	Email    string `json:"email,omitempty"`
	Username string `json:"username,omitempty"`
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
			var resp AuthResponse
			// register
			if req.Type == "register" {
				var existing model.User
				if req.Email != "" && db.Where("email = ?", req.Email).First(&existing).Error == nil {
					resp = AuthResponse{Status: "error", Message: "Email already registered"}
				} else if req.Username != "" && db.Where("username = ?", req.Username).First(&existing).Error == nil {
					resp = AuthResponse{Status: "error", Message: "Username already registered"}
				} else {
					hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
					user := model.User{ID: uuid.NewString(), Email: req.Email, Username: req.Username, PasswordHash: string(hash), Role: "student"}
					if err := db.Create(&user).Error; err != nil {
						resp = AuthResponse{Status: "error", Message: "Failed to create user"}
					} else {
						resp = AuthResponse{Status: "ok", UserID: user.ID, Role: user.Role}
					}
				}
				// login
			} else if req.Type == "login" {
				log.Println("[AuthConsumer] Received login request for:", req.Email, req.Username)
				var user model.User
				if req.Email != "" {
					if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
						resp = AuthResponse{Status: "error", Message: "Invalid credentials"}
						goto send
					}
				} else if req.Username != "" {
					if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
						resp = AuthResponse{Status: "error", Message: "Invalid credentials"}
						goto send
					}
				} else {
					resp = AuthResponse{Status: "error", Message: "Email or username required"}
					goto send
				}
				if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
					resp = AuthResponse{Status: "error", Message: "Invalid credentials"}
				} else {
					token, err := jwt.GenerateToken(user.ID, user.Email, user.Role, user.StudentID)
					if err != nil {
						resp = AuthResponse{Status: "error", Message: "Token generation failed"}
					} else {
						resp = AuthResponse{Status: "ok", Token: token, Role: user.Role, UserID: user.ID}
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
