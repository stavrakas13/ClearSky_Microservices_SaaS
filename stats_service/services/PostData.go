package services

import (
	"fmt"
	"log"
	"net/http"
	"stats_service/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type uploadPayload struct {
	Exam   models.Exam    `json:"exam" binding:"required"`
	Grades []models.Grade `json:"grades" binding:"required,dive"`
}

func PostData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var p uploadPayload
		if err := c.ShouldBindJSON(&p); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Upsert στο exams
		if err := db.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "class_id"}, {Name: "exam_date"}},
				DoUpdates: clause.AssignmentColumns([]string{"uni_id", "teacher_id", "mark_scale", "weights"}),
			}).
			Create(&p.Exam).Error; err != nil {
			/*c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return*/
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
				/*c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return*/
				return fmt.Errorf("failed to upsert grade for student %s: %w", g.StudentID, err)
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "✅ Εισαγωγή επιτυχής"})
	}
	// 3. Κάλεσε τον υπολογισμό των κατανομών
	//    (Αυτό θα μπορούσε επίσης να γίνει ασύγχρονα στέλνοντας ένα άλλο μήνυμα
	//    αν ο υπολογισμός είναι χρονοβόρος, αλλά για RPC μπορεί να είναι σύγχρονο)
	err := CalculateDistributions(db, exam.ClassID, exam.ExamDate)
	//err := services.CalculateDistributions(db, classID, examDate)

	if err != nil {
		return fmt.Errorf("failed to calculate distributions: %w", err)
	}
	log.Println("✅ Distributions calculated successfully.")
	return nil
}
