package services

import (
	"fmt"
	"log"
	"stats_service/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// HandlePersistAndCalculate αποθηκεύει τα δεδομένα και υπολογίζει τις κατανομές.
// Καλείται από τον RabbitMQ consumer.
func HandlePersistAndCalculate(db *gorm.DB, exam models.Exam, grades []models.Grade) error {
	log.Printf("INFO: Persisting data and calculating distributions for ClassID: %s, ExamDate: %s", exam.ClassID, exam.ExamDate)

	// 1. Upsert Exam
	if err := db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "class_id"}, {Name: "exam_date"}},
			DoUpdates: clause.AssignmentColumns([]string{"uni_id", "teacher_id", "mark_scale", "weights"}),
		}).
		Create(&exam).Error; err != nil {
		return fmt.Errorf("failed to upsert exam: %w", err)
	}
	log.Printf("INFO: Exam upserted for ClassID: %s, ExamDate: %s", exam.ClassID, exam.ExamDate)

	// 2. Upsert Grades
	if len(grades) > 0 {
		for _, g := range grades {
			// Αν τα grades που έρχονται στο payload δεν έχουν ClassID και ExamDate,
			// ή αν θέλεις να είσαι σίγουρος ότι παίρνουν αυτά της εξέτασης:
			g.ClassID = exam.ClassID
			g.ExamDate = exam.ExamDate

			g.TotalScore = CalculateTotalGrade(g.QuestionScores, exam.Weights, models.MarkScale(exam.MarkScale))
			if err := db.
				Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "class_id"}, {Name: "exam_date"}, {Name: "student_id"}},
					DoUpdates: clause.AssignmentColumns([]string{"question_scores", "total_score"}),
				}).
				Create(&g).Error; err != nil {
				return fmt.Errorf("failed to upsert grade for student %s (ClassID: %s, ExamDate: %s): %w", g.StudentID, g.ClassID, g.ExamDate, err)
			}
		}
		log.Printf("INFO: %d grades upserted for ClassID: %s, ExamDate: %s", len(grades), exam.ClassID, exam.ExamDate)
	} else {
		log.Printf("INFO: No grades provided in payload for ClassID: %s, ExamDate: %s", exam.ClassID, exam.ExamDate)
	}

	// 3. Κάλεσε τον υπολογισμό των κατανομών
	// Έλεγχος αν υπάρχουν όντως δεδομένα στη βάση πριν τον υπολογισμό (προαιρετικό, αλλά καλό)
	var count int64
	db.Model(&models.Grade{}).Where("class_id = ? AND exam_date = ?", exam.ClassID, exam.ExamDate).Count(&count)
	if count == 0 {
		log.Printf("WARNING: No grades found in DB after upsert for ClassID %s, ExamDate %s. Skipping distribution calculation.", exam.ClassID, exam.ExamDate)
		return nil // Ή μπορείς να επιστρέψεις ένα συγκεκριμένο σφάλμα/μήνυμα
	}

	// Κλήση της CalculateDistributions
	// Η CalculateDistributions θα διαβάσει τα grades από τη βάση.
	err := CalculateDistributions(db, exam.ClassID, exam.ExamDate)
	if err != nil {
		return fmt.Errorf("failed to calculate distributions: %w", err)
	}
	log.Println("INFO: Distributions calculation triggered successfully.")
	return nil
}
