package controllers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"student_request_review_service/db"
)

func UpdateInstructorResponse(params map[string]string, body map[string]interface{}) (string, error) {
	// input send by orchestrator in json form like:
	//{
	//	"params": {
	//	  "course_id": "101",
	//	  "exam_period": "spring 2025",
	//	  "user_id": "42"
	//	},
	//	"body": {
	//		"instructor_reply_message": "NO WAY!"
	//		"instructor_action": "Denied"
	//	}
	//}

	// get data from json
	courseIDStr, ok := params["course_id"]
	if !ok {
		return "", fmt.Errorf("missing course_id")
	}
	courseID, err := strconv.Atoi(courseIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid course_id format")
	}

	userIDStr, ok := params["user_id"]
	if !ok {
		return "", fmt.Errorf("missing user_id")
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return "", fmt.Errorf("invalid user_id format")
	}

	examPeriod, ok := params["exam_period"]
	if !ok {
		return "", fmt.Errorf("missing exam_period")
	}

	instructorReply, ok := body["instructor_reply_message"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid instructor_reply_message")
	}

	instructorAction, ok := body["instructor_action"].(string)
	if !ok {
		return "", fmt.Errorf("missing or invalid instructor_action")
	}

	query := `
		UPDATE reviews 
		SET instructor_reply_message = $1,
			instructor_action = $2,
			status = 'reviewed',
			reviewed_at = CURRENT_TIMESTAMP
		WHERE student_id = $3 AND course_id = $4 AND exam_period = $5	
	`

	_, err = db.DB.Exec(query, instructorReply, instructorAction, userID, courseID, examPeriod)
	if err != nil {
		return "", fmt.Errorf("failed to update review: %v", err)
	}

	response := map[string]interface{}{
		"message":                  "Instructor response updated successfully.",
		"user_id":                  userID,
		"course_id":                courseID,
		"exam_period":              examPeriod,
		"instructor_action":        instructorAction,
		"instructor_reply_message": instructorReply,
	}
	respBytes, _ := json.Marshal(response)
	return string(respBytes), nil
}
