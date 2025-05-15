// handlers/buy_handler.go
package handlers

import (
	"encoding/json"
	"log"

	"credits_service/dbService"

	amqp "github.com/rabbitmq/amqp091-go"
)

// BuyReq is the payload for a purchase request
type BuyReq struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"`
}

type BuyResponse struct {
	Status      string `json:"status"`          // "ok" or "error"
	Message     string `json:"message"`         // human-readable
	ErrorDetail string `json:"error,omitempty"` // optional, for debugging
}

func HandleBuy(d amqp.Delivery, ch *amqp.Channel) {
	var req BuyReq

	if err := json.Unmarshal(d.Body, &req); err != nil {
		log.Printf("Invalid JSON received: %v", err)
		sendBuyReplyAndNack(ch, d, BuyResponse{
			Status:      "error",
			Message:     "Invalid JSON format",
			ErrorDetail: err.Error(),
		}, false)
		return
	}

	success, err := dbService.BuyCredits(req.Name, req.Amount)
	if err != nil {
		log.Printf("DB error during BuyCredits: %v", err)
		sendBuyReplyAndNack(ch, d, BuyResponse{
			Status:      "error",
			Message:     "Could not process purchase",
			ErrorDetail: err.Error(),
		}, true)
		return
	}

	var res BuyResponse
	if success {
		res = BuyResponse{Status: "ok", Message: "Credits purchased successfully"}
	} else {
		res = BuyResponse{Status: "error", Message: "Unknown error occurred during purchase"}
	}

	if err := publishBuyReply(ch, d, res); err != nil {
		log.Printf("Failed to publish reply: %v", err)
		d.Nack(false, true)
		return
	}
	d.Ack(false)
}

func sendBuyReplyAndNack(ch *amqp.Channel, d amqp.Delivery, res BuyResponse, requeue bool) {
	if err := publishBuyReply(ch, d, res); err != nil {
		log.Printf("Failed to publish error response: %v", err)
	}
	d.Nack(false, requeue)
}

func publishBuyReply(ch *amqp.Channel, d amqp.Delivery, res BuyResponse) error {
	if d.ReplyTo == "" {
		return nil
	}

	body, err := json.Marshal(res)
	if err != nil {
		return err
	}

	return ch.Publish(
		"",        // default exchange
		d.ReplyTo, // routing key (callback queue)
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          body,
		},
	)
}
