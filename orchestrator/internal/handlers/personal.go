package handlers

// RabbitMQ handlers for retrieving personal grades information. The orchestrator
// exposes HTTP endpoints that forward the requests to the personal grades
// service using RPC-style messaging.

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

// When upload grades, update view grades too.

func UploadInitGradesInViewGrades(c *gin.Context, ch *amqp.Channel)  {}
func UploadFinalGradesInViewGrades(c *gin.Context, ch *amqp.Channel) {}

// HandleGetStudentCourses receives a JSON body containing a `student_id`,
// forwards the request over RabbitMQ and returns whatever payload the personal
// grades service responds with.

func HandleGetStudentCourses(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		CourseID int    `json:"class_id"`
		ExamDate string `json:"exam_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID := c.GetString("user_id") // from JWT middleware
	payload := map[string]interface{}{  // build own‐ID payload
		"student_id": studentID,
	}
	body, _ := json.Marshal(payload)
	resp, err := helperRequest(ch, "personal.get_courses", body)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

// HandleGetPersonalGrades does the same as HandleGetStudentCourses but expects
// class ID and exam date to look up the student's grades in a particular exam.

func HandleGetPersonalGrades(c *gin.Context, ch *amqp.Channel) {
	var req struct {
		CourseID int    `json:"class_id"`
		ExamDate string `json:"exam_date"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	studentID := c.GetString("user_id") // enforce own ID
	payload := map[string]interface{}{
		"class_id":   req.CourseID,
		"exam_date":  req.ExamDate,
		"student_id": studentID,
	}
	body, _ := json.Marshal(payload)
	resp, err := helperRequest(ch, "personal.get_grades", body)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}
