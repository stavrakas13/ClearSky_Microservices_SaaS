package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PurchaseRequest struct {
	Name   string `json:"name" binding:"required"`
	Amount int    `json:"amount" binding:"required,gt=0"`
}

type PurchaseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// handleCreditsAvail processes credits.avail events
func handleCreditsAvail(d amqp.Delivery) {}

// handleCreditsSpent processes credits.spent events
func handleCreditsSpent(d amqp.Delivery) {}

func HandleCreditsPurchased(c *gin.Context, ch *amqp.Channel) {

	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 2. Declare a temporary reply queue
	replyQ, err := ch.QueueDeclare(
		"",    // empty name = let broker generate a unique name
		false, // durable
		true,  // delete when unused (auto-delete)
		true,  // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PurchaseResponse{
			Status:  "error",
			Message: "failed to create reply queue",
			Error:   err.Error(),
		})
		return
	}

	msgs, err := ch.Consume(
		replyQ.Name,
		"",    // consumer tag
		true,  // auto-ack replies
		true,  // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PurchaseResponse{
			Status:  "error",
			Message: "failed to start consuming replies",
			Error:   err.Error(),
		})
		return
	}

	corrID := uuid.New().String()
	body, _ := json.Marshal(req)

	err = ch.Publish(
		"clearSky.events",   // exchange
		"credits.purchased", // routing key
		false,               // mandatory
		false,               // immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          body,
		},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, PurchaseResponse{
			Status:  "error",
			Message: "failed to publish request",
			Error:   err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusGatewayTimeout, PurchaseResponse{
				Status:  "error",
				Message: "service timeout",
			})
			return

		case d := <-msgs:
			if d.CorrelationId != corrID {
				// ignore stray messages
				continue
			}

			var resp PurchaseResponse
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				c.JSON(http.StatusInternalServerError, PurchaseResponse{
					Status:  "error",
					Message: "invalid reply format",
					Error:   err.Error(),
				})
				return
			}

			statusCode := http.StatusOK
			if resp.Status != "ok" {
				statusCode = http.StatusBadRequest
			}
			c.JSON(statusCode, resp)
			return
		}
	}
}
