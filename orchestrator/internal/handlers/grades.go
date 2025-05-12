package handlers

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func handleInitialGradesUploaded(d amqp.Delivery) {
	// Για dummy handler: decode generic JSON
	var payload map[string]interface{}
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		log.Printf("Invalid payload on grades.initial.uploaded: %v", err)
		d.Nack(false, false) // στέλνει στο DLQ
		return
	}
	log.Printf("Received grades.initial.uploaded payload: %+v", payload)
	d.Ack(false)

}

// handleFinalGradesUploaded processes grades.final.uploaded events
func handleFinalGradesUploaded(d amqp.Delivery) {}
