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

// HandlePersistAndCalculate publishes exam and grade data so the stats service
// can store them and compute distributions. It does not wait for a reply.
func HandlePersistAndCalculate(c *gin.Context, ch *amqp.Channel) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	body, _ := json.Marshal(payload)
	if err := ch.Publish(
		"clearSky.events",
		"stats.persist_and_calculate",
		false, false,
		amqp.Publishing{ContentType: "application/json", Body: body},
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"status": "queued"})
}

// DistributionRequest represents the payload for fetching grade distributions.
type DistributionRequest struct {
	ClassID  string `json:"class_id"`
	ExamDate string `json:"exam_date"`
}

// HandleGetDistributions requests pre-calculated grade distributions from the
// stats service and returns them to the caller.
func HandleGetDistributions(c *gin.Context, ch *amqp.Channel) {
	var req DistributionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	replyQ, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	msgs, err := ch.Consume(replyQ.Name, "", true, true, false, false, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	corrID := uuid.New().String()
	body, _ := json.Marshal(req)

	if err := ch.Publish(
		"clearSky.events",
		"stats.get_distributions",
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          body,
		},
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case d := <-msgs:
			if d.CorrelationId != corrID {
				continue
			}
			var resp interface{}
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, resp)
			return
		case <-ctx.Done():
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "timeout waiting for service"})
			return
		}
	}
}
