// saas25-12-main/View_personal_grades/services/student_courses_service.go
// ή backend/services/student_courses_service.go

package services

import (
	"fmt"
	"log"

	"View_personal_grades/models"

	"gorm.io/gorm"
)

// StudentCourseInfo είναι η δομή που θα επιστραφεί, ταιριάζοντας με το JSON του frontend
// (εκτός του reviewRequested που θα προστεθεί από τον orchestrator).

type StudentCourseInfo struct {
	Name       string `json:"name"`       // Θα είναι το ClassID
	ExamPeriod string `json:"examPeriod"` // Θα είναι το ExamDate
	Status     string `json:"status"`
	// ReviewRequested bool   `json:"reviewRequested"` // Αυτό θα το χειριστεί ο orchestrator
}

// GetStudentCoursesWithStatus ανακτά τα μαθήματα για έναν φοιτητή και την κατάσταση βαθμολόγησής τους.
func GetStudentCoursesWithStatus(db *gorm.DB, studentID string) ([]StudentCourseInfo, error) {
	log.Printf("INFO: Fetching courses and their status for studentID: %s", studentID)

	var studentGrades []models.Grade // Για να πάρουμε τα μοναδικά μαθήματα/εξεταστικές του φοιτητή

	// Βήμα 1: Βρες όλες τις μοναδικές εγγραφές (class_id, exam_date, is_finalized)
	// από τον πίνακα 'grades' για τον συγκεκριμένο φοιτητή.
	// Χρησιμοποιούμε το Group για να πάρουμε μοναδικούς συνδυασμούς.
	// Το GORM θα επιλέξει αυτόματα τα πεδία του models.Grade που χρειαζόμαστε.
	err := db.Model(&models.Grade{}).
		Select("class_id, exam_date"). // Επιλέγουμε μόνο τα απαραίτητα πεδία
		Where("student_id = ?", studentID).
		Group("class_id, exam_date, is_finalized"). // Ομαδοποίηση για μοναδικότητα
		Order("exam_date DESC, class_id ASC").      // Προαιρετική ταξινόμηση
		Find(&studentGrades).Error

	if err != nil {
		log.Printf("Error fetching distinct courses for studentID %s: %v", studentID, err)
		return nil, fmt.Errorf("failed to fetch courses for student %s: %w", studentID, err)
	}

	if len(studentGrades) == 0 {
		log.Printf("INFO: No courses/grades found for studentID: %s", studentID)
		return []StudentCourseInfo{}, nil // Επιστροφή κενού slice αν δεν βρεθούν μαθήματα
	}

	// Βήμα 2: Μετατροπή των αποτελεσμάτων στο επιθυμητό format StudentCourseInfo
	var coursesInfo []StudentCourseInfo
	for _, gradeEntry := range studentGrades {
		status := "open"
		if gradeEntry.IsFinalized {
			status = "closed"
		}

		coursesInfo = append(coursesInfo, StudentCourseInfo{
			Name:       gradeEntry.ClassID,  // Χρησιμοποιούμε το ClassID ως Name
			ExamPeriod: gradeEntry.ExamDate, // Χρησιμοποιούμε το ExamDate ως ExamPeriod
			Status:     status,
			// ReviewRequested παραμένει εκτός, θα το βάλει ο orchestrator
		})
	}

	log.Printf("INFO: Found %d course entries for studentID: %s", len(coursesInfo), studentID)
	return coursesInfo, nil
}
