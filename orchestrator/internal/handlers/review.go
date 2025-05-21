package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

// Assume all data received from json.

// helperRequest sends the payload to the given routing key on ExchangeKey and waits for a JSON response
func helperRequest(ch *amqp.Channel, routingKey string, payload []byte) (map[string]interface{}, error) {
	fmt.Printf("Outgoing payload: %s\n routing key: %s\n", payload, routingKey)

	corrID := "abc123"
	replyQueue, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return nil, err
	}

	msgs, err := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	err = ch.Publish(
		"clearSky.events", // publish to the direct exchange
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:   "application/json",
			CorrelationId: corrID,
			ReplyTo:       replyQueue.Name,
			Body:          payload,
		},
	)
	if err != nil {
		return nil, err
	}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case msg := <-msgs:
			if msg.CorrelationId == corrID {
				var response map[string]interface{}
				err := json.Unmarshal(msg.Body, &response)
				return response, err
			}
		case <-timeout:
			return nil, errors.New("timeout waiting for response")
		}
	}
}

// handleReviewRequested processes new request events
// -> sends 2 events: student.postNewRequest & instructor.insertStudentRequest
func HandlePostNewRequest(c *gin.Context, ch *amqp.Channel) {
	// receive message
	var req struct {
		CourseID       int    `json:"course_id"`
		UserID         int    `json:"user_id"`
		StudentMessage string `json:"student_message"`
		ExamPeriod     string `json:"exam_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// add message payload
	payload, _ := json.Marshal(map[string]interface{}{
		"body": map[string]interface{}{
			"exam_period":     req.ExamPeriod,
			"course_id":       req.CourseID,
			"user_id":         req.UserID,
			"student_message": req.StudentMessage,
		},
	})

	responseStudent, err := helperRequest(ch, "student.postNewRequest", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": responseStudent})
	responseInstructor, err := helperRequest(ch, "instructor.insertStudentRequest", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": responseInstructor})
}

// handleGetRequestStatus processes student sees request status events
// -> sends 1 event: student.getRequestStatus
func HandleGetRequestStatus(c *gin.Context, ch *amqp.Channel) {
	// receive message
	var req struct {
		CourseID   int    `json:"course_id"`
		UserID     int    `json:"user_id"`
		ExamPeriod string `json:"exam_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// add message payload
	payload, _ := json.Marshal(map[string]interface{}{
		"body": map[string]interface{}{
			"exam_period": req.ExamPeriod,
			"course_id":   req.CourseID,
			"user_id":     req.UserID,
		},
	})
	responseStudent, err := helperRequest(ch, "student.getRequestStatus", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": responseStudent})
}

// handlePostResponse processes responses on review requests
// -> sends 2 events: student.updateInstructorResponse & instructor.postResponse
func HandlePostResponse(c *gin.Context, ch *amqp.Channel) {
	// receive message
	var req struct {
		CourseID               int    `json:"course_id"`
		UserID                 int    `json:"user_id"`
		ExamPeriod             string `json:"exam_period"`
		InstructorReplyMessage string `json:"instructor_reply_message"`
		InstructorAction       string `json:"instructor_action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// add message payload
	payload, _ := json.Marshal(map[string]interface{}{
		"body": map[string]interface{}{
			"exam_period":              req.ExamPeriod,
			"course_id":                req.CourseID,
			"user_id":                  req.UserID,
			"instructor_reply_message": req.InstructorReplyMessage,
			"instructor_action":        req.InstructorAction,
		},
	})

	responseStudent, err := helperRequest(ch, "student.updateInstructorResponse", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": responseStudent})
	responseInstructor, err := helperRequest(ch, "instructor.postResponse", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": responseInstructor})
}

// handleGetRequestList processes instructor get list of pending requests
// -> sends 1 event: instructor.getRequestsList
func HandleGetRequestList(c *gin.Context, ch *amqp.Channel) {
	// receive message
	var req struct {
		CourseID   int    `json:"course_id"`
		ExamPeriod string `json:"exam_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// add message payload
	payload, _ := json.Marshal(map[string]interface{}{
		"body": map[string]interface{}{
			"exam_period": req.ExamPeriod,
			"course_id":   req.CourseID,
		},
	})
	responseInstructor, err := helperRequest(ch, "instructor.getRequestsList", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": responseInstructor})
}

// handleGetRequestInfo processes instructor sees request details
// -> sends 1 event: instructor.getRequestInfo
func HandleGetRequestInfo(c *gin.Context, ch *amqp.Channel) {
	// receive message
	var req struct {
		CourseID   int    `json:"course_id"`
		UserID     int    `json:"user_id"`
		ExamPeriod string `json:"exam_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// add message payload
	payload, _ := json.Marshal(map[string]interface{}{
		"body": map[string]interface{}{
			"exam_period": req.ExamPeriod,
			"course_id":   req.CourseID,
			"user_id":     req.UserID,
		},
	})
	responseInstructor, err := helperRequest(ch, "instructor.getRequestInfo", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": responseInstructor})
}
