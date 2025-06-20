// file: user_management_service/messaging/rabbit.go
package messaging

import (
	"encoding/json"
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	Conn    *amqp.Connection
	Channel *amqp.Channel
)

// Init connects to RabbitMQ, declares exchanges & auth queue+bindings.
func Init() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@rabbitmq:5672/"
	}
	var err error
	for i := 0; i < 10; i++ { // Try 10 times
		Conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Printf("RabbitMQ dial failed: %v (retrying in 3s)", err)
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatalf("RabbitMQ dial: %v", err)
	}

	// Initialize Channel after successful connection
	Channel, err = Conn.Channel()
	if err != nil {
		log.Fatalf("RabbitMQ channel: %v", err)
	}

	// 1) Declare the orchestrator.commands exchange for RPC commands
	if err := Channel.ExchangeDeclare(
		"orchestrator.commands", "topic", true, false, false, false, nil,
	); err != nil {
		log.Fatalf("Declare orchestrator.commands: %v", err)
	}

	// 2) Declare the clearSky.events exchange for publishing domain events
	if err := Channel.ExchangeDeclare(
		"clearSky.events", "topic", true, false, false, false, nil,
	); err != nil {
		log.Fatalf("Declare clearSky.events: %v", err)
	}

	// 3) Declare and bind the auth.request queue to orchestrator.commands
	queue := "auth.request"
	if _, err := Channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		log.Fatalf("QueueDeclare %s: %v", queue, err)
	}
	for _, key := range []string{"auth.register", "auth.login", "auth.delete", "auth.change_password"} {
		if err := Channel.QueueBind(queue, key, "orchestrator.commands", false, nil); err != nil {
			log.Fatalf("QueueBind %s → %s: %v", queue, key, err)
		}
	}

	// 4) Declare and bind the queue for login/register
	for _, key := range []string{"user.login", "user.register"} {
		if err := Channel.QueueBind(
			queue, key, "clearSky.events", false, nil,
		); err != nil {
			log.Fatalf("QueueBind %s: %v", key, err)
		}
	}
}

// PublishEvent στέλνει ένα event στο clearSky.events με το δοσμένο routingKey
func PublishEvent(routingKey string, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("PublishEvent marshal: %v", err)
		return
	}
	err = Channel.Publish(
		"clearSky.events", // exchange (fixed name)
		routingKey,        // routing key
		false, false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Printf("PublishEvent publish: %v", err)
	}
}
