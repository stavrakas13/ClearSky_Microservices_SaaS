package services

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

func CalculateDistributions(db *gorm.DB, classID string, examDate string) error {
	type Grade struct {
		ClassID   string
		ExamDate  string
		TotalMark float64
		Q01       float64
		Q02       float64
		Q03       float64
		Q04       float64
		Q05       float64
		Q06       float64
		Q07       float64
		Q08       float64
		Q09       float64
		Q10       float64
	}

	var grades []Grade
	err := db.Table("grades").
		Where("class_id = ? AND exam_date = ?", classID, examDate).
		Find(&grades).Error
	if err != nil {
		return err
	}

	distributions := make(map[string]map[int]int)
	fields := []string{"total_mark", "q01", "q02", "q03", "q04", "q05", "q06", "q07", "q08", "q09", "q10"}

	// αρχικοποιούμε τους πίνακες
	for _, f := range fields {
		distributions[f] = make(map[int]int)
		for i := 0; i <= 10; i++ {
			distributions[f][i] = 0
		}
	}

	// αναλυση των βαθμών
	for _, g := range grades {
		distributions["total_mark"][int(g.TotalMark)]++
		distributions["q01"][int(g.Q01)]++
		distributions["q02"][int(g.Q02)]++
		distributions["q03"][int(g.Q03)]++
		distributions["q04"][int(g.Q04)]++
		distributions["q05"][int(g.Q05)]++
		distributions["q06"][int(g.Q06)]++
		distributions["q07"][int(g.Q07)]++
		distributions["q08"][int(g.Q08)]++
		distributions["q09"][int(g.Q09)]++
		distributions["q10"][int(g.Q10)]++
	}

	// εκτύπωση στο terminal
	for field, dist := range distributions {
		fmt.Printf("\n📊 Κατανομή για: %s\n", field)
		for value, count := range dist {
			fmt.Printf("Βαθμός %d: %d φοιτητές\n", value, count)

			// αποθήκευση στη βάση
			err := db.Exec(`
				INSERT INTO grade_distributions (class_id, exam_date, category, value, count)
				VALUES (?, ?, ?, ?, ?)
			`, classID, examDate, field, value, count).Error

			if err != nil {
				log.Printf("❌ DB insert failed for %s/%d: %v", field, value, err)
			}
		}
	}

	return nil
}
