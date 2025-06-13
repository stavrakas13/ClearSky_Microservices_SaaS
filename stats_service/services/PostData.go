package services

// Helper used by the RabbitMQ consumer to persist incoming exam data.

import (
	"fmt"
	"log"
	"stats_service/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UploadPayload mirrors the JSON structure sent over RabbitMQ for persisting
// exam information and related grades.
type UploadPayload struct {
	Exam   models.Exam    `json:"exam"`
	Grades []models.Grade `json:"grades"`
}

// PostData stores exam metadata and grade entries inside the database. It is
// invoked by the RabbitMQ consumer rather than via a REST endpoint.
func PostData(db *gorm.DB, p UploadPayload) error {

	// Upsert στο exams
	if err := db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "class_id"}, {Name: "exam_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"uni_id", "teacher_id", "mark_scale", "weights"}),
		}).
		Create(&p.Exam).Error; err != nil {
		return fmt.Errorf("failed to upsert exam: %w", err)
	}

	// Εισαγωγή / upsert για κάθε Grade
	for _, g := range p.Grades {
		// Υπολογισμός total_score
		g.TotalScore = CalculateTotalGrade(g.QuestionScores, p.Exam.Weights, models.MarkScale(p.Exam.MarkScale))
		if err := db.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "class_id"}, {Name: "exam_date"}, {Name: "student_id"}},
				DoUpdates: clause.AssignmentColumns([]string{"question_scores", "total_score"}),
			}).
			Create(&g).Error; err != nil {
			return fmt.Errorf("failed to upsert grade for student %s: %w", g.StudentID, err)
		}
	}

	log.Println("INFO: exam and grades successfully stored")
	return nil

}
