// file: user_management_service/messaging/rabbit.go
package messaging

import (
	"encoding/json"
	"log"
	"os"

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
	Conn, err = amqp.Dial(url)
	if err != nil {
		log.Fatalf("RabbitMQ dial: %v", err)
	}
	Channel, err = Conn.Channel()
	if err != nil {
		log.Fatalf("RabbitMQ channel: %v", err)
	}

	// 1) commands exchange
	if err := Channel.ExchangeDeclare(
		"orchestrator.commands", "topic", true, false, false, false, nil,
	); err != nil {
		log.Fatalf("Declare orchestrator.commands: %v", err)
	}

	// 2) domain‐events exchange
	if err := Channel.ExchangeDeclare(
		"clearSky.events", "topic", true, false, false, false, nil,
	); err != nil {
		log.Fatalf("Declare clearSky.events: %v", err)
	}

	// 3) auth.request queue & bindings
	queue := "auth.request"
	if _, err := Channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		log.Fatalf("QueueDeclare %s: %v", queue, err)
	}
	for _, key := range []string{"auth.register", "auth.login"} {
		if err := Channel.QueueBind(queue, key, "orchestrator.commands", false, nil); err != nil {
			log.Fatalf("QueueBind %s → %s: %v", queue, key, err)
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
		"clearSky.events", // exchange
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
