package controllers

import (
	"instructor_review_reply_service/db"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetRequestInfo(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}
	query := `SELECT student_id, course_id, student_message, review_created_at FROM reviews WHERE review_id = $1`
	row := db.DB.QueryRow(query, reviewID)

	var review ReviewStruct

	err = row.Scan(
		&review.Student_id,
		&review.Course_id,
		&review.Student_message,
		&review.Review_created_at,
	)
	log.Println("Scan error:", err)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, review)

}
