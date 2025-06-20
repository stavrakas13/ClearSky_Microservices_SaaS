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
		StudentID string `json:"student_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Marshal the request and publish it to the RPC queue. helperRequest takes
	// care of waiting for the reply and unmarshalling the JSON response.

	payload, _ := json.Marshal(req)
	resp, err := helperRequest(ch, "personal.get_courses", payload)
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
		CourseID  int    `json:"class_id"`
		ExamDate  string `json:"exam_date"`
		StudentID string `json:"student_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, _ := json.Marshal(req)
	resp, err := helperRequest(ch, "personal.get_grades", payload)
	if err != nil {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}
