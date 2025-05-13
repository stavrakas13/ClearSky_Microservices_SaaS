// main.go
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"registration_service/dbService"
	"registration_service/handlers"

	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func main() {
	// ----------------------------------------------------------------------
	// 1. Environment & DB init
	// ----------------------------------------------------------------------
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found; relying on system environment variables.")
	}

	dbService.InitDB()

	// ----------------------------------------------------------------------
	// 2. Rabbit MQ connection / channel
	// ----------------------------------------------------------------------
	rmqURL := os.Getenv("RABBITMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(rmqURL)
	failOnErr(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnErr(err, "Failed to open channel")
	defer ch.Close()

	// ----------------------------------------------------------------------
	// 3. Declare exchange & queue (idempotent; safe if already exist)
	// ----------------------------------------------------------------------
	exchange := "clearSky.events"
	routingKey := "institution.registered"

	err = ch.ExchangeDeclare( //post-office
		exchange, // name
		"direct", // type
		true,     // durable
		false,    // auto-delete
		false,    // internal
		false,    // no-wait
		nil,      // args
	)
	failOnErr(err, "Failed to declare exchange")

	q, err := ch.QueueDeclare(
		routingKey, // queue name == routing key
		true,       // durable
		false,      // delete when unused
		false,      // exclusive
		false,      // no-wait
		nil,        // args
	)
	failOnErr(err, "Failed to declare queue")

	err = ch.QueueBind(
		q.Name,     // queue
		routingKey, // routing key
		exchange,   // exchange
		false,
		nil,
	)
	failOnErr(err, "Failed to bind queue")

	// ----------------------------------------------------------------------
	// 4. QoS & consumer
	// ----------------------------------------------------------------------
	err = ch.Qos(1, 0, false) // one message at a time
	failOnErr(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name,
		"registration_service", // consumer tag
		false,                  // auto-ack -> false (we ack manually)
		false,                  // exclusive
		false,                  // no-local
		false,                  // no-wait
		nil,                    // args
	)
	failOnErr(err, "Failed to register consumer")

	// ----------------------------------------------------------------------
	// 5. Graceful shutdown handling
	// ----------------------------------------------------------------------
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// ----------------------------------------------------------------------
	// 6. Message loop
	// ----------------------------------------------------------------------

	const workers = 20

	log.Println(" [*] Waiting for registration messages …")
	for i := 0; i < workers; i++ {
		go func(id int) {
			log.Printf("Worker %d is ready!", id)
			for d := range msgs {
				handlers.HandleRegister(d, ch)
			}
		}(i)

	}
	select {
	case <-sigs:
		log.Println("Shutdown requested, exiting …")
		return
	}
}
