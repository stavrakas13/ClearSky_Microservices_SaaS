package mq

import (
	"fmt"
	"log"

	"student_request_review_service/routes"

	"github.com/streadway/amqp"
)

func StartConsumer() {
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
			log.Printf("Received message: %s", d.Body)

			response, err := routes.Routing(d.Body)
			if err != nil {
				log.Println("Error processing message:", err)
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
	}()
}
