package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"orchestrator/internal/middleware"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

// submissionLogResponse matches the shape your JS microservice replies with:
type submissionLogResponse struct {
	Status  string          `json:"status"`            // "ok" or "error"
	Message string          `json:"message,omitempty"` // error message
	Data    json.RawMessage `json:"data,omitempty"`    // actual rows
}

// randomCorrelationID generates a random hex string for correlation.
func randomCorrelationID(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HandleSubmissionLogs asks the JS microservice for all submission logs
func HandleSubmissionLogs(c *gin.Context, ch *amqp.Channel) {
	// Get user context from JWT using middleware helpers
	role := middleware.GetRole(c)
	studentID := middleware.GetStudentID(c)
	userID := middleware.GetUserID(c)

	// Build request with user context
	requestPayload := map[string]interface{}{
		"role":    role,
		"user_id": userID,
	}

	// Add student_id for student users
	if role == "student" && studentID != "" {
		requestPayload["student_id"] = studentID
	}

	// 1) Declare a temporary reply queue
	replyQ, err := ch.QueueDeclare(
		"",    // let RabbitMQ generate a random name
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "queue declare failed: " + err.Error()})
		return
	}

	msgs, err := ch.Consume(
		replyQ.Name,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "consume on reply queue failed: " + err.Error()})
		return
	}

	// 2) Generate a correlation ID
	corrID, err := randomCorrelationID(16)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate correlation ID"})
		return
	}

	// 3) Publish the request to the same exchange/routing key your JS service listens on
	body, _ := json.Marshal(requestPayload)
	if err := ch.Publish(
		"clearSky.events", // RABBITMQ_EXCHANGE
		"stats.avail",     // RABBITMQ_SEND_AVAIL_KEY
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          body,
		},
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "publish failed: " + err.Error()})
		return
	}

	// 4) Wait for the matching response (with timeout!)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "timeout waiting for submission logs"})
			return

		case d := <-msgs:
			if d.CorrelationId != corrID {
				continue // not ours, skip
			}

			var resp submissionLogResponse
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid response format"})
				return
			}

			if resp.Status != "ok" {
				c.JSON(http.StatusBadGateway, gin.H{"error": resp.Message})
				return
			}

			// Success! return the raw data.
			c.Data(http.StatusOK, "application/json", resp.Data)
			return
		}
	}
}

// DistributionRequest represents the payload for fetching grade distributions.
type DistributionRequest struct {
	ClassID  string `json:"class_id"`
	ExamDate string `json:"exam_date"`
}

func ForwardToStatistics(ch *amqp.Channel, fileData []byte, filename string) {
	log.Println("[ForwardToStatistics] Encoding data for statistics")

	// Base64 encode the file contents
	encoded := base64.StdEncoding.EncodeToString(fileData)

	// Prepare the persistent message
	msg := amqp.Publishing{
		ContentType:  "text/plain",
		DeliveryMode: amqp.Persistent, // ✅ Makes message durable
		MessageId:    filename,        // Optional metadata
		Timestamp:    time.Now(),      // Optional timestamp
		Body:         []byte(encoded),
	}

	log.Println("[ForwardToStatistics] Publishing to postgrades.statistics")

	// Publish to exchange with the durable routing key
	err := ch.Publish(
		"clearSky.events",       // 🔁 Exchange name (must exist and be durable)
		"postgrades.statistics", // 🎯 Routing key (must match queue binding)
		false,                   // mandatory
		false,                   // immediate
		msg,
	)

	if err != nil {
		log.Printf("[ForwardToStatistics] Failed to publish statistics message: %v\n", err)
	} else {
		log.Println("[ForwardToStatistics] Statistics message published successfully")
	}
}

type rpcResponse struct {
	Status  string              `json:"status"`
	Message string              `json:"message,omitempty"`
	Data    map[string]struct { // e.g. "grade", "Q1"…
		Categories []int `json:"categories"`
		Data       []int `json:"data"`
	} `json:"data,omitempty"`
}

type getGradesRequest struct {
	Course            string `json:"course"            binding:"required"`
	DeclarationPeriod string `json:"declarationPeriod" binding:"required"`
	ClassTitle        string `json:"classTitle"        binding:"required"`
}

// HandleGetGrades is your Gin handler
func HandleGetGrades(ch *amqp.Channel) gin.HandlerFunc {

	return func(c *gin.Context) {
		// 1) bind JSON
		var req getGradesRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// 2) declare a temporary reply queue
		replyQ, err := ch.QueueDeclare(
			"",    // let RabbitMQ name it
			false, // durable
			true,  // delete when unused
			true,  // exclusive
			false, // no-wait
			nil,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "queue declare: " + err.Error()})
			return
		}

		msgs, err := ch.Consume(
			replyQ.Name,
			"",    // consumer
			true,  // auto-ack
			false, // exclusive
			false, // no-local
			false, // no-wait
			nil,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "consume reply queue: " + err.Error()})
			return
		}

		// 3) publish the RPC request
		corrID, err := randomCorrelationID(16)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "generate corrID: " + err.Error()})
			return
		}

		body, _ := json.Marshal(req)
		if err := ch.Publish(
			"clearSky.events", // exchange
			"stats.get",       // routing key
			false, false,
			amqp.Publishing{
				ContentType:   "application/json",
				CorrelationId: corrID,
				ReplyTo:       replyQ.Name,
				Body:          body,
			},
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "publish RPC: " + err.Error()})
			return
		}

		// 4) wait for the reply
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				c.JSON(http.StatusGatewayTimeout, gin.H{"error": "timeout waiting for grades"})
				return

			case d := <-msgs:
				if d.CorrelationId != corrID {
					continue
				}

				var resp rpcResponse
				if err := json.Unmarshal(d.Body, &resp); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "bad RPC response"})
					return
				}

				if resp.Status != "ok" {
					c.JSON(http.StatusBadGateway, gin.H{"error": resp.Message})
					return
				}

				c.JSON(http.StatusOK, resp.Data)
				return
			}
		}
	}
}

// Add helper functions
func GetRole(c *gin.Context) string {
	if role, exists := c.Get("role"); exists && role != nil {
		if str, ok := role.(string); ok {
			return str
		}
	}
	return ""
}

func GetStudentID(c *gin.Context) string {
	if studentID, exists := c.Get("student_id"); exists && studentID != nil {
		if str, ok := studentID.(string); ok {
			return str
		}
	}
	return ""
}

func GetUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists && userID != nil {
		if str, ok := userID.(string); ok {
			return str
		}
	}
	return ""
}
