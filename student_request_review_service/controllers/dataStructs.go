package controllers

import "time"

type ReviewRequest struct {
	StudentMessage string `json:"student_message" binding:"required"`
}

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
