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
	var req SpendReq
	var res Response

	defer d.Ack(false)

	// 1. Parse JSON ---------------------------------------------------------
	if err := json.Unmarshal(d.Body, &req); err != nil {
		res.Status = "error"
		res.Message = "Invalid JSON"
		res.Err = nil
		publishReply(ch, d, res)
		return
	}

	IsComplete, err := dbService.Diminish(req.Name, req.Amount)

	if err != nil {
		res.Status = "error"
		res.Message = "Error in internal process"
		res.Err = err
		publishReply(ch, d, res)
		return
	}

	if IsComplete {
		res.Status = "OK"
		res.Message = "Valid"
		res.Err = nil
		publishReply(ch, d, res)
		return
	} else {
		res.Status = "error"
		res.Message = "Not enough credits"
		res.Err = nil
		publishReply(ch, d, res)
		return
	}

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
