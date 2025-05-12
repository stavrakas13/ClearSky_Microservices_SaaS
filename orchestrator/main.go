package main

import (
	"log"
	"os"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// 1. Διαβάζω τη URL από env var
	amqpURL := os.Getenv("RABBITMQ_URL")
	if amqpURL == "" {
		amqpURL = "amqp://guest:guest@localhost:5672/"
	}

	// 2. Σύνδεση & κανάλι
	conn, err := amqp.Dial(amqpURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// 3. Δήλωση ενός topic exchange
	exchangeName := "orchestrator.commands"
	err = ch.ExchangeDeclare(
		exchangeName, // name
		"topic",      // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	failOnError(err, "Failed to declare exchange")

	// 4. Παράδειγμα publish σε routing key "service.a.doTask"
	go func() {
		for {
			body := `{"orderId":123,"action":"process"}`
			err = ch.Publish(
				exchangeName,       // exchange
				"service.a.doTask", // routing key
				false,              // mandatory
				false,              // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        []byte(body),
				},
			)
			failOnError(err, "Failed to publish message")
			log.Printf(" [x] Sent %s", body)
			time.Sleep(5 * time.Second)
		}
	}()

	// 5. Δήλωση queue & binding για να καταναλώσουμε μηνύματα που στείλει κάποιος σε "orchestrator.commands"
	queueName := "orchestrator.inbox"
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // args
	)
	failOnError(err, "Failed to declare queue")

	err = ch.QueueBind(
		queueName,        // queue name
		"orchestrator.*", // routing key pattern
		exchangeName,     // exchange
		false,
		nil,
	)
	failOnError(err, "Failed to bind queue")

	msgs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		false,     // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	failOnError(err, "Failed to register consumer")

	// 6. Loop που περιμένει και επεξεργάζεται μηνύματα
	forever := make(chan bool)
	go func() {
		for d := range msgs {
			log.Printf(" [x] Received %s with key %s", d.Body, d.RoutingKey)
			// → εδώ βάζεις την business logic
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
