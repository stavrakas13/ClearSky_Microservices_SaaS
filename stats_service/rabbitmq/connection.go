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
		rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/" // Για Docker Compose
		// rabbitMQURL = "amqp://guest:guest@localhost:5672/" // Για τοπικό τρέξιμο
		log.Printf("INFO: RABBITMQ_URL not set, using default: %s", rabbitMQURL)
	}

	var err error
	Conn, err = amqp.Dial(rabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ at %s: %w", rabbitMQURL, err)
	}

	Channel, err = Conn.Channel()
	if err != nil {
		Conn.Close()
		return fmt.Errorf("failed to open a channel: %w", err)
	}

	log.Println("INFO: RabbitMQ connection and channel initialized successfully for stats_service.")
	return nil
}

func Close() {
	if Channel != nil {
		err := Channel.Close()
		if err != nil {
			log.Printf("WARNING: Failed to close RabbitMQ channel: %v", err)
		} else {
			log.Println("INFO: RabbitMQ channel closed.")
		}
	}
	if Conn != nil {
		err := Conn.Close()
		if err != nil {
			log.Printf("WARNING: Failed to close RabbitMQ connection: %v", err)
		} else {
			log.Println("INFO: RabbitMQ connection closed.")
		}
	}
}
