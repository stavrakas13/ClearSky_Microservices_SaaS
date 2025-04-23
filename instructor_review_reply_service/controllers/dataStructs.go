package controllers

import "time"

type ReviewSummary struct {
	StudentID       int       `json:"student_id"`
	CourseID        int       `json:"course_id"`
	ReviewCreatedAt time.Time `json:"review_created_at"`
}

/*
	 type loggedInUser struct {
		userID    int
		firstName string
		lastName  string
		userRole  string
	}
*/
type ReviewStruct struct {
	Review_id                int        `json:"review_id"`
	Student_id               int        `json:"student_id"`
	Course_id                int        `json:"course_id"`
	Student_message          string     `json:"student_message"`
	Status                   string     `json:"status"`
	Instructor_reply_message *string    `json:"instructor_reply_message"`
	Review_created_at        time.Time  `json:"review_created_at"`
	Reviewed_at              *time.Time `json:"reviewed_at"`
}

// DUMMY DATA TO TEST THE SERVICE
// LOGGED IN USER & COURSE INFO SENT BY ORCHESTRATOR ???
// OR GET REQEUST ???

/* var loggedInInstructor = loggedInUser{
	userID:    55555,
	firstName: "Jonh",
	lastName:  "Watchon",
	userRole:  "Instructor",
} */
