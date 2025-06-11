package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/gin-gonic/gin"
    amqp "github.com/rabbitmq/amqp091-go"
)

// HandleGetStudentCourses retrieves the student's courses and their status via RabbitMQ.
func HandleGetStudentCourses(c *gin.Context, ch *amqp.Channel) {
    var req struct {
        StudentID string `json:"student_id"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    payload, _ := json.Marshal(req)
    resp, err := helperRequest(ch, "personal.get_courses", payload)
    if err != nil {
        c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"data": resp})
}

// HandleGetPersonalGrades fetches a student's grades for a specific exam.
func HandleGetPersonalGrades(c *gin.Context, ch *amqp.Channel) {
    var req struct {
        ClassID   string `json:"class_id"`
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
