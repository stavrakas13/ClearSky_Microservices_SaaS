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

	// Declare the exchange: type fanout (δημοσίευση σε όλους τους subscribers)
	err = ch.ExchangeDeclare(
		"user_events", // exchange name
		"fanout",      // type
		true,          // durable
		false,         // auto-deleted
		false,         // internal
		false,         // no-wait
		nil,           // arguments
	)
	if err != nil {
		log.Println("⚠️ Could not declare exchange:", err)
		return
	}

	log.Println("✅ Connected to RabbitMQ and declared exchange 'user_events'")
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
