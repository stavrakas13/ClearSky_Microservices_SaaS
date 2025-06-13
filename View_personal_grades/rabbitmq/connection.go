package rabbitmq

// This package holds the RabbitMQ connection logic for the personal grades
// service.  It exposes a global AMQP channel used by the consumer.

import (
    "fmt"
    "log"
    "os"

    amqp "github.com/rabbitmq/amqp091-go"
)

// Conn and Channel are shared between the consumer workers. They are initialised
// by Init() and closed via Close().
var Conn *amqp.Connection
var Channel *amqp.Channel

// Init establishes a connection to RabbitMQ using the URL from the
// RABBITMQ_URL environment variable. If none is provided it falls back to the
// default guest credentials used in the repository's docker-compose setup. It
// then opens a channel that will be shared by all workers.

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

// Close cleans up the AMQP channel and connection. It is safe to call even if
// the connection was never opened.

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
