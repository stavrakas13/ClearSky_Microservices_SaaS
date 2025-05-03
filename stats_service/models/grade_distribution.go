package models

type GradeDistribution struct {
	ClassID  string `json:"class_id" gorm:"primaryKey"`
	ExamDate string `json:"exam_date" gorm:"primaryKey"`
	Category string `json:"category" gorm:"primaryKey"`
	Value    int    `json:"value" gorm:"primaryKey"`
	Count    int    `json:"count"`
}
