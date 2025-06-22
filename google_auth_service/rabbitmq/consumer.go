package rabbitmq

import (
	"context"
	"encoding/json"
	"google_auth_service/database"
	"google_auth_service/utils"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/oauth2/google"
)

type GoogleAuthRequest struct {
	Type  string `json:"type"`
	Token string `json:"token"`
	Role  string `json:"role,omitempty"`
}

type GoogleAuthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	Token     string `json:"token,omitempty"`
	Email     string `json:"email,omitempty"`
	Role      string `json:"role,omitempty"`
	StudentID string `json:"student_id,omitempty"`
}

var allowedEmailsConsumer = map[string]bool{
	"dimitris.thiv@gmail.com":   true,
	"dimliakis2001@gmail.com":   true,
	"rostav55@gmail.com":        true,
	"anastasvasilis4@gmail.com": true,
}

func StartGoogleAuthConsumer() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatal("RabbitMQ connection failed:", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("RabbitMQ channel failed:", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare("clearSky.events", "direct", true, false, false, false, nil)
	if err != nil {
		log.Fatal("Exchange declare failed:", err)
	}

	queue := "google_auth.request"
	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.Fatal("Queue declare failed:", err)
	}

	// Bind to the correct routing key
	err = ch.QueueBind(queue, "auth.login.google", "clearSky.events", false, nil)
	if err != nil {
		log.Fatal("Queue bind failed:", err)
	}

	msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal("Consume failed:", err)
	}

	go func() {
		for d := range msgs {
			var req GoogleAuthRequest
			if err := json.Unmarshal(d.Body, &req); err != nil {
				continue
			}

			resp := GoogleAuthResponse{}
			email, err := verifyGoogleToken(req.Token)
			if err != nil {
				resp.Status = "error"
				resp.Message = "Invalid Google token"
			} else if !isEmailAllowed(email) {
				resp.Status = "error"
				resp.Message = "Access denied: Email not authorized"
			} else {
				// Find or create user with proper role handling
				var user database.User
				result := database.DB.First(&user, "email = ?", email)

				role := req.Role
				if role == "" {
					role = "institution_representative" // Default for Google users
				}

				var studentID string
				if result.Error != nil {
					// Create new user - only assign student_id if role is student
					if role == "student" {
						studentID = generateStudentID()
					}
					user = database.User{
						Email:     email,
						Role:      role,
						StudentID: studentID,
						Provider:  "google",
					}
					database.DB.Create(&user)
				} else {
					// Only use student_id for students
					if user.Role == "student" {
						studentID = user.StudentID
						if studentID == "" && role == "student" {
							studentID = generateStudentID()
							user.StudentID = studentID
							database.DB.Save(&user)
						}
					}
				}

				userIDStr := strconv.Itoa(int(user.ID))
				token, _ := utils.GenerateJWT(userIDStr, email, role, studentID)
				resp.Status = "ok"
				resp.Token = token
				resp.Email = email
				resp.Role = role
				resp.StudentID = studentID
			}

			body, _ := json.Marshal(resp)
			if d.ReplyTo != "" && d.CorrelationId != "" {
				ch.Publish(
					"", d.ReplyTo, false, false,
					amqp.Publishing{
						ContentType:   "application/json",
						CorrelationId: d.CorrelationId,
						Body:          body,
					},
				)
			}
			d.Ack(false)
		}
	}()
}

// generateStudentID creates a unique student ID
func generateStudentID() string {
	return "STU" + uuid.New().String()[:8]
}

// Helper to verify Google token and extract email
func verifyGoogleToken(idToken string) (string, error) {
	ctx := context.Background()
	oauth2Service, err := google.DefaultClient(ctx, "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		return "", err
	}
	resp, err := oauth2Service.Get("https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=" + idToken)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var tokenInfo struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		return "", err
	}
	if tokenInfo.Email == "" {
		return "", http.ErrNoCookie
	}
	return tokenInfo.Email, nil
}

// Helper function to check if email is allowed
func isEmailAllowed(email string) bool {
	return allowedEmailsConsumer[email]
}
