package handlers

import (
	"encoding/json"
	"log"

	"credits_service/dbService"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AvailableReq struct {
	Name string `json:"name"`
}

type AvailableResp struct {
	Status      string `json:"status"`            // "ok" or "error"
	Credits     int    `json:"credits,omitempty"` // only on success
	Message     string `json:"message"`           // human-readable
	ErrorDetail string `json:"error,omitempty"`   // optional error text
}

func AvailableHandler(d amqp.Delivery, ch *amqp.Channel) {
	var req AvailableReq
	log.Printf("We are inside the microservices for return available credits")
	if err := json.Unmarshal(d.Body, &req); err != nil {
		log.Printf("Invalid JSON in AvailableHandler: %v", err)
		sendAvailableReplyAndNack(ch, d, AvailableResp{
			Status:      "error",
			Message:     "Invalid JSON",
			ErrorDetail: err.Error(),
		}, false)
		return
	}

	credits, err := dbService.AvailableCredits(req.Name)
	if err != nil {
		log.Printf("DB error in AvailableHandler: %v", err)
		sendAvailableReplyAndNack(ch, d, AvailableResp{
			Status:      "error",
			Message:     "Could not fetch balance",
			ErrorDetail: err.Error(),
		}, true)
		return
	}

	res := AvailableResp{
		Status:  "ok",
		Credits: credits,
		Message: "Current balance",
	}

	if err := publishAvailableReply(ch, d, res); err != nil {
		log.Printf("Publish reply failed in AvailableHandler: %v", err)
		d.Nack(false, true)
		return
	}
	log.Printf(res.Status)
	log.Printf("Available credits %d", res.Credits)
	d.Ack(false)
}

// sendAvailableReplyAndNack publishes the response and nacks the message
func sendAvailableReplyAndNack(ch *amqp.Channel, d amqp.Delivery, res AvailableResp, requeue bool) {
	if err := publishAvailableReply(ch, d, res); err != nil {
		log.Printf("Failed to publish AvailableResp: %v", err)
	}
	d.Nack(false, requeue)
}

// publishAvailableReply serializes a response and publishes it to d.ReplyTo.
func publishAvailableReply(ch *amqp.Channel, d amqp.Delivery, res AvailableResp) error {
	if d.ReplyTo == "" {
		return nil
	}
	body, err := json.Marshal(res)
	if err != nil {
		return err
	}
	return ch.Publish(
		"",        // default exchange
		d.ReplyTo, // callback queue
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: d.CorrelationId,
			Body:          body,
		},
	)
}
