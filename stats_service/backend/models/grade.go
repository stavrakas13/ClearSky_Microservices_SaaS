package models

type Grade struct {
	ExamDate       string
	ClassID        string
	StudentID      string
	QuestionScores FloatSlice
	TotalScore     float64
}
