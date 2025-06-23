package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"instructor_review_reply_service/db"
)

// PostReply processes instructor responses for a review request
func PostReply(body map[string]interface{}) (string, error) {
	log.Println("PostReply: invoked with body:", body)

	// extract fields
	courseID, ok := body["course_id"].(string)
	if !ok {
		err := fmt.Errorf("missing or invalid course_id")
		log.Println("PostReply: error extracting course_id:", err)
		return "", err
	}
	userID, ok := body["user_id"].(string)
	if !ok {
		err := fmt.Errorf("missing or invalid user_id")
		log.Println("PostReply: error extracting user_id:", err)
		return "", err
	}
	examPeriod, ok := body["exam_period"].(string)
	if !ok {
		err := fmt.Errorf("missing or invalid exam_period")
		log.Println("PostReply: error extracting exam_period:", err)
		return "", err
	}

	instructorReply, ok := body["instructor_reply_message"].(string)
	if !ok {
		err := fmt.Errorf("missing or invalid instructor_reply_message")
		log.Println("PostReply: error extracting instructor_reply_message:", err)
		return "", err
	}
	instructorAction, ok := body["instructor_action"].(string)
	if !ok {
		err := fmt.Errorf("missing or invalid instructor_action")
		log.Println("PostReply: error extracting instructor_action:", err)
		return "", err
	}

	log.Printf("PostReply: updating review for student_id=%s, course_id=%s, exam_period=%s", userID, courseID, examPeriod)
	// update statement
	query := `
		UPDATE reviews 
		SET instructor_reply_message = $1,
			instructor_action = $2,
			status = 'reviewed',
			reviewed_at = CURRENT_TIMESTAMP
		WHERE student_id = $3 AND course_id = $4 AND exam_period = $5	
	`
	result, err := db.DB.Exec(query, instructorReply, instructorAction, userID, courseID, examPeriod)
	if err != nil {
		log.Printf("PostReply: update error: %v", err)
		return "", fmt.Errorf("failed to update review: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	log.Printf("PostReply: rows affected: %d", rowsAffected)
	if rowsAffected == 0 {
		log.Println("PostReply: no rows updated, possible invalid identifiers")
		failResponse := map[string]interface{}{ 
			"message": "Failed to update instructor response in database on student end.",
		}
		failRespBytes, _ := json.Marshal(failResponse) // nolint: errcheck
		log.Println("PostReply: returning failure response to orchestrator")
		return string(failRespBytes), nil
	}

	successResponse := map[string]interface{}{ 
		"message": "Instructor response updated successfully on instructor end.",
	}
	respBytes, err := json.Marshal(successResponse)
	if err != nil {
		log.Printf("PostReply: response marshal error: %v", err)
		return "", fmt.Errorf("failed to marshal response: %v", err)
	}
	log.Println("PostReply: returning success response to orchestrator")
	return string(respBytes), nil
}
