package models

/*type MarkScale struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}*/

type Exam struct {
	ExamDate  string
	ClassID   string
	UniID     string
	TeacherID string
	MarkScale MarkScale  `gorm:"type:jsonb"`
	Weights   FloatSlice `gorm:"type:jsonb"`
}
