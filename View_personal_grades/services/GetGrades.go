package services

import (
	"View_personal_grades/models"

	"fmt"
	"log"

	"gorm.io/gorm"
)

// HandleGetGrades ανακτά τις βαθμολογίες για μια συγκεκριμένη τάξη και ημερομηνία εξέτασης.
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
