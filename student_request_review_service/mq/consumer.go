package mq

import (
	"fmt"

	"student_request_review_service/routes"

	"github.com/streadway/amqp"
)

// function to handle errors
func errorHandling(err error, msg string) {
	if err != nil {
		fmt.Printf("%s: %s\n", msg, err)
	}
}

func StartConsumer() {

	// keys for student events
	exchangeKey := "clearSky.events"
	routingKeysStudent := []string{
		"student.postNewRequest",
		"student.getRequestStatus",
		"student.updateInstructorResponse",
	}

	// declare direct exchange for event routing
	err := Mqch.ExchangeDeclare(
		exchangeKey, // name
		"direct",    // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	errorHandling(err, "Failed to declare exchange")

	// declare a durable queue
	queue, err := Mqch.QueueDeclare(
		"student_queue", // queue name
		true,            // durable
		false,           // delete when unused
		false,           // not exclusive
		false,           // no-wait
		nil,
	)
	errorHandling(err, "Failed to declare queue")

	// bind the queue to each routing key
	for _, key := range routingKeysStudent {
		err := Mqch.QueueBind(
			queue.Name,
			key,
			exchangeKey,
			false,
			nil,
		)
		errorHandling(err, "Failed to bind queue with key "+key)
	}

	// start consuming messages
	msgs, err := Mqch.Consume(
		queue.Name,
		"student_consumer", // consumer tag
		false,              // manual acks!
		false,              // not exclusive
		false,              // no-local (not supported)
		false,              // no-wait
		nil,
	)
	errorHandling(err, "Failed to register consumer")

	fmt.Println("Consumer Declared.")
	fmt.Printf(" [*] Waiting for messages on: %s\n", queue.Name)

	go func() {
		for d := range msgs {
			fmt.Printf("Received message: %s", d.Body)

			response, err := routes.Routing(d.RoutingKey, d.Body)
			if err != nil {
				fmt.Printf("Error processing message for routing key %s: %v", d.RoutingKey, err)
				response = fmt.Sprintf(`{"error": "%s"}`, err.Error())
			}

			fmt.Printf("Reply: %s\n", response)

			err = Mqch.Publish(
				"",        // default exchange for reply
				d.ReplyTo, // reply queue
				false,
				false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				},
			)

			if err != nil {
				fmt.Println("Reply failed.")
				fmt.Println(err)
				d.Nack(false, true) // requeue on publish failure
			} else {
				fmt.Printf("Sent reply to %s\n", d.ReplyTo)
				d.Ack(false)
			}
		}
	}()
}

// FIRST IMPLEMENTATION

/*
	q := "grades.review.requested"

	// Declare queue if not exists.
	_, err := Mqch.QueueDeclare(
		q,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println("Queue declaration failed.")
		fmt.Println(err)

	}
	fmt.Println("Queue declared.")
	fmt.Println(q)

	// Consumer
	msgs, err := Mqch.Consume(
		q,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		fmt.Println("Consumer failed")
		fmt.Println(err)
	}
	fmt.Println("Consumer Declared.")
	fmt.Printf(" [*] Waiting for messages on: %s\n", q)



	go func() {
		for d := range msgs {
			fmt.Printf("Received message: %s", d.Body)

			response, err := routes.Routing(d.Body)
			if err != nil {
				fmt.Println("Error processing message:", err)
				response = fmt.Sprintf(`{"error": "%s"}`, err.Error())
			}
			fmt.Printf("Reply: %s\n", response)

			// Send reply
			err = Mqch.Publish(
				"",
				d.ReplyTo,
				false,
				false,
				amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId,
					Body:          []byte(response),
				},
			)
			if err != nil {
				fmt.Println("Reply failed.")
				fmt.Println(err)
			} else {
				fmt.Printf("Sent reply to %s", d.ReplyTo)
			}
		}
	}()*/
