package messaging

import (
	"encoding/json"
	"log"

	"github.com/rabbitmq/amqp091-go"
)

type AuthResponse struct {
	Status  string `json:"status"` // "ok" ή "error"
	Message string `json:"message,omitempty"`
	Token   string `json:"token,omitempty"`
	Role    string `json:"role,omitempty"`
}

func SendResponse(ch *amqp091.Channel, queue string, corrID string, resp AuthResponse) {
	body, err := json.Marshal(resp)
	if err != nil {
		log.Println("❌ Failed to marshal response:", err)
		return
	}

	err = ch.Publish(
		"",    // default exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp091.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			Body:          body,
		},
	)
	if err != nil {
		log.Println("❌ Failed to publish response:", err)
	}
}
