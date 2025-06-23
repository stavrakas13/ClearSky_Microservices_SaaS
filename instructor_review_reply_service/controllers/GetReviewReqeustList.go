package controllers

import (
	"encoding/json"
	"fmt"
	"instructor_review_reply_service/db"
)

func GetReviewReqeustList(body map[string]interface{}) (string, error) {
	// input send by orchestrator in json form like:
	// {
	//   "body": {
	//     "course_id": "101"
	//	 }
	// }

	// extract data from input.
	courseID, ok := body["course_id"].(string)
	if !ok {
		return "", fmt.Errorf("missing course_id")
	}
	query := `
		SELECT student_id, course_id, review_created_at 
		FROM reviews 
		WHERE course_id = $1 AND status = 'pending'
		`
	rows, err := db.DB.Query(query, courseID)
	if err != nil {
		return "", fmt.Errorf("query error: %v", err)
	}
	defer rows.Close()

	var requestlist []ReviewSummary
	for rows.Next() {
		var summary ReviewSummary
		err := rows.Scan(&summary.StudentID, &summary.CourseID, &summary.ReviewCreatedAt)
		if err != nil {
			fmt.Println("Scan error:", err)
			continue
		}
		requestlist = append(requestlist, summary)
	}
	if len(requestlist) == 0 {
		emptyResponse := map[string]interface{}{
			"message": "No pending review requests found.",
			"data":    []ReviewSummary{},
		}
		respBytes, _ := json.Marshal(emptyResponse)
		return string(respBytes), nil
	}

	successResponse := map[string]interface{}{
		"message": "Pending review requests retrieved successfully.",
		"data":    requestlist,
	}
	respBytes, err := json.Marshal(successResponse)
	if err != nil {
		return "", fmt.Errorf("failed to marshal response: %v", err)
	}

	return string(respBytes), nil
}

/* func GetReviewReqeustList(c *gin.Context) {

	query := `SELECT student_id, course_id, review_created_at FROM reviews WHERE status = 'pending'`
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
			continue
		}
		requestlist = append(requestlist, summary)
	}
	if len(requestlist) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No pending review requests"})
	} else {
		c.JSON(http.StatusOK, requestlist)
	}
}
*/
