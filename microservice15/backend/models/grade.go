package models

type Grade struct {
	ID        uint    `gorm:"primaryKey"`
	StudentID string  `gorm:"column:student_id"`
	CourseID  string  `gorm:"column:course_id"`
	Grade     float64 `gorm:"column:grade"`
}
