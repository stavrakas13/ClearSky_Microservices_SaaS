package services

// Logic for storing exam information and grades when messages arrive from
// RabbitMQ.

import (
	"errors"

	"View_personal_grades/models"

	"gorm.io/gorm"
)

// UploadPayload mirrors the JSON body sent over RabbitMQ.
type UploadPayload struct {
	Exam   models.Exam    `json:"exam"`
	Grades []models.Grade `json:"grades"`
}

// PostData inserts exam metadata and grades into the DB. It is called by the
// RabbitMQ consumer rather than exposed as an HTTP endpoint.
func PostData(db *gorm.DB, p UploadPayload) error {
	var existingExam models.Exam
	err := db.Where("class_id = ? AND exam_date = ?", p.Exam.ClassID, p.Exam.ExamDate).First(&existingExam).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if err := db.Create(&p.Exam).Error; err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if err := db.Model(&existingExam).Updates(map[string]interface{}{
			"uni_id":     p.Exam.UniID,
			"teacher_id": p.Exam.TeacherID,
			"mark_scale": p.Exam.MarkScale,
			"weights":    p.Exam.Weights,
		}).Error; err != nil {
			return err
		}
	}

	for _, g := range p.Grades {
		g.TotalScore = CalculateTotalGrade(g.QuestionScores, p.Exam.Weights, models.MarkScale(p.Exam.MarkScale))

		var existingGrade models.Grade
		err := db.Where("class_id = ? AND exam_date = ? AND student_id = ?", g.ClassID, g.ExamDate, g.StudentID).First(&existingGrade).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := db.Create(&g).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if err := db.Model(&existingGrade).Updates(map[string]interface{}{
				"question_scores": g.QuestionScores,
				"total_score":     g.TotalScore,
			}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
