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
		"clearSky.events", // 🔁 Exchange name (must exist and be durable)
		"postgrades.view", // 🎯 Routing key (must match queue binding)
		false,             // mandatory
		false,             // immediate
		msg,
	)

	if err != nil {
		log.Printf("[ForwardToStatistics] Failed to publish statistics message: %v\n", err)
	} else {
		log.Println("[ForwardToStatistics] Statistics message published successfully")
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

	log.Printf("[HandleGetPersonalGrades] 📥 request for student_id: %s", studentID)

	// Build request with student_id from JWT
	req := struct {
		StudentID string `json:"student_id"`
	}{
		StudentID: studentID,
	}

	// 2. Declare reply queue
	replyQ, err := ch.QueueDeclare(
		"", false, true, true, false, nil,
	)
	if err != nil {
		log.Printf("[HandleGetGradesByAM] ❌ failed to declare reply queue: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create reply queue"})
		return
	}

	// 3. Start consuming
	msgs, err := ch.Consume(
		replyQ.Name, "", true, true, false, false, nil,
	)
	if err != nil {
		log.Printf("[HandleGetGradesByAM] ❌ failed to consume: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to consume reply"})
		return
	}

	// 4. Publish request
	corrID := uuid.New().String()
	body, _ := json.Marshal(req)
	err = ch.Publish(
		"clearSky.events", // Exchange
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
		log.Printf("[HandleGetGradesByAM] ❌ publish failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish request"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Println("[HandleGetGradesByAM] ⏰ timeout waiting for reply")
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "Service timeout"})
			return

		case d := <-msgs:
			if d.CorrelationId != corrID {
				log.Printf("[HandleGetGradesByAM] 🔍 ignoring message with corrID=%s", d.CorrelationId)
				continue
			}

			var gradesResp struct {
				Status string        `json:"status"`
				Data   []interface{} `json:"data"`
			}

			if err := json.Unmarshal(d.Body, &gradesResp); err != nil {
				log.Printf("[HandleGetGradesByAM] ❌ unmarshal failed: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid response format"})
				return
			}

			statusCode := http.StatusOK
			if gradesResp.Status != "ok" {
				statusCode = http.StatusBadRequest
			}

			c.JSON(statusCode, gradesResp)
			return
		}
	}
}
