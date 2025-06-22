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

	log.Printf("[RPC] Preparing request → Exchange: %q, RoutingKey: %q, CorrID: %s, Payload: %s", exchange, routingKey, corrID, string(body))

	replyQ, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		log.Printf("[RPC] Failed to declare reply queue: %v", err)
		return nil, err
	}

	msgs, err := ch.Consume(replyQ.Name, "", true, true, false, false, nil)
	if err != nil {
		log.Printf("[RPC] Failed to consume from reply queue: %v", err)
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
		log.Printf("[RPC] Failed to publish message: %v", err)
		return nil, err
	}

	log.Printf("[RPC] Published message. Awaiting response...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			log.Printf("[RPC] Timeout waiting for response for CorrID: %s", corrID)
			return nil, context.DeadlineExceeded
		case d := <-msgs:
			if d.CorrelationId != corrID {
				log.Printf("[RPC] Skipping unrelated CorrID: %s", d.CorrelationId)
				continue
			}
			var resp map[string]interface{}
			if err := json.Unmarshal(d.Body, &resp); err != nil {
				log.Printf("[RPC] Failed to unmarshal response: %v", err)
				return nil, err
			}
			log.Printf("[RPC] Received response for CorrID %s: %s", corrID, string(d.Body))
			return resp, nil
		}
	}
}

// User Registration
func HandleUserRegister(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     string `json:"role,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Register] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password are required"})
		return
	}

	if req.Role == "" {
		req.Role = "student"
	}
	log.Printf("[Register] Registering user: %s with role: %s", req.Username, req.Role)

	payload := map[string]interface{}{
		"type":     "register",
		"username": req.Username,
		"password": req.Password,
		"role":     req.Role,
	}
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		log.Printf("[Register] RPC error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[Register] Registration response: %+v", resp)
	c.JSON(http.StatusOK, resp)
}

// User Login
func HandleUserLogin(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Login] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	log.Printf("[Login] Logging in user: %s", req.Username)

	payload := map[string]interface{}{
		"type":     "login",
		"username": req.Username,
		"password": req.Password,
	}
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		log.Printf("[Login] RPC error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	if role, ok := resp["role"]; !ok || role == "" {
		resp["role"] = resp["Role"]
	}
	log.Printf("[Login] Login response: %+v", resp)
	c.JSON(http.StatusOK, resp)
}

// User Delete
func HandleUserDelete(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Username string `json:"username"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[Delete] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	log.Printf("[Delete] Deleting user: %s", req.Username)

	payload := map[string]interface{}{
		"type":     "delete",
		"username": req.Username,
	}
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		log.Printf("[Delete] RPC error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[Delete] Deletion response: %+v", resp)
	c.JSON(http.StatusOK, resp)
}

// Google Login
func HandleUserGoogleLogin(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Token string `json:"token"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[GoogleLogin] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	log.Printf("[GoogleLogin] Attempting Google login with token.")

	payload := map[string]interface{}{
		"type":  "google_login",
		"token": req.Token,
	}
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		log.Printf("[GoogleLogin] RPC error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[GoogleLogin] Login response: %+v", resp)
	c.JSON(http.StatusOK, resp)
}

// Change Password
func HandleUserChangePassword(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		Username    string `json:"username" binding:"required"`
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[ChangePassword] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[ChangePassword] Changing password for user: %s", req.Username)

	payload := map[string]interface{}{
		"type":         "change_password",
		"username":     req.Username,
		"old_password": req.OldPassword,
		"new_password": req.NewPassword,
	}
	log.Printf("[ChangePassword] Publishing RPC payload: %+v", payload)
	resp, err := rpcRequest(ch, "", "auth.request", payload)
	if err != nil {
		log.Printf("[ChangePassword] RPC error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("[ChangePassword] Response: %+v", resp)
	c.JSON(http.StatusOK, resp)
}

// Dummy consumer handler
func HandleUserCreated(d amqp.Delivery) {
	log.Printf("[Handler] user.created event received: %s", string(d.Body))
	if err := d.Ack(false); err != nil {
		log.Printf("[Handler] Failed to ACK message: %v", err)
	}
}
