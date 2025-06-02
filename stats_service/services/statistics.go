package services

import (
	"errors"
	"fmt"
	"log"
	"math"
	"stats_service/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Το Grade struct που χρησιμοποιείται εσωτερικά από το CalculateDistributions
// Πρέπει να ταιριάζει με τον πίνακα 'grades'
type GradeForStats struct {
	ExamDate       string            `gorm:"column:exam_date"`
	ClassID        string            `gorm:"column:class_id"`
	StudentID      string            `gorm:"column:student_id"`
	QuestionScores models.FloatSlice `gorm:"type:jsonb"`
	TotalScore     float64           `gorm:"column:total_score"`
}

// TableName ορίζει το όνομα του πίνακα για το GORM για το GradeForStats
func (GradeForStats) TableName() string {
	return "grades"
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
	var grades []GradeForStats // Χρησιμοποίησε το τοπικό struct
	var exam models.Exam

	log.Printf("INFO: Calculating distributions for ClassID: %s, ExamDate: %s", classID, examDate)

	if err := db.Where("class_id = ? AND exam_date = ?", classID, examDate).Find(&grades).Error; err != nil {
		return fmt.Errorf("error fetching grades for distributions: %w", err)
	}

	if len(grades) == 0 {
		log.Printf("WARNING: No grades found in DB for ClassID %s, ExamDate %s to calculate distributions.", classID, examDate)
		return nil
	}

	if err := db.Where("class_id = ? AND exam_date = ?", classID, examDate).First(&exam).Error; err != nil {
		return fmt.Errorf("error fetching exam details for distributions: %w", err)
	}

	numQuestions := 0
	if exam.Weights != nil {
		numQuestions = len(exam.Weights)
	}

	if numQuestions == 0 && len(grades) > 0 && grades[0].QuestionScores != nil {
		if len(grades[0].QuestionScores) > 0 {
			numQuestions = len(grades[0].QuestionScores)
			log.Printf("WARNING: No weights found for exam %s / %s. Using number of question scores from first grade: %d", classID, examDate, numQuestions)
		} else {
			log.Printf("WARNING: No weights or question scores found for exam %s / %s. Only total score distribution will be calculated.", classID, examDate)
		}
	}
	if numQuestions == 0 && (len(grades) == 0 || grades[0].QuestionScores == nil) {
		log.Printf("CRITICAL: No weights and no question scores available for exam %s / %s. Cannot calculate question distributions. Will only process total_score.", classID, examDate)
		// Δεν κάνουμε return error, θα προχωρήσει μόνο με το total_score
	}

	distributions := make(map[string]map[int]int)
	distributions["total_score"] = make(map[int]int)
	for v := int(exam.MarkScale.Min); v <= int(exam.MarkScale.Max); v++ {
		distributions["total_score"][v] = 0
	}

	if numQuestions > 0 {
		for i := 0; i < numQuestions; i++ {
			qName := fmt.Sprintf("q%02d", i+1)
			distributions[qName] = make(map[int]int)
			for v := 0; v <= 10; v++ { // Υποθέτουμε κλίμακα 0-10 για τις ερωτήσεις
				distributions[qName][v] = 0
			}
		}
	}

	for _, g := range grades {
		roundedTotal := int(math.Round(g.TotalScore))
		// Έλεγχος αν το roundedTotal είναι εντός του αναμενόμενου εύρους του distributions["total_score"]
		if _, ok := distributions["total_score"][roundedTotal]; ok {
			distributions["total_score"][roundedTotal]++
		} else {
			// Αν είναι εκτός, μπορείς να το αγνοήσεις ή να το καταγράψεις
			log.Printf("WARNING: Rounded total score %d for student %s is out of expected scale [%.0f-%.0f].",
				roundedTotal, g.StudentID, exam.MarkScale.Min, exam.MarkScale.Max)
			// Για να μην κρασάρει, μπορείς να το βάλεις στο πλησιέστερο όριο ή να το αγνοήσεις
			if roundedTotal < int(exam.MarkScale.Min) {
				roundedTotal = int(exam.MarkScale.Min)
			}
			if roundedTotal > int(exam.MarkScale.Max) {
				roundedTotal = int(exam.MarkScale.Max)
			}
			if _, okInit := distributions["total_score"][roundedTotal]; okInit { // Έλεγχος ξανά μετά τη διόρθωση
				distributions["total_score"][roundedTotal]++
			}
		}

		if g.QuestionScores != nil && numQuestions > 0 {
			for i, score := range g.QuestionScores {
				if i >= numQuestions {
					log.Printf("WARNING: Grade for student %s has more question scores (%d) than exam weights/questions defined (%d). Skipping extra scores.", g.StudentID, len(g.QuestionScores), numQuestions)
					break
				}
				qName := fmt.Sprintf("q%02d", i+1)
				roundedScore := int(math.Round(score))
				if _, ok := distributions[qName][roundedScore]; ok {
					distributions[qName][roundedScore]++
				} else {
					log.Printf("WARNING: Rounded score %d for %s (student %s) is out of expected scale (0-10).", roundedScore, qName, g.StudentID)
				}
			}
		}
	}

	for category, distMap := range distributions {
		log.Printf("📊 Saving distribution for: %s (ClassID: %s, ExamDate: %s)", category, classID, examDate)
		for value, count := range distMap {
			if count == 0 {
				continue
			}
			gradeDist := models.GradeDistribution{
				ClassID:  classID,
				ExamDate: examDate,
				Category: category,
				Value:    value,
				Count:    count,
			}
			err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "class_id"}, {Name: "exam_date"}, {Name: "category"}, {Name: "value"}},
				DoUpdates: clause.AssignmentColumns([]string{"count"}),
			}).Create(&gradeDist).Error

			if err != nil {
				log.Printf("❌ DB insert/update failed for distribution %s/%d: %v", category, value, err)
			}
		}
	}
	log.Printf("INFO: Distributions saved for ClassID: %s, ExamDate: %s", classID, examDate)
	return nil
}

func GetDistributions(db *gorm.DB, classID string, examDate string) ([]models.GradeDistribution, error) {
	var distributions []models.GradeDistribution
	log.Printf("INFO: Retrieving distributions for ClassID: %s, ExamDate: %s", classID, examDate)

	if err := db.Where("class_id = ? AND exam_date = ?", classID, examDate).Find(&distributions).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve distributions from DB for ClassID %s, ExamDate %s: %w", classID, examDate, err)
	}

	if len(distributions) == 0 {
		log.Printf("INFO: No pre-calculated distributions found for ClassID %s, ExamDate %s.", classID, examDate)
	}

	log.Printf("INFO: Retrieved %d distribution records.", len(distributions))
	return distributions, nil
}
