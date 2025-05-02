package models

type MarkScale struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type Exam struct {
	ExamDate  string `gorm:"primaryKey;column:exam_date"`
	ClassID   string `gorm:"primaryKey;column:class_id"`
	UniID     string
	TeacherID string
	MarkScale MarkScale `gorm:"type:jsonb"`
	Weights   []float64 `gorm:"type:jsonb"`
}
