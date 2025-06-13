package services

// Business logic for persisting exam data and computing grade distributions.

import (
	"fmt"
	"log"
	"stats_service/models"

	"gorm.io/gorm"
)

// HandlePersistAndCalculate αποθηκεύει τα δεδομένα και υπολογίζει τις κατανομές.
// Καλείται από τον RabbitMQ consumer.
func HandlePersistAndCalculate(db *gorm.DB, exam models.Exam, grades []models.Grade) error {
	log.Printf("INFO: Persisting data and calculating distributions for ClassID: %s, ExamDate: %s", exam.ClassID, exam.ExamDate)

	if err := PostData(db, UploadPayload{Exam: exam, Grades: grades}); err != nil {
		return err
	}

	// 1. Κάλεσε τον υπολογισμό των κατανομών
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
