package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"orchestrator/internal/middleware"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

// helperRequest sends the payload to the given routing key on ExchangeKey and waits for a JSON response
func helperRequest(ch *amqp.Channel, routingKey string, payload []byte) (map[string]interface{}, error) {
	log.Printf("helperRequest: routingKey=%s, payload=%s", routingKey, payload)
	
	corrID := "abc123"
	replyQueue, err := ch.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		log.Printf("helperRequest: QueueDeclare error: %v", err)
		return nil, err
	}

	msgs, err := ch.Consume(replyQueue.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Printf("helperRequest: Consume error: %v", err)
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
		log.Printf("helperRequest: Publish error: %v", err)
		return nil, err
	}

	timeout := time.After(5 * time.Second)
	for {
		select {
		case msg := <-msgs:
			if msg.CorrelationId == corrID {
				var response map[string]interface{}
				err := json.Unmarshal(msg.Body, &response)
				if err != nil {
					log.Printf("helperRequest: Unmarshal error: %v", err)
				}
				log.Printf("helperRequest: received response: %+v", response)
				return response, err
			}
		case <-timeout:
			log.Printf("helperRequest: timeout waiting for response")
			return nil, errors.New("timeout waiting for response")
		}
	}
}

// HandlePostNewRequest processes new request events
// -> sends 2 events: student.postNewRequest & instructor.insertStudentRequest
func HandlePostNewRequest(c *gin.Context, ch *amqp.Channel) {
	log.Printf("HandlePostNewRequest invoked")

	// Get student info from JWT using middleware helpers
	studentID := middleware.GetStudentID(c)
	userID := middleware.GetStudentID(c)

	if !middleware.IsStudent(c) {
		log.Printf("HandlePostNewRequest: forbidden, not a student")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only students can submit review requests"})
		return
	}

	if studentID == "" {
		log.Printf("HandlePostNewRequest: missing student ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Student ID is required for review requests"})
		return
	}

	var req struct {
		CourseID       string `json:"course_id"`
		StudentMessage string `json:"student_message"`
		ExamPeriod     string `json:"exam_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HandlePostNewRequest: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	log.Printf("HandlePostNewRequest: payload struct %+v", req)

	payload, _ := json.Marshal(map[string]interface{}{ // nolint: errcheck
		"body": map[string]interface{}{
			"exam_period":     req.ExamPeriod,
			"course_id":       req.CourseID,
			"user_id":         userID,
			"student_id":      studentID,
			"student_message": req.StudentMessage,
		},
	})

	responseStudent, err := helperRequest(ch, "student.postNewRequest", payload)
	if err != nil {
		log.Printf("HandlePostNewRequest: student.postNewRequest error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("HandlePostNewRequest: responseStudent %+v", responseStudent)
	c.JSON(http.StatusOK, gin.H{"data": responseStudent})

	responseInstructor, err := helperRequest(ch, "instructor.insertStudentRequest", payload)
	if err != nil {
		log.Printf("HandlePostNewRequest: instructor.insertStudentRequest error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("HandlePostNewRequest: responseInstructor %+v", responseInstructor)
	c.JSON(http.StatusOK, gin.H{"data": responseInstructor})
}

// HandleGetRequestStatus processes student sees request status events
// -> sends 1 event: student.getRequestStatus
func HandleGetRequestStatus(c *gin.Context, ch *amqp.Channel) {
	log.Printf("HandleGetRequestStatus invoked")

	studentID := middleware.GetStudentID(c)
	userID := middleware.GetStudentID(c)

	if !middleware.IsStudent(c) {
		log.Printf("HandleGetRequestStatus: forbidden, not a student")
		c.JSON(http.StatusForbidden, gin.H{"error": "Only students can check request status"})
		return
	}

	var req struct {
		CourseID   string `json:"course_id"`
		ExamPeriod string `json:"exam_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HandleGetRequestStatus: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	log.Printf("HandleGetRequestStatus: payload struct %+v", req)

	payload, _ := json.Marshal(map[string]interface{}{ // nolint: errcheck
		"body": map[string]interface{}{
			"exam_period": req.ExamPeriod,
			"course_id":   req.CourseID,
			"user_id":     userID,
			"student_id":  studentID,
		},
	})

	responseStudent, err := helperRequest(ch, "student.getRequestStatus", payload)
	if err != nil {
		log.Printf("HandleGetRequestStatus: error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("HandleGetRequestStatus: responseStudent %+v", responseStudent)
	c.JSON(http.StatusOK, gin.H{"data": responseStudent})
}

// HandlePostResponse processes responses on review requests
// -> sends 2 events: student.updateInstructorResponse & instructor.postResponse
func HandlePostResponse(c *gin.Context, ch *amqp.Channel) {
	log.Printf("HandlePostResponse invoked")

	var req struct {
		CourseID               string `json:"course_id"`
		UserID                 string `json:"user_id"`
		ExamPeriod             string `json:"exam_period"`
		InstructorReplyMessage string `json:"instructor_reply_message"`
		InstructorAction       string `json:"instructor_action"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HandlePostResponse: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	log.Printf("HandlePostResponse: payload struct %+v", req)

	payload, _ := json.Marshal(map[string]interface{}{ // nolint: errcheck
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
		log.Printf("HandlePostResponse: student.updateInstructorResponse error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("HandlePostResponse: responseStudent %+v", responseStudent)
	c.JSON(http.StatusOK, gin.H{"data": responseStudent})

	responseInstructor, err := helperRequest(ch, "instructor.postResponse", payload)
	if err != nil {
		log.Printf("HandlePostResponse: instructor.postResponse error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("HandlePostResponse: responseInstructor %+v", responseInstructor)
	c.JSON(http.StatusOK, gin.H{"data": responseInstructor})
}

// HandleGetRequestList processes instructor get list of pending requests
// -> sends 1 event: instructor.getRequestsList
func HandleGetRequestList(c *gin.Context, ch *amqp.Channel) {
	log.Printf("HandleGetRequestList invoked")

	var req struct {
		CourseID   string `json:"course_id"`
		ExamPeriod string `json:"exam_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HandleGetRequestList: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	log.Printf("HandleGetRequestList: payload struct %+v", req)

	payload, _ := json.Marshal(map[string]interface{}{ // nolint: errcheck
		"body": map[string]interface{}{
			"exam_period": req.ExamPeriod,
			"course_id":   req.CourseID,
		},
	})

	responseInstructor, err := helperRequest(ch, "instructor.getRequestsList", payload)
	if err != nil {
		log.Printf("HandleGetRequestList: error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("HandleGetRequestList: responseInstructor %+v", responseInstructor)
	c.JSON(http.StatusOK, gin.H{"data": responseInstructor})
}

// HandleGetRequestInfo processes instructor sees request details
// -> sends 1 event: instructor.getRequestInfo
func HandleGetRequestInfo(c *gin.Context, ch *amqp.Channel) {
	log.Printf("HandleGetRequestInfo invoked")

	var req struct {
		CourseID   string `json:"course_id"`
		UserID     string `json:"user_id"`
		ExamPeriod string `json:"exam_period"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("HandleGetRequestInfo: bind error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	log.Printf("HandleGetRequestInfo: payload struct %+v", req)

	payload, _ := json.Marshal(map[string]interface{}{ // nolint: errcheck
		"body": map[string]interface{}{
			"exam_period": req.ExamPeriod,
			"course_id":   req.CourseID,
			"user_id":     req.UserID,
		},
	})

	responseInstructor, err := helperRequest(ch, "instructor.getRequestInfo", payload)
	if err != nil {
		log.Printf("HandleGetRequestInfo: error: %v", err)
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	log.Printf("HandleGetRequestInfo: responseInstructor %+v", responseInstructor)
	c.JSON(http.StatusOK, gin.H{"data": responseInstructor})
}
