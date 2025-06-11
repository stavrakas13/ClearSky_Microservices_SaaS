package rabbitmq

import (
    "fmt"
    "log"
    "os"

    amqp "github.com/rabbitmq/amqp091-go"
)

var Conn *amqp.Connection
var Channel *amqp.Channel

// Init establishes the RabbitMQ connection and opens a channel.
func Init() error {
    rabbitMQURL := os.Getenv("RABBITMQ_URL")
    if rabbitMQURL == "" {
        rabbitMQURL = "amqp://guest:guest@rabbitmq:5672/"
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

    log.Println("INFO: RabbitMQ connection ready for personal_grades_service")
    return nil
}

// Close cleans up the channel and connection.
func Close() {
    if Channel != nil {
        if err := Channel.Close(); err != nil {
            log.Printf("WARNING: Failed to close RabbitMQ channel: %v", err)
        }
    }
    if Conn != nil {
        if err := Conn.Close(); err != nil {
            log.Printf("WARNING: Failed to close RabbitMQ connection: %v", err)
        }
    }
}
