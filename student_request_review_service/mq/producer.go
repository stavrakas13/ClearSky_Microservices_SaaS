package mq

import (
	"log"
	"os"

	"github.com/streadway/amqp"
)

var MQChannel *amqp.Channel

func InitRabbitMQ() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}

	err = ch.ExchangeDeclare(
		"review_exchange", // exchange name
		"fanout",          // type
		true,              // durable
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare exchange: %v", err)
	}

	MQChannel = ch
}

func PublishReviewCreated(message []byte) error {
	err := MQChannel.Publish(
		"review_exchange", // exchange
		"",                // routing key
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
		})
	return err
}
