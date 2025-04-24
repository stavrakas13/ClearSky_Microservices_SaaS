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
	Student_id               int        `json:"student_id"`
	Course_id                int        `json:"course_id"`
	Student_message          string     `json:"student_message"`
	Status                   string     `json:"status"`
	Instructor_reply_message *string    `json:"instructor_reply_message"`
	Instructor_action        *string    `jason:"Instructor_action"`
	Review_created_at        time.Time  `json:"review_created_at"`
	Reviewed_at              *time.Time `json:"reviewed_at"`
}

type InstructorReply struct {
	InstructorReply  string `json:"instructor_reply_message" binding:"required"`
	InstructorAction string `json:"instructor_action" binding:"required"`
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
