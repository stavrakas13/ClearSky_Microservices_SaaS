package rabbitmq

import (
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

var conn *amqp.Connection
var ch *amqp.Channel

func Connect() {
	var err error
	conn, err = amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Println("⚠️ Could not connect to RabbitMQ:", err)
		return
	}

	ch, err = conn.Channel()
	if err != nil {
		log.Println("⚠️ Could not open RabbitMQ channel:", err)
		return
	}

	if err := ch.ExchangeDeclare(
		"clearsky.events", "topic", true, false, false, false, nil,
	); err != nil {
		log.Fatalf("Declare clearsky.events: %v", err)
	}

	queue := "google_auth.request"
	if _, err := ch.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		log.Fatalf("QueueDeclare %s: %v", queue, err)
	}
	if err := ch.QueueBind(
		queue, "user.login.google", "clearsky.events", false, nil,
	); err != nil {
		log.Fatalf("QueueBind user.login.google: %v", err)
	}
}

func PublishLoginEvent(email string) {
	if ch == nil {
		log.Println("⚠️ RabbitMQ channel not initialized, skipping publish")
		return
	}

	body := `{"event":"user_logged_in","email":"` + email + `"}`

	err := ch.Publish(
		"user_events", // exchange
		"",            // routing key (για fanout: κενό)
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(body),
		},
	)
	if err != nil {
		log.Println("⚠️ Failed to publish message:", err)
	} else {
		log.Printf("📤 Published user_logged_in for %s\n", email)
	}
}
