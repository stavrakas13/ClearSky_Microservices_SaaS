// student_request_review_service
package main

import (
	"database/sql"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// DUMMY DATA TO TEST THE SERVICE
// LOGGED IN USER & COURSE INFO SENT BY ORCHESTRATOR ???
// OR GET REQEUST ???

type loggedInUser struct {
	userID    int
	firstName string
	lastName  string
	userRole  string
}

type ReviewRequest struct {
	StudentMessage string `json:"student_message" binding:"required"`
}

var loggedInStudent = loggedInUser{
	userID:    11111,
	firstName: "Sherlock",
	lastName:  "Holmes",
	userRole:  "Student",
}

func main() {

	// CONNECT TO REVIEWS DB
	// URL for docker connection
	reviewsdbURL := "postgres://postgres:root@db:5432/reviewsdb?sslmode=disable"
	// URL for local connection
	// reviewsdbURL := "postgres://postgres:root@localhost:5432/reviews?sslmode=disable"

	db, dbconnectfail := sql.Open("postgres", reviewsdbURL)

	if dbconnectfail != nil {
		log.Fatal("Connection to db failed on student end.", dbconnectfail)
	} else {
		log.Println("Reviwsdb connected successfully on student end.")
	}

	defer db.Close()

	// Start GIN server.
	server := gin.Default()

	// ENDPOINT
	server.GET("/review_request", func(ctx *gin.Context) {
		ctx.String(200, "Welcome to reviews request page, student courses list.")
	})

	// POST A NEW REVIEW REQUEST
	server.POST("/new_review_request/:course_id", func(ctx *gin.Context) {

		// COURSE ID FROM URL ???
		courseIDStr := ctx.Param("course_id")
		courseID, _ := strconv.Atoi(courseIDStr)

		// MESSAGE TO INSTRUCTOR FROM REQUEST BODY
		var reqBody ReviewRequest

		if err := ctx.ShouldBindJSON(&reqBody); err != nil {
			ctx.JSON(400, gin.H{"error": "Invalid JSON body"})
			return
		}
		if reqBody.StudentMessage == "" {
			ctx.JSON(400, gin.H{"error": "Student message is required"})
			return
		}

		log.Printf("Inserting review: student_id=%d, course_id=%d, message=%s", loggedInStudent.userID, courseID, reqBody.StudentMessage)

		query := `INSERT INTO reviews (student_id, course_id, student_message) VALUES ($1, $2, $3)`
		_, inserterr := db.Exec(query, loggedInStudent.userID, courseID, reqBody.StudentMessage)

		if inserterr != nil {
			log.Println("Insert error:", inserterr)

		}

		ctx.JSON(200, gin.H{
			"message":         "Review request submitted successfully.",
			"user_id":         loggedInStudent.userID,
			"course_id":       courseID,
			"student_message": reqBody.StudentMessage,
		})
	})

	server.Run(":8087")
}
