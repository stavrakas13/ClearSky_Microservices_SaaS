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

	// input send by orchestrator in json form like:
	// {
	//   "params": {
	//     "exam_period": "spring 2025",
	//     "course_id": "101"
	//     "user_id": 42,
	//   }
	// }

	// Extract data
	userIDStr, ok := params["user_id"]
	if !ok {
		return "", fmt.Errorf("missing user_id")
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid user_id format")
	}

	courseIDStr, ok := params["course_id"]
	if !ok {
		return "", fmt.Errorf("missing course_id")
	}
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid course_id format")
	}

	examPeriod, ok := params["exam_period"]
	if !ok {
		return "", fmt.Errorf("missing exam_period")
	}
	// search db using student_id & course_id & exam_period.
	query := `
		SELECT student_id, course_id, exam_period, student_message, status, instructor_reply_message, review_created_at, reviewed_at 
		FROM reviews 
		WHERE student_id = $1 AND course_id = $2 AND exam_period = $3`

	row := db.DB.QueryRow(query, userID, courseID, examPeriod)

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
