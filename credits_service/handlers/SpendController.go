package handlers

import (
	"encoding/json"
	"log"

	"credits_service/dbService"

	amqp "github.com/rabbitmq/amqp091-go"
)

type SpendReq struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"` // Capitalized & correct type
	// code int `json:"code"`
}

type Response struct {
	Status  string `json:"status"`  // "ok", "conflict", "error"
	Message string `json:"message"` // details for humans
	Err     error  `json:"err"`
}

func Spending(d amqp.Delivery, ch *amqp.Channel) {
	log.Printf("[Spending] Received message. CorrelationID=%s, ReplyTo=%s", d.CorrelationId, d.ReplyTo)

	var req SpendReq
	var res Response

	// Ensure the message is acknowledged at the end, no matter what.
	defer func() {
		if err := d.Ack(false); err != nil {
			log.Printf("[Spending] Failed to ack message: %v", err)
		}
	}()

	// 1. Parse JSON ---------------------------------------------------------
	if err := json.Unmarshal(d.Body, &req); err != nil {
		log.Printf("[Spending] JSON unmarshal error: %v | Body=%s", err, string(d.Body))
		res.Status = "error"
		res.Message = "Invalid JSON"
		res.Err = nil
		publishReply(ch, d, res)
		return
	}
	log.Printf("[Spending] Parsed request: %+v", req)

	// 2. Attempt to diminish credits ---------------------------------------
	isComplete, err := dbService.Diminish(req.Name, req.Amount)
	log.Printf("[Spending] dbService.Diminish(Name=%s, Amount=%d) => isComplete=%t, err=%v", req.Name, req.Amount, isComplete, err)

	if err != nil {
		res.Status = "error"
		res.Message = "Error in internal process or not enough credits"
		res.Err = err
		publishReply(ch, d, res)
		return
	}

	if isComplete {
		res.Status = "OK"
		res.Message = "Valid spent of your credits"
		res.Err = nil
		publishReply(ch, d, res)
		return
	}

	// If we reach here, it means credits were diminished but not fully consumed (business rule dependent)
	res.Status = "conflict"
	res.Message = "Partial credits spent; remaining balance exists"
	res.Err = nil
	publishReply(ch, d, res)
}

func publishReply(ch *amqp.Channel, d amqp.Delivery, res Response) {
	// fire-and-forget call; nothing to send back
	if d.ReplyTo == "" {
		log.Printf("[publishReply] ReplyTo empty; not sending any response. CorrelationID=%s", d.CorrelationId)
		return
	}

	body, errMarshal := json.Marshal(res)
	if errMarshal != nil {
		log.Printf("[publishReply] Failed to marshal response: %v | Response=%+v", errMarshal, res)
		return
	}

	log.Printf("[publishReply] Publishing reply. CorrelationID=%s, Body=%s", d.CorrelationId, string(body))

	if err := ch.Publish(
		"",        // default exchange because we address the queue directly
		d.ReplyTo, // queue the caller named
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          body,
		},
	); err != nil {
		log.Printf("[publishReply] Failed to publish reply: %v", err)
	}
}
