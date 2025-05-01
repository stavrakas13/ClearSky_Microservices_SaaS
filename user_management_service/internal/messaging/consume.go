package messaging

import (
	"encoding/json"
	"log"
	"user_management_service/internal/model"

	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	jwtutil "user_management_service/pkg/jwt"
)

type AuthRequest struct {
	Type          string `json:"type"` // "register" ή "login"
	Email         string `json:"email"`
	Password      string `json:"password"`
	Role          string `json:"role,omitempty"` // μόνο για register
	ReplyTo       string `json:"reply_to"`       // όνομα callback queue
	CorrelationID string `json:"correlation_id"` // για συσχέτιση
}

func ConsumeAuthQueue(ch *amqp091.Channel, db *gorm.DB) {
	q, _ := ch.QueueDeclare("auth.request", true, false, false, false, nil)
	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	go func() {
		for d := range msgs {
			var req AuthRequest
			if err := json.Unmarshal(d.Body, &req); err != nil {
				log.Println("❌ Invalid message:", err)
				continue
			}

			switch req.Type {
			case "register":
				handleRegister(ch, db, req)
			case "login":
				handleLogin(ch, db, req)
			default:
				log.Println("❌ Unknown request type:", req.Type)
			}
		}
	}()
}

func handleRegister(ch *amqp091.Channel, db *gorm.DB, req AuthRequest) {
	var existing model.User
	if err := db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		SendResponse(ch, req.ReplyTo, req.CorrelationID, AuthResponse{
			Status:  "error",
			Message: "Email already registered",
		})
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := model.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hash),
		Role:         req.Role,
	}

	if err := db.Create(&user).Error; err != nil {
		SendResponse(ch, req.ReplyTo, req.CorrelationID, AuthResponse{
			Status:  "error",
			Message: "Failed to create user",
		})
		return
	}

	SendResponse(ch, req.ReplyTo, req.CorrelationID, AuthResponse{
		Status:  "ok",
		Message: "User created successfully",
	})
}

func handleLogin(ch *amqp091.Channel, db *gorm.DB, req AuthRequest) {
	var user model.User
	if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		SendResponse(ch, req.ReplyTo, req.CorrelationID, AuthResponse{
			Status:  "error",
			Message: "Invalid credentials (email)",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		SendResponse(ch, req.ReplyTo, req.CorrelationID, AuthResponse{
			Status:  "error",
			Message: "Invalid credentials (password)",
		})
		return
	}

	token, err := jwtutil.GenerateToken(user.ID, user.Email, user.Role)
	if err != nil {
		SendResponse(ch, req.ReplyTo, req.CorrelationID, AuthResponse{
			Status:  "error",
			Message: "Failed to generate token",
		})
		return
	}

	SendResponse(ch, req.ReplyTo, req.CorrelationID, AuthResponse{
		Status: "ok",
		Token:  token,
		Role:   user.Role,
	})
}
