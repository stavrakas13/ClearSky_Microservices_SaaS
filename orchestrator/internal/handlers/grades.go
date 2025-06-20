package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
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
		"clearSky.events", // <<< same exchange your worker binds to
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
			ForwardToStatistics(ch, buf.Bytes(), file.Filename) //update statistics ms
			c.JSON(status, resp)
			return
		}
	}
}

func UploadExcelFinal(c *gin.Context, ch *amqp.Channel) {
	log.Println("[UploadExcelFinal] Receiving file...")

	// 1) Receive + quick template validation
	file, err := c.FormFile("file")
	if err != nil {
		log.Println("[UploadExcelFinal] No file received")
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file received"})
		return
	}

	if filepath.Ext(file.Filename) != ".xlsx" {
		log.Printf("[UploadExcelFinal] Invalid file extension: %s\n", file.Filename)
		c.JSON(http.StatusBadRequest, gin.H{"error": "only .xlsx files allowed"})
		return
	}

	log.Printf("[UploadExcelFinal] Validating file: %s\n", file.Filename)
	src, _ := file.Open()
	defer src.Close()

	var buf bytes.Buffer
	if _, err = io.Copy(&buf, src); err != nil {
		log.Printf("[UploadExcelFinal] Failed to read file: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	if _, err := excelize.OpenReader(bytes.NewReader(buf.Bytes())); err != nil {
		log.Printf("[UploadExcelFinal] Excel validation failed: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid Excel file"})
		return
	}

	// 2) Build RPC envelope
	log.Println("[UploadExcelFinal] Declaring reply queue...")
	replyQ, err := ch.QueueDeclare(
		"", false, true, true, false, nil,
	)
	if err != nil {
		log.Printf("[UploadExcelFinal] Failed to declare reply queue: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot declare reply queue"})
		return
	}

	log.Printf("[UploadExcelFinal] Consuming from reply queue: %s\n", replyQ.Name)
	msgs, err := ch.Consume(
		replyQ.Name, "", true, true, false, false, nil,
	)
	if err != nil {
		log.Printf("[UploadExcelFinal] Failed to consume from reply queue: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot consume reply"})
		return
	}

	corrID := uuid.New().String()
	log.Printf("[UploadExcelFinal] Correlation ID: %s\n", corrID)

	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	log.Println("[UploadExcelFinal] Publishing file to postgrades.final...")

	if err := ch.Publish(
		"clearSky.events",
		"postgrades.final",
		false, false,
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			MessageId:     file.Filename,
			Body:          []byte(encoded),
		},
	); err != nil {
		log.Printf("[UploadExcelFinal] Failed to publish message: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish file"})
		return
	}

	log.Println("[UploadExcelFinal] Waiting for reply from worker...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Println("[UploadExcelFinal] Timeout waiting for reply")
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "service timeout"})
			return

		case d := <-msgs:
			if d.CorrelationId != corrID {
				log.Println("[UploadExcelFinal] Ignoring unrelated message")
				continue
			}

			log.Println("[UploadExcelFinal] Received reply from worker")
			var resp ExcelUploadResponse
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				log.Printf("[UploadExcelFinal] Failed to unmarshal reply: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid reply format"})
				return
			}

			log.Println("[UploadExcelFinal] Calling HandleCreditsSpent...")
			if err := HandleCreditsSpent(ch); err != nil { //update credits ms
				log.Printf("[UploadExcelFinal] Failed to publish credits spent: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"status":  "error",
					"message": "failed to publish credits spent",
					"error":   err.Error(),
				})
				return
			}

			log.Println("[UploadExcelFinal] Upload successful, credits deducted")
			ForwardToStatistics(ch, buf.Bytes(), file.Filename) //update statistics ms
			c.JSON(http.StatusOK, gin.H{
				"status":  resp.Status,
				"message": "final grades uploaded and credits deducted",
				"details": resp,
			})
			return
		}
	}
}
