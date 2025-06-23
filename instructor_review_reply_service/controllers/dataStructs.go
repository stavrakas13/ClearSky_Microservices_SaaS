package controllers

import "time"

type ReviewSummary struct {
	StudentID       string    `json:"student_id"`
	CourseID        string    `json:"course_id"`
	ReviewCreatedAt time.Time `json:"review_created_at"`
}

type ReviewStruct struct {
	Student_id               string     `json:"student_id"`
	Course_id                string     `json:"course_id"`
	Exam_period              string     `json:"exam_period"`
	Student_message          string     `json:"student_message"`
	Status                   string     `json:"status"`
	Instructor_reply_message *string    `json:"instructor_reply_message"`
	Instructor_action        *string    `json:"instructor_action"`
	Review_created_at        time.Time  `json:"review_created_at"`
	Reviewed_at              *time.Time `json:"reviewed_at"`
}

type InstructorReply struct {
	InstructorReply  string `json:"instructor_reply_message" binding:"required"`
	InstructorAction string `json:"instructor_action" binding:"required"`
}
