package handlers

// RabbitMQ handlers for retrieving personal grades information. The orchestrator
// exposes HTTP endpoints that forward the requests to the personal grades
// service using RPC-style messaging.

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"orchestrator/internal/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// When upload grades, update view grades too.

func ForwardToView(ch *amqp.Channel, fileData []byte, filename string) {
	log.Println("[ForwardToStatistics] Encoding data for VIEWING THEM")

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

	log.Println("[ForwardToVIEW] Publishing to postgrades.VIEW")

	// Publish to exchange with the durable routing key
	err := ch.Publish(
		"clearSky.events", // 🔁 Exchange name (must exist and be durable)
		"postgrades.view", // 🎯 Routing key (must match queue binding)
		false,             // mandatory
		false,             // immediate
		msg,
	)

	if err != nil {
		log.Printf("[ForwardToVIEW] Failed to publish VIEW message: %v\n", err)
	} else {
		log.Println("[ForwardToVIEW] VIEW message published successfully")
	}
}

func HandleGetPersonalGrades(c *gin.Context, ch *amqp.Channel) {
	log.Println("[HandleGetPersonalGrades] → entered")

	// Get student_id from JWT context using middleware helper
	studentID := middleware.GetStudentID(c)
	if studentID == "" {
		log.Printf("[HandleGetPersonalGrades] ❌ student ID not found in context")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID is required. Please ensure you're logged in as a student."})
		return
	}

	log.Printf("[HandleGetPersonalGrades] 📥 Received request for student_id: %s", studentID)

	// Build request with student_id from JWT
	req := struct {
		StudentID string `json:"AM"`
	}{
		StudentID: studentID,
	}
	log.Println("[HandleGetPersonalGrades] 🔧 Built request payload")

	// Declare reply queue
	replyQ, err := ch.QueueDeclare(
		"", false, true, true, false, nil,
	)
	if err != nil {
		log.Printf("[HandleGetPersonalGrades] ❌ Failed to declare reply queue: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reply queue"})
		return
	}
	log.Printf("[HandleGetPersonalGrades] ✅ Reply queue declared: %s", replyQ.Name)

	// Start consuming
	msgs, err := ch.Consume(
		replyQ.Name, "", true, true, false, false, nil,
	)
	if err != nil {
		log.Printf("[HandleGetPersonalGrades] ❌ Failed to start consuming: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to consume reply"})
		return
	}
	log.Println("[HandleGetPersonalGrades] 🟢 Started consuming from reply queue")

	// Publish request
	corrID := uuid.New().String()
	body, _ := json.Marshal(req)
	log.Printf("[HandleGetPersonalGrades] 📦 Publishing message with Correlation ID: %s", corrID)
	err = ch.Publish(
		"clearSky.events",
		"view.avail",
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          body,
		},
	)
	if err != nil {
		log.Printf("[HandleGetPersonalGrades] ❌ Publish failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish request"})
		return
	}
	log.Println("[HandleGetPersonalGrades] 🚀 Request published successfully")

	// Set timeout for reply
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("[HandleGetPersonalGrades] ⏳ Waiting for response...")
	for {
		select {
		case <-ctx.Done():
			log.Println("[HandleGetPersonalGrades] ⏰ Timeout while waiting for reply")
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Service timeout"})
			return

		case d := <-msgs:
			log.Printf("[HandleGetPersonalGrades] 📬 Received message with Correlation ID: %s", d.CorrelationId)

			if d.CorrelationId != corrID {
				log.Printf("[HandleGetPersonalGrades] 🔄 Ignoring mismatched Correlation ID: %s", d.CorrelationId)
				continue
			}

			var gradesResp struct {
				Status string        `json:"status"`
				Data   []interface{} `json:"data"`
			}

			if err := json.Unmarshal(d.Body, &gradesResp); err != nil {
				log.Printf("[HandleGetPersonalGrades] ❌ Failed to unmarshal response: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response format"})
				return
			}

			log.Printf("[HandleGetPersonalGrades] ✅ Successfully received response with status: %s", gradesResp.Status)

			statusCode := http.StatusOK
			if gradesResp.Status != "ok" {
				log.Printf("[HandleGetPersonalGrades] ⚠️ Non-ok status received: %s", gradesResp.Status)
				statusCode = http.StatusBadRequest
			}

			log.Printf("[HandleGetPersonalGrades] 📤 Sending JSON response with status code: %d", statusCode)
			c.JSON(statusCode, gradesResp)
			return
		}
	}
}
