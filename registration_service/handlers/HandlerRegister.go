// handler.go
package handlers

import (
	"encoding/json"
	"log"
	"registration_service/dbService"

	amqp "github.com/rabbitmq/amqp091-go"
)

// UserRequest mirrors the JSON that comes over the wire
type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Director string `json:"director"`
}

// Response is sent back to the orchestrator
type Response struct {
	Status  string `json:"status"`  // "ok", "conflict", "error"
	Message string `json:"message"` // details for humans
	// code    int    `json:"code"`
}

func HandleRegister(d amqp.Delivery, ch *amqp.Channel) {
	var req UserRequest
	var res Response

	// Acknowledge the message no matter what (multiple = false)
	defer d.Ack(false)

	// 1. Parse JSON ---------------------------------------------------------
	if err := json.Unmarshal(d.Body, &req); err != nil {
		res.Status = "error"
		res.Message = "Invalid JSON"
		publishReply(ch, d, res)
		return
	}

	// 2. Business logic -----------------------------------------------------
	code, err := dbService.AddInstitution(req.Name, req.Email, req.Director)
	if err != nil {
		if code == 2 {
			res.Status = "conflict"
			res.Message = "Institution already registered"
		} else {
			res.Status = "error"
			res.Message = "Database error"
		}
		publishReply(ch, d, res)
		return
	}

	// 3. Success ------------------------------------------------------------
	res.Status = "ok"
	res.Message = "Institution registered successfully"
	publishReply(ch, d, res)
}

func publishReply(ch *amqp.Channel, d amqp.Delivery, res Response) {
	if d.ReplyTo == "" {
		// fire-and-forget call; nothing to send back
		return
	}

	body, _ := json.Marshal(res)

	err := ch.Publish(
		"",        // default exchange because we address the queue directly
		d.ReplyTo, // queue the caller named
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          body,
		},
	)
	if err != nil {
		log.Printf(" [!] Failed to publish reply: %v", err)
	}
}
