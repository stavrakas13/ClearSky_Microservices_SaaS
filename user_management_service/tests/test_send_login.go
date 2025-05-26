package main

import (
	"encoding/json"
	"log"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	loginReq := map[string]string{
		"type":     "register",
		"email":    "student2@example.com",
		"password": "mypassword123",
		"role":     "student",
	}
	body, _ := json.Marshal(loginReq)

	corrID := uuid.New().String()
	replyQueue, _ := ch.QueueDeclare("", false, true, true, false, nil)

	err = ch.Publish(
		"orchestrator.commands", // exchange
		"auth.register",         // <-- change from "auth.login" to "auth.register"
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQueue.Name,
			Body:          body,
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	msgs, _ := ch.Consume(replyQueue.Name, "", true, true, false, false, nil)
	log.Println("Waiting for response...")
	for d := range msgs {
		log.Printf("Received response: %s", d.Body)
		break
	}
}
