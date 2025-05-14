package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"student_request_review_service/db"
)

/* func GetReviewStatus(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}
	query := `SELECT student_id, course_id, student_message, status, instructor_reply_message, review_created_at, reviewed_at FROM reviews WHERE review_id = $1`
	row := db.DB.QueryRow(query, reviewID)

	var review ReviewStruct

	err = row.Scan(
		&review.Student_id,
		&review.Course_id,
		&review.Student_message,
		&review.Status,
		&review.Instructor_reply_message,
		&review.Review_created_at,
		&review.Reviewed_at,
	)
	log.Println("Scan error:", err)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}

	c.JSON(http.StatusOK, review)

} */

func GetReviewStatus(params map[string]string) (string, error) {

	// input from orchestrator in json form
	//{
	//	"review_id": "101",
	//}

	reviewIDStr, ok := params["review_id"]
	if !ok {
		return "", fmt.Errorf("missing review_id")
	}
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid review ID")
	}
	// insert into db
	query := `SELECT student_id, course_id, student_message, status, instructor_reply_message, review_created_at, reviewed_at FROM reviews WHERE review_id = $1`
	row := db.DB.QueryRow(query, reviewID)

	var review ReviewStruct
	err = row.Scan(
		&review.Student_id,
		&review.Course_id,
		&review.Student_message,
		&review.Status,
		&review.Instructor_reply_message,
		&review.Review_created_at,
		&review.Reviewed_at,
	)
	if err != nil {
		fmt.Println("Scan error:", err)
		return "", fmt.Errorf("review not found")
	}

	resBytes, _ := json.Marshal(review)
	return string(resBytes), nil
}
