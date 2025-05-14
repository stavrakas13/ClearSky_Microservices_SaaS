package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"student_request_review_service/db"
)

/* func PostNewReviewRequest(c *gin.Context) {
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
	_, err = db.DB.Exec(query, userID, courseID, reqBody.StudentMessage)
	if err != nil {
		log.Println("Insert error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert review"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "Review request submitted successfully.",
		"user_id":         userID,
		"course_id":       courseID,
		"student_message": reqBody.StudentMessage,
	})
} */

func PostNewReviewRequest(_ map[string]string, body map[string]interface{}) (string, error) {
	// input send by orchestrator in json form like:
	//{
	//	"course_id": "101",
	//	"user_id": "42",
	//	"student_message": "Please recheck my assignment."
	//}

	courseIDStr, ok := body["course_id"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid course_id")
	}
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid course_id format")
	}

	userIDFloat, ok := body["user_id"].(float64) // JSON numbers default to float64
	if !ok {
		return "", fmt.Errorf("missing or invalid user_id")
	}
	userID := int(userIDFloat)

	studentMessage, ok := body["student_message"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid student_message")
	}

	query := `INSERT INTO reviews (student_id, course_id, student_message) VALUES ($1, $2, $3)`
	_, err = db.DB.Exec(query, userID, courseID, studentMessage)
	if err != nil {
		fmt.Println("Insert error:", err)
		return "", fmt.Errorf("failed to insert review")
	}

	// Return success response
	response := map[string]interface{}{
		"message":         "Review request submitted successfully.",
		"user_id":         userID,
		"course_id":       courseID,
		"student_message": studentMessage,
	}
	respBytes, _ := json.Marshal(response)
	return string(respBytes), nil
}
