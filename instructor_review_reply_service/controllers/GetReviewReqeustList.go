package controllers

import (
	"encoding/json"
	"fmt"
	"instructor_review_reply_service/db"
	"log"
)

func GetReviewRequestList(body map[string]interface{}) (string, error) {
	log.Println("GetReviewRequestList: invoked with body:", body)

	// extract data from input. -> username FROM JWT -> LOGGED IN USER INSTRUCTOR
	username, ok := body["username"].(string)
	if !ok {
		err := fmt.Errorf("missing or invalid username")
		log.Println("GetReviewRequestList: error extracting username:", err)
		return "", err
	}

	// search logged-in user's courses

	log.Printf("GetReviewRequestList: querying course_id for instructor_id=%s", username)

	q := `
		SELECT course_id
		FROM instructors
		WHERE instructor_name = $1
	`
	rows, err := db.DB.Query(q, username)
	if err != nil {
		log.Printf("GetReviewRequestList: course_id query error: %v", err)
		return "", fmt.Errorf("course_id query error: %v", err)
	}
	defer rows.Close()

	var courseIDs []string
	for rows.Next() {
		var cid string
		if err := rows.Scan(&cid); err != nil {
			log.Printf("GetReviewRequestList: course_id scan error: %v", err)
			continue
		}
		courseIDs = append(courseIDs, cid)
	}
	if err := rows.Err(); err != nil {
		log.Printf("GetReviewRequestList: course_id rows iteration error: %v", err)
	}

	if len(courseIDs) == 0 {
		log.Printf("GetReviewRequestList: no courses found for instructor_id=%s", username)
		emptyResponse := map[string]interface{}{
			"message": "No courses found for the provided instructor ID.",
			"data":    []ReviewSummary{},
		}
		respBytes, _ := json.Marshal(emptyResponse)
		return string(respBytes), nil
	}

	log.Printf("GetReviewRequestList: found %d course(s): %v", len(courseIDs), courseIDs)

	// go through courses_id and check for pending reviews.

	var requestList []ReviewSummary
	reviewQuery := `
		SELECT student_id, course_id, exam_period
		FROM reviews
		WHERE course_id = $1 AND status = 'pending'
	`

	for _, courseID := range courseIDs {
		log.Printf("GetReviewRequestList: querying pending reviews for course_id=%s", courseID)

		reviewRows, err := db.DB.Query(reviewQuery, courseID)
		if err != nil {
			log.Printf("GetReviewRequestList: review query error for course_id=%s: %v", courseID, err)
			continue
		}

		for reviewRows.Next() {
			var summary ReviewSummary
			err := reviewRows.Scan(&summary.StudentID, &summary.CourseID, &summary.Exam_period)
			if err != nil {
				log.Printf("GetReviewRequestList: scan error: %v", err)
				continue
			}
			requestList = append(requestList, summary)
			log.Printf("GetReviewRequestList: found request: %+v", summary)
		}
		if err := reviewRows.Err(); err != nil {
			log.Printf("GetReviewRequestList: rows iteration error for course_id=%s: %v", courseID, err)
		}
		reviewRows.Close()
	}

	if len(requestList) == 0 {
		log.Println("GetReviewRequestList: no pending review requests found across all courses")
		emptyResponse := map[string]interface{}{
			"message": "No pending review requests found.",
			"data":    []ReviewSummary{},
		}
		respBytes, _ := json.Marshal(emptyResponse)
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

/* EXAMPLE INPUT:

{
  "username": "instructor"
}

EXAMPLE OUTPUT:

{
  "message": "Pending review requests retrieved successfully.",
  "data": [
    {
      "student_id": "student_a",
      "course_id": "course_1",
      "review_created_at": "2025-06-20T15:04:05Z"
    },
    {
      "student_id": "student_b",
      "course_id": "course_2",
      "review_created_at": "2025-06-21T10:15:30Z"
    }
  ]
}

OR

{
  "message": "No pending review requests found.",
  "data": []
}

OR

{
  "message": "No courses found for the provided instructor ID.",
  "data": []
}

*/
