package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

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
	ErrorDetail string `json:"error,omitempty"` // optional error text
}

type AddInstitutionResp struct {
	Status      string `json:"status"`          // "ok" or "error"
	Message     string `json:"message"`         // human-readable
	ErrorDetail string `json:"error,omitempty"` // optional error text
}

type AddInstitutionReq struct {
	Name string `json:"name" binding:"required"`
}

// PublishAddInstitution publishes an "add.new" event for a given institution.
func PublishAddInstitution(ch *amqp.Channel, req AddInstitutionReq) error {
	corrID := uuid.New().String()
	body, err := json.Marshal(req)
	if err != nil {
		return err
	}
	return ch.Publish(
		"clearSky.events", // exchange
		"add.new",         // routing key
		false, false,      // mandatory, immediate
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			Body:          body,
		},
	)
}

func HandleInstitutionRegistered(c *gin.Context, ch *amqp.Channel) {
	log.Println("Calling registration service...")

	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status:      "error",
			ErrorDetail: err.Error(),
		})
		return
	}

	// Declare a temporary reply queue.
	replyQ, err := ch.QueueDeclare(
		"",    // empty → broker-named queue
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status:      "error",
			ErrorDetail: "queue declare failed: " + err.Error(),
		})
		return
	}

	// Start consuming from the reply queue.
	msgs, err := ch.Consume(
		replyQ.Name,
		"",    // consumer tag
		true,  // auto-ack
		true,  // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status:      "error",
			ErrorDetail: "consume start failed: " + err.Error(),
		})
		return
	}

	// Publish the "institution.registered" event.
	corrID := uuid.New().String()
	reqBody, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status:      "error",
			ErrorDetail: "marshal request failed: " + err.Error(),
		})
		return
	}

	if err := ch.Publish(
		"clearSky.events",        // exchange
		"institution.registered", // routing key
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          reqBody,
		},
	); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status:      "error",
			ErrorDetail: "publish failed: " + err.Error(),
		})
		return
	}

	// Wait for a reply or timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusGatewayTimeout, Response{
				Status:      "error",
				ErrorDetail: "timeout waiting for service",
			})
			return

		case d := <-msgs:
			if d.CorrelationId != corrID {
				continue
			}
			var resp Response
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				c.JSON(http.StatusInternalServerError, Response{
					Status:      "error",
					ErrorDetail: "unmarshal reply failed: " + err.Error(),
				})
				return
			}

			statusCode := http.StatusOK
			if resp.Status != "ok" {
				statusCode = http.StatusBadRequest
			}
			c.JSON(statusCode, resp)

			// After successful registration, publish add.new event.
			addReq := AddInstitutionReq{Name: req.Name}
			if err := PublishAddInstitution(ch, addReq); err != nil {
				log.Printf("failed to publish add institution: %v", err)
			}
			return
		}
	}
}
