package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Helper for RPC via RabbitMQ
func rpcRequest(ch *amqp.Channel, exchange, routingKey string, reqBody interface{}) (map[string]interface{}, error) {
	body, _ := json.Marshal(reqBody)
	corrID := uuid.New().String()
	replyQ, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, err
	}
	msgs, err := ch.Consume(replyQ.Name, "", true, true, false, false, nil)
	if err != nil {
		return nil, err
	}
	err = ch.Publish(
		exchange,
		routingKey,
		false, false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQ.Name,
			Body:          body,
		},
	)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return nil, context.DeadlineExceeded
		case d := <-msgs:
			if d.CorrelationId != corrID {
				continue
			}
			var resp map[string]interface{}
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				return nil, err
			}
			return resp, nil
		}
	}
}

// User Registration
func HandleUserRegister(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Username string `json:"username,omitempty"`
		Password string `json:"password,omitempty"`
		Role     string `json:"role,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	payload := map[string]interface{}{
		"type":     "register",
		"username": req.Username,
		"password": req.Password,
		"role":     req.Role,
	}
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// User Login
func HandleUserLogin(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	payload := map[string]interface{}{
		"type":     "login",
		"username": req.Username,
		"password": req.Password,
	}
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	// Ensure role is present in response
	if role, ok := resp["role"]; !ok || role == "" {
		resp["role"] = resp["Role"] // fallback if capitalized
	}
	c.JSON(http.StatusOK, resp)
}

// User Delete (example, adjust as needed)
func HandleUserDelete(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	payload := map[string]interface{}{
		"type":     "delete",
		"username": req.Username,
	}
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Google Login (calls google_auth_service via RabbitMQ)
func HandleUserGoogleLogin(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	payload := map[string]interface{}{
		"type":  "google_login",
		"token": req.Token,
	}
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// Change user password via RPC
func HandleUserChangePassword(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload := map[string]interface{}{
		"type":         "change_password",
		"username":     req.Username,
		"old_password": req.OldPassword,
		"new_password": req.NewPassword,
	}
	log.Printf("[HandleUserLogin] → publishing RPC payload: %+v\n", payload)
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// HandleUserCreated is a dummy handler that logs and ACKs the message.
func HandleUserCreated(d amqp.Delivery) {
	log.Printf("[Handler] user.created received: %s", string(d.Body))
	// Acknowledge message
	if err := d.Ack(false); err != nil {
		log.Printf("[Handler] Ack failed: %v", err)
	}
}
