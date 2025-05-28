package rabbitmq

import (
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var Conn *amqp.Connection
var Channel *amqp.Channel

func Init() error {
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		// Τοπικό URL αν δεν υπάρχει μεταβλητή περιβάλλοντος
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
		// Για Docker Compose, θα ήταν κάτι σαν "amqp://guest:guest@rabbitmq:5672/"
	}

	var err error
	Conn, err = amqp.Dial(rabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	Channel, err = Conn.Channel()
	if err != nil {
		Conn.Close()
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	log.Println("✅ RabbitMQ connection and channel initialized for stats_service")
	return nil
}

func Close() {
	if Channel != nil {
		Channel.Close()
	}
	if Conn != nil {
		Conn.Close()
	}
	log.Println("🚪 RabbitMQ connection and channel closed for stats_service")
}
