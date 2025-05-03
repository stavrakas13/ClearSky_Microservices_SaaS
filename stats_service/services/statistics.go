package services

import (
	//"encoding/json"
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

	return nil
}
