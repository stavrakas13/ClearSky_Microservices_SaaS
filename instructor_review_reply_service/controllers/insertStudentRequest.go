package controllers

import (
	"encoding/json"
	"fmt"
	"instructor_review_reply_service/db"
	"strconv"
)

func InsertStudentRequest(params map[string]string, body map[string]interface{}) (string, error) {

	// input send by orchestrator in json form like:
	// {
	//   "params": {
	//     "exam_period": "spring 2025",
	//     "course_id": "101"
	//   },
	//   "body": {
	//     "user_id": 42,
	//     "student_message": "Please recheck my assignment."
	//   }
	// }
	// extract data from input.
	courseIDStr, ok := params["course_id"]
	if !ok {
		return "", fmt.Errorf("missing or invalid course_id")
	}

	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid course_id format")
	}

	examPeriod, ok := params["exam_period"]
	if !ok {
		return "", fmt.Errorf("missing exam_period")
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

	// add review to db
	query := `INSERT INTO reviews (student_id, course_id, exam_period, student_message) VALUES ($1, $2, $3, $4)`
	_, err = db.DB.Exec(query, userID, courseID, examPeriod, studentMessage)
	if err != nil {
		fmt.Println("Insert error:", err)
		return "", fmt.Errorf("failed to insert review")
	}

	// Return success response
	response := map[string]interface{}{
		"message":         "Review request submitted successfully.",
		"user_id":         userID,
		"course_id":       courseID,
		"exam_period":     examPeriod,
		"student_message": studentMessage,
	}
	respBytes, _ := json.Marshal(response)
	return string(respBytes), nil
}
