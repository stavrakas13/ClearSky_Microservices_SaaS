package models

type Grade struct {
	ExamDate       string     `gorm:"primaryKey;column:exam_date" json:"exam_date"`
	ClassID        string     `gorm:"primaryKey;column:class_id" json:"class_id"`
	StudentID      string     `gorm:"primaryKey;column:student_id" json:"student_id"`
	QuestionScores FloatSlice `gorm:"type:jsonb" json:"question_scores"`
	TotalScore     float64    `gorm:"column:total_score" json:"total_score"`
	IsFinalized    bool       `gorm:"column:is_finalized;default:true"`
}
