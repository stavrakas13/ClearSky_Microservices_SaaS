package models

type MarkScale struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type Exam struct {
	ClassID   string        `gorm:"primaryKey;column:class_id" json:"class_id"`
	ExamDate  string        `gorm:"primaryKey;column:exam_date" json:"exam_date"`
	UniID     string        `gorm:"column:uni_id" json:"uni_id"`
	TeacherID string        `gorm:"column:teacher_id" json:"teacher_id"`
	MarkScale JSONMarkScale `gorm:"type:jsonb" json:"mark_scale"`
	Weights   FloatSlice    `gorm:"type:jsonb" json:"weights"`
}
