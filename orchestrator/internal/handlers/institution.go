// handlers/institution.go
package handlers

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// UserRequest is the payload we expect from the client.
type UserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Director string `json:"director"`
}

const storagePath = "requests.json"

// Response is the structure sent back to the client.
type Response struct {
	Status      string `json:"status"`                // "ok", "conflict", "error"
	Message     string `json:"message,omitempty"`     // optional human-readable message
	ErrorDetail string `json:"errorDetail,omitempty"` // optional detailed error
}

// HandleInstitutionRegistered receives a registration request, logs it to disk as NDJSON,
// publishes an AMQP event, waits for the worker reply, and then returns the worker’s response.
func HandleInstitutionRegistered(c *gin.Context, ch *amqp.Channel) {
	log.Println("→ HandleInstitutionRegistered called")

	var req UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{"error", "", err.Error()})
		return
	}
	log.Printf("✅ Parsed UserRequest: %+v", req)

	// append to ./requests.json
	if data, err := json.Marshal(req); err != nil {
		log.Printf("❌ marshal for storage: %v", err)
	} else if f, err := os.OpenFile(storagePath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644,
	); err != nil {
		log.Printf("❌ open %s: %v", storagePath, err)
	} else {
		defer f.Close()
		if _, err := f.Write(append(data, '\n')); err != nil {
			log.Printf("❌ write to %s: %v", storagePath, err)
		} else {
			log.Println("✅ Stored request in", storagePath)
		}
	}

	// 3️⃣ Declare a temporary reply queue
	log.Println("… Declaring temporary reply queue")
	replyQ, err := ch.QueueDeclare(
		"",    // name (empty → let the broker generate)
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Printf("❌ Queue declare failed: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Status:      "error",
			ErrorDetail: "queue declare failed: " + err.Error(),
		})
		return
	}
	log.Printf("✅ Reply queue declared: %s", replyQ.Name)

	// 4️⃣ Start consuming on the reply queue
	log.Println("… Starting consumer on reply queue")
	msgs, err := ch.Consume(
		replyQ.Name, // queue
		"",          // consumer tag
		true,        // auto-ack
		true,        // exclusive
		false,       // no-local
		false,       // no-wait
		nil,         // args
	)
	if err != nil {
		log.Printf("❌ Consume start failed: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Status:      "error",
			ErrorDetail: "consume start failed: " + err.Error(),
		})
		return
	}
	log.Println("✅ Consumer started")

	// 5️⃣ Publish the institution.registered event
	corrID := uuid.New().String()
	body, err := json.Marshal(req)
	if err != nil {
		log.Printf("❌ Marshal request failed: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Status:      "error",
			ErrorDetail: "marshal request failed: " + err.Error(),
		})
		return
	}
	log.Printf("… Publishing event with CorrelationId=%s", corrID)
	if err := ch.Publish(
		"clearSky.events",        // exchange
		"institution.registered", // routing key
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          body,
		},
	); err != nil {
		log.Printf("❌ Publish failed: %v", err)
		c.JSON(http.StatusInternalServerError, Response{
			Status:      "error",
			ErrorDetail: "publish failed: " + err.Error(),
		})
		return
	}
	log.Println("✅ Message published, awaiting reply")

	// 6️⃣ Wait for a reply or timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Printf("⏱ Timeout waiting for reply (CorrelationId=%s)", corrID)
			c.JSON(http.StatusGatewayTimeout, Response{
				Status:      "error",
				ErrorDetail: "timeout waiting for service",
			})
			return

		case d := <-msgs:
			log.Printf("… Received message: CorrelationId=%s", d.CorrelationId)
			if d.CorrelationId != corrID {
				log.Printf("🔄 Correlation ID mismatch (expected=%s, got=%s), skipping", corrID, d.CorrelationId)
				continue
			}
			log.Println("✅ Correlation ID matches, processing reply")

			var resp Response
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				log.Printf("❌ Unmarshal reply failed: %v", err)
				c.JSON(http.StatusInternalServerError, Response{
					Status:      "error",
					ErrorDetail: "unmarshal reply failed: " + err.Error(),
				})
				return
			}
			log.Printf("✅ Parsed response: %+v", resp)

			// 7️⃣ Return the worker’s response
			statusCode := http.StatusOK
			if resp.Status != "ok" {
				statusCode = http.StatusBadRequest
				log.Printf("⚠ Service returned error status: %s", resp.Status)
			}
			c.JSON(statusCode, resp)
			return
		}
	}
}

func GetInstitutions(c *gin.Context) {
	f, err := os.Open(storagePath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusOK, []UserRequest{})
		return
	} else if err != nil {
		log.Printf("❌ open %s: %v", storagePath, err)
		c.JSON(http.StatusInternalServerError, Response{"error", "", "could not read stored institutions"})
		return
	}
	defer f.Close()

	var list []UserRequest
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var req UserRequest
		if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
			log.Printf("⚠ skipping malformed line: %v", err)
			continue
		}
		list = append(list, req)
	}
	c.JSON(http.StatusOK, list)
}
