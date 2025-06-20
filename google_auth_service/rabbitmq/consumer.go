package rabbitmq

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"google_auth_service/utils"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/oauth2/google"
)

type GoogleAuthRequest struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type GoogleAuthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
	Email   string `json:"email,omitempty"`
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

	queue := "google_auth.request"
	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.Fatal("Queue declare failed:", err)
	}
	err = ch.QueueBind(queue, "auth.login.google", "orchestrator.commands", false, nil)
	if err != nil {
		log.Fatal("Queue bind failed:", err)
	}

	msgs, err := ch.Consume(queue, "", true, false, false, false, nil)
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
			} else {
				token, _ := utils.GenerateJWT(uuid.NewString(), email, "student")
				resp.Status = "ok"
				resp.Token = token
				resp.Email = email
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
		}
	}()
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
