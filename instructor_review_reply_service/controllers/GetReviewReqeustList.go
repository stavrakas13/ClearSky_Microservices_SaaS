package controllers

func GetReviewReqeustList(params map[string]string, body map[string]interface{}) (string, error) {
	return "response", nil

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
