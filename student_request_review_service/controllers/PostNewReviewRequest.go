package controllers

import (
	"log"
	"net/http"
	"strconv"
	"student_request_review_service/db"

	"github.com/gin-gonic/gin"
)

func PostNewReviewRequest(c *gin.Context) {
	courseIDStr := c.Param("course_id")
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid course ID"})
		return
	}

	var reqBody ReviewRequest
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	query := `INSERT INTO reviews (student_id, course_id, student_message) VALUES ($1, $2, $3)`
	_, err = db.DB.Exec(query, loggedInStudent.userID, courseID, reqBody.StudentMessage)
	if err != nil {
		log.Println("Insert error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert review"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Review request submitted successfully.",
		"user_id":         loggedInStudent.userID,
		"course_id":       courseID,
		"student_message": reqBody.StudentMessage,
	})
}
