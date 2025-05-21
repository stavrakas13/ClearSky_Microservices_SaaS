package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/xuri/excelize/v2"
)

// Reply coming back from the grades worker
type ExcelUploadResponse struct {
	Status  string `json:"status"`  // "ok" | "error"
	Message string `json:"message"` // free-text
	Error   string `json:"error,omitempty"`
}

// UploadExcelInit – Gin controller
//
// Expects a multipart field named "file" with a .xlsx inside.
// Publishes the workbook to RabbitMQ (base-64 string) and waits up to 10 s
// for a JSON reply from the worker.
func UploadExcelInit(c *gin.Context, ch *amqp.Channel) {
	//------------------------------------------------------------
	// 1) Receive + quick template validation
	//------------------------------------------------------------
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file received"})
		return
	}
	if filepath.Ext(file.Filename) != ".xlsx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only .xlsx files allowed"})
		return
	}

	src, _ := file.Open()
	defer src.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, src); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	// Very light-weight check so we don’t send garbage downstream
	if _, err := excelize.OpenReader(bytes.NewReader(buf.Bytes())); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Excel file"})
		return
	}

	//------------------------------------------------------------
	// 2) Build RPC envelope
	//------------------------------------------------------------
	replyQ, err := ch.QueueDeclare(
		"",    // broker generates a random name
		false, // durable
		true,  // auto-delete
		true,  // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot declare reply queue"})
		return
	}

	msgs, err := ch.Consume(
		replyQ.Name, "", true, true, false, false, nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot consume reply"})
		return
	}

	corrID := uuid.New().String()

	// ----- publish base-64 string as text/plain -----
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())

	if err := ch.Publish(
		"clearSky.event", // <<< same exchange your worker binds to
		"postgrades.init",
		false, false,
		amqp.Publishing{
			ContentType:   "text/plain", // makes the message readable in any CLI
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			MessageId:     file.Filename,
			Body:          []byte(encoded),
		},
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish file"})
		return
	}

	//------------------------------------------------------------
	// 3) Wait for the worker’s reply (10 s timeout)
	//------------------------------------------------------------
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "service timeout"})
			return

		case d := <-msgs:
			if d.CorrelationId != corrID {
				continue // stray message
			}

			var resp ExcelUploadResponse
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid reply format"})
				return
			}

			status := http.StatusOK
			if resp.Status != "ok" {
				status = http.StatusBadRequest
			}
			c.JSON(status, resp)
			return
		}
	}
}
