package controllers

import (
	"instructor_review_reply_service/db"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func PostReply(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	// Ginstructor reply from request body
	var reqBody InstructorReply
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}
	query := `
		UPDATE reviews 
		SET instructor_reply_message = $1,
			instructor_action = $2,
			status = 'reviewed',
			reviewed_at = CURRENT_TIMESTAMP
		WHERE review_id = $3
	`
	_, err = db.DB.Exec(query, reqBody.InstructorReply, reqBody.InstructorAction, reviewID)
	if err != nil {
		log.Println("Update error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":                  "Reply submitted successfully.",
		"review_id":                reviewID,
		"instructor_reply_message": reqBody.InstructorReply,
		"instructor_action":        reqBody.InstructorAction,
	})
}
