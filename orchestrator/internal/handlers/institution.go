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

type AddInstitutionReq struct {
	Name    string `json:"name" binding:"required"`
	Credits int    `json:"credits" binding:"required"`
}

type AddInstitutionResp struct {
	Status      string `json:"status"`          // "ok" or "error"
	Message     string `json:"message"`         // human-readable
	ErrorDetail string `json:"error,omitempty"` // optional error text
}

func HandleAddInstitution(c *gin.Context, ch *amqp.Channel) {
	log.Println("Add Institution API called")

	var req AddInstitutionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AddInstitutionResp{
			Status:      "error",
			ErrorDetail: err.Error(),
		})
		return
	}

	replyQ, err := ch.QueueDeclare(
		"",    // name: empty → broker auto-generates
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AddInstitutionResp{
			Status:      "error",
			ErrorDetail: "queue declare failed: " + err.Error(),
		})
		return
	}

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
		c.JSON(http.StatusInternalServerError, AddInstitutionResp{
			Status:      "error",
			ErrorDetail: "consume start failed: " + err.Error(),
		})
		return
	}

	corrID := uuid.New().String()
	reqBody, _ := json.Marshal(req)

	if err := ch.Publish(
		"clearSky.events", // exchange
		"add.new",         // routing key
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          reqBody,
		},
	); err != nil {
		c.JSON(http.StatusInternalServerError, AddInstitutionResp{
			Status:      "error",
			ErrorDetail: "publish failed: " + err.Error(),
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusGatewayTimeout, AddInstitutionResp{
				Status:      "error",
				ErrorDetail: "timeout waiting for service",
			})
			return

		case d := <-msgs:
			if d.CorrelationId != corrID {
				continue
			}

			var resp AddInstitutionResp
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				c.JSON(http.StatusInternalServerError, AddInstitutionResp{
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
			return
		}
	}
}

func HandleInstitutionRegistered(c *gin.Context, ch *amqp.Channel) {
	log.Printf("We are calling registration service in a little bit...")
	var req UserRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AvailableResp{
			Status:      "error",
			ErrorDetail: err.Error(),
		})
		return
	}

	replyQ, err := ch.QueueDeclare(
		"",    // name: empty → broker generates one
		false, // durable
		true,  // auto-delete when unused
		true,  // exclusive
		false, // no-wait
		nil,
	)
	log.Printf("Queue declared")
	if err != nil {
		c.JSON(http.StatusInternalServerError, AvailableResp{
			Status:      "error",
			ErrorDetail: "queue declare failed: " + err.Error(),
		})
		log.Printf("JSON ERROR")
		return
	}

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

	corrID := uuid.New().String()
	reqBody, _ := json.Marshal(req)

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
			// ignore other correlation IDs
			if d.CorrelationId != corrID {
				continue
			}

			var resp AvailableResp
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

			HandleAddInstitution(c, ch)
			return
		}
	}
}
