package controllers

import (
	"instructor_review_reply_service/db"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// NEED FOR INSTRUCTOR DB
// TO PULL ALL REQUESTS FROM ALL COURCES -> loggedInInstructor.userID

func GetReviewReqeustList(c *gin.Context) {

	query := `SELECT student_id, course_id,  review_created_at FROM reviews WHERE status = 'pending'`
	rows, err := db.DB.Query(query)
	if err != nil {
		log.Println("Query error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	defer rows.Close()
	var requestlist []ReviewSummary
	for rows.Next() {
		var summary ReviewSummary
		err := rows.Scan(&summary.StudentID, &summary.CourseID, &summary.ReviewCreatedAt)
		if err != nil {
			log.Println("Scan error:", err)
			continue // or handle more gracefully
		}
		requestlist = append(requestlist, summary)
	}
	c.JSON(http.StatusOK, requestlist)
}
