package models

type Grade struct {
	ExamDate       string    `gorm:"primaryKey;column:exam_date"`
	ClassID        string    `gorm:"primaryKey;column:class_id"`
	StudentID      string    `gorm:"primaryKey;column:student_id"`
	QuestionScores []float64 `gorm:"type:jsonb"`
	TotalScore     float64
}
