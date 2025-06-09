// saas25-12-main/View_personal_grades/services/personal_grades_service.go (ή παρόμοιο)
package services

import (
	"fmt"
	"log"

	// Βεβαιώσου ότι το import path για τα models είναι σωστό για το project σου
	"View_personal_grades/models" // Πιθανό path, προσάρμοσέ το!

	"gorm.io/gorm"
)

// PersonalGradesResponse είναι η δομή για το πεδίο "grades" της τελικής απάντησης.
// Χρησιμοποιούμε map[string]float64 για ευελιξία με Q1, Q2, ...
type PersonalGradesResponse map[string]float64

// GetStudentPersonalGrades ανακτά τους αναλυτικούς βαθμούς ενός φοιτητή για μια εξέταση.
// Επιστρέφει ένα map κατάλληλο για το πεδίο "grades" του JSON response.
func GetStudentPersonalGrades(db *gorm.DB, classID string, examDate string, studentID string) (PersonalGradesResponse, error) {
	log.Printf("INFO: Fetching personal grades for StudentID: %s, ClassID: %s, ExamDate: %s", studentID, classID, examDate)

	var gradeRecord models.Grade // Αναμένουμε μία εγγραφή, αφού το studentID είναι μέρος του primary key

	// Ψάχνουμε για τη συγκεκριμένη εγγραφή βαθμολογίας
	// Το First θα επιστρέψει σφάλμα gorm.ErrRecordNotFound αν δεν βρεθεί η εγγραφή
	err := db.Where("class_id = ? AND exam_date = ? AND student_id = ?", classID, examDate, studentID).
		First(&gradeRecord).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("INFO: No grade record found for StudentID: %s, ClassID: %s, ExamDate: %s", studentID, classID, examDate)
			// Επιστρέφουμε nil και ένα custom error ή απλά nil, nil αν το "μη εύρεση" δεν είναι σφάλμα για τον orchestrator
			return nil, fmt.Errorf("no grade record found for student %s in class %s on %s", studentID, classID, examDate)
		}
		log.Printf("ERROR: Failed to fetch grade record for StudentID %s: %v", studentID, err)
		return nil, fmt.Errorf("database error fetching grade record: %w", err)
	}

	// Δημιουργία του map για την απάντηση
	gradesResponse := make(PersonalGradesResponse)
	gradesResponse["total"] = gradeRecord.TotalScore

	// Προσθήκη των QuestionScores
	// Το models.Grade έχει QuestionScores []float64
	// Το frontend περιμένει Q1, Q2, ...
	if gradeRecord.QuestionScores != nil {
		for i, score := range gradeRecord.QuestionScores {
			questionKey := fmt.Sprintf("Q%d", i+1)
			gradesResponse[questionKey] = score
		}
	}

	log.Printf("INFO: Successfully fetched personal grades for StudentID: %s", studentID)
	return gradesResponse, nil
}

// Η παλιά σου συνάρτηση HandleGetGrades, αν τη χρειάζεσαι ακόμα για άλλο σκοπό,
// μπορεί να παραμείνει ως έχει ή να μετονομαστεί.
// Αν η GetStudentPersonalGrades την αντικαθιστά πλήρως, μπορείς να την αφαιρέσεις.
/*
func HandleGetGrades(db *gorm.DB, classID string, examDate string, student_id string) ([]models.Grade, error) {
	log.Printf("INFO: Fetching grades for ClassID: %s, ExamDate: %s", classID, examDate)

	var grades []models.Grade
	err := db.Where("class_id = ? AND exam_date = ? AND student_id=?", classID, examDate, student_id).Find(&grades).Error
	if err != nil {
		return nil, fmt.Errorf("failed to fetch grades for ClassID %s, ExamDate %s: %w", classID, examDate, err)
	}

	if len(grades) == 0 {
		log.Printf("INFO: No grades found for ClassID: %s, ExamDate: %s", classID, examDate)
	} else {
		log.Printf("INFO: Found %d grades for ClassID: %s, ExamDate: %s", len(grades), classID, examDate)
	}

	return grades, nil
}
*/
