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
}

func HandleRegister(d amqp.Delivery, ch *amqp.Channel) {
	log.Println("→ HandleRegister called")
	// Acknowledge the message no matter what (multiple = false)
	defer func() {
		d.Ack(false)
		log.Println("… Message acknowledged")
	}()

	var req UserRequest
	var res Response

	// 1. Parse JSON ---------------------------------------------------------
	log.Println("… Parsing JSON payload")
	if err := json.Unmarshal(d.Body, &req); err != nil {
		log.Printf("❌ JSON unmarshal error: %v", err)
		res.Status = "error"
		res.Message = "Invalid JSON"
		publishReply(ch, d, res)
		return
	}
	log.Printf("✅ Parsed UserRequest: %+v", req)

	// 2. Business logic -----------------------------------------------------
	log.Println("… Calling dbService.AddInstitution")
	code, err := dbService.AddInstitution(req.Name, req.Email, req.Director)
	if err != nil {
		if code == 2 {
			log.Printf("⚠ Conflict: institution %q already registered", req.Name)
			res.Status = "conflict"
			res.Message = "Institution already registered"
		} else {
			log.Printf("❌ Database error for %q: %v", req.Name, err)
			res.Status = "error"
			res.Message = "Database error"
		}
		publishReply(ch, d, res)
		return
	}

	// 3. Success ------------------------------------------------------------
	log.Printf("✅ Institution %q registered (code %d)", req.Name, code)
	res.Status = "ok"
	res.Message = "Institution registered successfully"
	publishReply(ch, d, res)
}

func publishReply(ch *amqp.Channel, d amqp.Delivery, res Response) {
	if d.ReplyTo == "" {
		log.Println("… No ReplyTo set; skipping reply publish")
		return
	}

	body, _ := json.Marshal(res)
	log.Printf("… Publishing reply to %q (CorrelationId=%s): %+v", d.ReplyTo, d.CorrelationId, res)

	err := ch.Publish(
		"",        // default exchange
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
		log.Printf("❌ Failed to publish reply: %v", err)
	} else {
		log.Println("✅ Reply published")
	}
}
