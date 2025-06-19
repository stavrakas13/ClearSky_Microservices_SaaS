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

type PurchaseRequest struct {
	Name   string `json:"name" binding:"required"`
	Amount int    `json:"amount" binding:"required,gt=0"`
}

type PurchaseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type SpendReq struct {
	Name   string `json:"name"`
	Amount int    `json:"amount"` // Capitalized & correct type
	// code int `json:"code"`
}

type SpendResponse struct {
	Status  string `json:"status"`  // "ok", "conflict", "error"
	Message string `json:"message"` // details for humans
	Err     error  `json:"err"`
	Error   string `json:"error,omitempty"`
}

type AvailableReq struct {
	Name string `json:"name" binding:"required"`
}

type AvailableResp struct {
	Status      string `json:"status"`            // "ok" or "error"
	Credits     int    `json:"credits,omitempty"` // only on success
	Message     string `json:"message"`           // human-readable
	ErrorDetail string `json:"error,omitempty"`   // optional error text
}

func HandleCreditsAvail(c *gin.Context, ch *amqp.Channel) {
	log.Printf("We made the API CALL")
	var req AvailableReq
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
		c.JSON(http.StatusInternalServerError, AvailableResp{
			Status:      "error",
			ErrorDetail: "consume start failed: " + err.Error(),
		})
		return
	}

	corrID := uuid.New().String()
	reqBody, _ := json.Marshal(req)

	if err := ch.Publish(
		"clearSky.events", // exchange
		"credits.avail",   // routing key
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          reqBody,
		},
	); err != nil {
		c.JSON(http.StatusInternalServerError, AvailableResp{
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
			c.JSON(http.StatusGatewayTimeout, AvailableResp{
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
				c.JSON(http.StatusInternalServerError, AvailableResp{
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

// this function will be used after uploaded final grades.
func HandleCreditsSpent(ch *amqp.Channel) error {
	type Payload struct {
		Name   string  `json:"name"`
		Amount float64 `json:"amount"`
	}
	body := Payload{
		Name:   "NTUA",
		Amount: 1,
	}
	jsonbody, err := json.Marshal(body)
	if err != nil {
		return err
	}
	return ch.Publish(
		"clearSky.events",
		"credits.spent",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         jsonbody,
		},
	)
}

func HandleCreditsPurchased(c *gin.Context, ch *amqp.Channel) {
	log.Println("[HandleCreditsPurchased] → entered")

	// 1. Bind JSON
	var req PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[HandleCreditsPurchased] ❌ bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[HandleCreditsPurchased] 📥 request: name=%s amount=%d", req.Name, req.Amount)

	// 2. Declare a temporary reply queue
	log.Println("[HandleCreditsPurchased] ⏳ declaring reply queue")
	replyQ, err := ch.QueueDeclare(
		"",    // empty name = broker-generated
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Printf("[HandleCreditsPurchased] ❌ QueueDeclare failed: %v", err)
		c.JSON(http.StatusInternalServerError, PurchaseResponse{
			Status:  "error",
			Message: "failed to create reply queue",
			Error:   err.Error(),
		})
		return
	}
	log.Printf("[HandleCreditsPurchased] ✅ declared reply queue: %s", replyQ.Name)

	// 3. Start consuming replies
	log.Printf("[HandleCreditsPurchased] ⏳ start consuming on %s", replyQ.Name)
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
		log.Printf("[HandleCreditsPurchased] ❌ Consume failed: %v", err)
		c.JSON(http.StatusInternalServerError, PurchaseResponse{
			Status:  "error",
			Message: "failed to start consuming replies",
			Error:   err.Error(),
		})
		return
	}
	log.Println("[HandleCreditsPurchased] ✅ consumer started")

	// 4. Publish the event
	corrID := uuid.New().String()
	body, _ := json.Marshal(req)
	log.Printf("[HandleCreditsPurchased] ⏳ publishing to exchange=clearSky.events routingKey=credits.purchased corrID=%s", corrID)
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
		log.Printf("[HandleCreditsPurchased] ❌ Publish failed: %v", err)
		c.JSON(http.StatusInternalServerError, PurchaseResponse{
			Status:  "error",
			Message: "failed to publish request",
			Error:   err.Error(),
		})
		return
	}
	log.Println("[HandleCreditsPurchased] ✅ published, waiting for reply...")

	// 5. Wait for reply (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Println("[HandleCreditsPurchased] ⏰ timeout waiting for reply")
			c.JSON(http.StatusGatewayTimeout, PurchaseResponse{
				Status:  "error",
				Message: "service timeout",
			})
			return

		case d := <-msgs:
			log.Printf("[HandleCreditsPurchased] 🔔 got delivery corrID=%s", d.CorrelationId)
			if d.CorrelationId != corrID {
				log.Printf("[HandleCreditsPurchased] 🔍 ignoring stray message (corrID=%s)", d.CorrelationId)
				continue
			}

			log.Println("[HandleCreditsPurchased] ⏳ unmarshalling reply")
			var resp PurchaseResponse
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				log.Printf("[HandleCreditsPurchased] ❌ invalid reply format: %v", err)
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
			log.Printf("[HandleCreditsPurchased] ✅ replying to client with status=%d message=%q", statusCode, resp.Message)
			c.JSON(statusCode, resp)
			return
		}
	}
}
