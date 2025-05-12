package handlers

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// HandleUserCreated is a dummy handler that logs and ACKs the message.
func HandleUserCreated(d amqp.Delivery) {
	log.Printf("[Handler] user.created received: %s", string(d.Body))
	// Acknowledge message
	if err := d.Ack(false); err != nil {
		log.Printf("[Handler] Ack failed: %v", err)
	}
}
