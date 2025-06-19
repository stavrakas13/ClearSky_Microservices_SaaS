package controllers

import (
	"encoding/json"
	"fmt"
	"instructor_review_reply_service/db"
)

func PostReply(body map[string]interface{}) (string, error) {

	// input send by orchestrator in json form like:
	//{
	//	"body": {
	//	  "course_id": "101",
	//	  "exam_period": "spring 2025",
	//	  "user_id": "42"
	//	  "instructor_reply_message": "NO WAY!"
	//	  "instructor_action": "Denied"
	//	}
	//}

	courseID, ok := body["course_id"]
	if !ok {
		return "", fmt.Errorf("missing course_id")
	}

	userID, ok := body["user_id"].(string)
	if !ok {
		return "", fmt.Errorf("missing user_id")
	}

	examPeriod, ok := body["exam_period"].(string)
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

	result, err := db.DB.Exec(query, instructorReply, instructorAction, userID, courseID, examPeriod)
	if err != nil {
		return "", fmt.Errorf("failed to update review: %v", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		failResponse := map[string]interface{}{
			"message": "Failed to update instructor response in database on student end.",
		}
		failRespBytes, _ := json.Marshal(failResponse)
		return string(failRespBytes), nil
	}

	response := map[string]interface{}{
		"message": "Instructor response updated successfully on instructor end.",
	}
	respBytes, _ := json.Marshal(response)
	return string(respBytes), nil

}

/* func PostReply(c *gin.Context) {
	reviewIDStr := c.Param("review_id")
	reviewID, err := strconv.Atoi(reviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid review ID"})
		return
	}

	// Ginstructor reply from request body
	var reqBody InstructorReply
	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}
	query := `
		UPDATE reviews
		SET instructor_reply_message = $1,
			instructor_action = $2,
			status = 'reviewed',
			reviewed_at = CURRENT_TIMESTAMP
		WHERE review_id = $3
	`
	_, err = db.DB.Exec(query, reqBody.InstructorReply, reqBody.InstructorAction, reviewID)
	if err != nil {
		log.Println("Update error:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update review"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":                  "Reply submitted successfully.",
		"review_id":                reviewID,
		"instructor_reply_message": reqBody.InstructorReply,
		"instructor_action":        reqBody.InstructorAction,
	})
}
*/
