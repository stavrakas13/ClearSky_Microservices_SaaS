package services

import (
	//"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"stats_service/models"

	"gorm.io/gorm"
)

type Grade struct {
	ExamDate       string            `gorm:"column:exam_date"`
	ClassID        string            `gorm:"column:class_id"`
	StudentID      string            `gorm:"column:student_id"`
	QuestionScores models.FloatSlice `gorm:"type:jsonb"`
	TotalScore     float64           `gorm:"column:total_score"`
}

// CheckIfGradesExist ελέγχει αν υπάρχουν εγγραφές στον πίνακα grades
// για το συγκεκριμένο classID και examDate.
func CheckIfGradesExist(db *gorm.DB, classID string, examDate string) bool {
	var grade models.Grade // Χρειαζόμαστε ένα struct για να προσπαθήσει το GORM να γεμίσει

	// Θέλουμε να δούμε αν υπάρχει *τουλάχιστον μία* εγγραφή.
	// Το .First() θα επιστρέψει gorm.ErrRecordNotFound αν δεν βρει τίποτα.
	// Χρησιμοποιούμε .Select("class_id") για να κάνουμε το query πιο ελαφρύ,
	// καθώς δεν χρειαζόμαστε όλα τα δεδομένα της γραμμής, μόνο την ύπαρξή της.
	result := db.Select("class_id").Where("class_id = ? AND exam_date = ?", classID, examDate).First(&grade)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			// Δεν βρέθηκε καμία εγγραφή, άρα δεν υπάρχουν δεδομένα.
			log.Printf("INFO: No grades found for ClassID: %s, ExamDate: %s", classID, examDate)
			return false
		}
		// Κάποιο άλλο σφάλμα συνέβη κατά την εκτέλεση του query.
		log.Printf("ERROR: Could not query grades for ClassID: %s, ExamDate: %s - %v", classID, examDate, result.Error)
		return false
	}

	// Αν δεν υπάρχει σφάλμα, σημαίνει ότι βρέθηκε τουλάχιστον μία εγγραφή.
	log.Printf("INFO: Grades found for ClassID: %s, ExamDate: %s", classID, examDate)
	return true
}

func CalculateDistributions(db *gorm.DB, classID string, examDate string) error {
	var grades []Grade
	var exam models.Exam
	if err := db.Where("class_id = ? AND exam_date = ?", classID, examDate).Find(&grades).Error; err != nil {
		return err
	}

	if err := db.Where("class_id = ? AND exam_date = ?", classID, examDate).First(&exam).Error; err != nil {
		return err
	}

	numQuestions := len(exam.Weights)
	if numQuestions == 0 {
		return fmt.Errorf("no weights found for exam %s / %s", classID, examDate)
	}

	if CheckIfGradesExist(db, classID, examDate) == false {
		// Αρχικοποίηση κατανομής για κάθε ερώτηση
		distributions := make(map[string]map[int]int)
		for i := 0; i < numQuestions; i++ {
			qName := fmt.Sprintf("q%02d", i+1)
			distributions[qName] = make(map[int]int)
			for v := int(exam.MarkScale.Min); v <= int(exam.MarkScale.Max); v++ {
				distributions[qName][v] = 0
			}
		}

		// Επεξεργασία κάθε βαθμού
		for _, g := range grades {
			for i, score := range g.QuestionScores {
				if i >= numQuestions {
					continue
				}
				qName := fmt.Sprintf("q%02d", i+1)
				rounded := int(math.Round(score))
				if _, ok := distributions[qName][rounded]; ok {
					distributions[qName][rounded]++
				}
			}
		}

		// Εκτύπωση και αποθήκευση στη βάση
		for q, dist := range distributions {
			fmt.Printf("\n📊 Κατανομή για: %s\n", q)
			for value, count := range dist {
				fmt.Printf("Βαθμός %d: %d φοιτητές\n", value, count)

				err := db.Exec(`
					INSERT INTO grade_distributions (class_id, exam_date, category, value, count)
					VALUES (?, ?, ?, ?, ?)
					ON CONFLICT (class_id, exam_date, category, value)
					DO UPDATE SET count = EXCLUDED.count
				`, classID, examDate, q, value, count).Error

				if err != nil {
					log.Printf("❌ DB insert failed for %s/%d: %v", q, value, err)
				}
			}
		}
	}
	log.Printf("❌ No need for DB insert already statistics exist for %s/%s", classID, examDate)
	return nil
}
