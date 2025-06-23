package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"instructor_review_reply_service/db"
)

func GetReviewRequestList(body map[string]interface{}) (string, error) {
	log.Println("GetReviewRequestList: invoked with body:", body)

	// extract data from input.
	courseID, ok := body["course_id"].(string)
	if !ok {
		err := fmt.Errorf("missing or invalid course_id")
		log.Println("GetReviewRequestList: error extracting course_id:", err)
		return "", err
	}
	log.Printf("GetReviewRequestList: querying pending reviews for course_id=%s", courseID)

	query := `
		SELECT student_id, course_id, review_created_at
		FROM reviews
		WHERE course_id = $1 AND status = 'pending'
	`
	rows, err := db.DB.Query(query, courseID)
	if err != nil {
		log.Printf("GetReviewRequestList: query error: %v", err)
		return "", fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	var requestList []ReviewSummary
	for rows.Next() {
		var summary ReviewSummary
		err := rows.Scan(&summary.StudentID, &summary.CourseID, &summary.ReviewCreatedAt)
		if err != nil {
			log.Printf("GetReviewRequestList: scan error: %v", err)
			continue
		}
		requestList = append(requestList, summary)
		log.Printf("GetReviewRequestList: found request: %+v", summary)
	}

	if err = rows.Err(); err != nil {
		log.Printf("GetReviewRequestList: rows iteration error: %v", err)
	}

	if len(requestList) == 0 {
		log.Println("GetReviewRequestList: no pending requests found")
		emptyResponse := map[string]interface{}{ 
			"message": "No pending review requests found.",
			"data":    []ReviewSummary{},
		}
		respBytes, _ := json.Marshal(emptyResponse) // nolint: errcheck
		log.Println("GetReviewRequestList: returning empty response")
		return string(respBytes), nil
	}

	log.Printf("GetReviewRequestList: total pending requests: %d", len(requestList))
	successResponse := map[string]interface{}{ 
		"message": "Pending review requests retrieved successfully.",
		"data":    requestList,
	}
	respBytes, err := json.Marshal(successResponse)
	if err != nil {
		log.Printf("GetReviewRequestList: response marshal error: %v", err)
		return "", fmt.Errorf("failed to marshal response: %v", err)
	}
	log.Println("GetReviewRequestList: returning success response")

	return string(respBytes), nil
}
