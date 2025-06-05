// stats_service/services/statistics_retrieval.go (ή όπως το ονομάσεις)
package services

import (
	"fmt"
	"log"
	"stats_service/models" // Βεβαιώσου ότι το path είναι σωστό

	"gorm.io/gorm"
)

// GetGradeDistributions ανακτά τις προ-υπολογισμένες κατανομές βαθμών
// για μια συγκεκριμένη τάξη και ημερομηνία εξέτασης.
func GetGradeDistributions(db *gorm.DB, classID string, examDate string) ([]models.GradeDistribution, error) {
	var distributions []models.GradeDistribution

	// Κάνε query στον πίνακα grade_distributions.
	// Το GORM θα συμπεράνει το όνομα του πίνακα από το struct models.GradeDistribution
	// αν ακολουθεί τις συμβάσεις (π.χ. grade_distributions) ή αν έχει μια μέθοδο TableName().
	// Για σαφήνεια, μπορείς να χρησιμοποιήσεις db.Table("grade_distributions").Where(...)
	err := db.Where("class_id = ? AND exam_date = ?", classID, examDate).Find(&distributions).Error

	if err != nil {
		log.Printf("Error retrieving grade distributions for ClassID %s, ExamDate %s: %v", classID, examDate, err)
		return nil, fmt.Errorf("failed to retrieve grade distributions: %w", err)
	}

	if len(distributions) == 0 {
		log.Printf("No grade distributions found for ClassID %s, ExamDate %s", classID, examDate)
		// Επιστρέφει κενό slice και όχι σφάλμα, ή ένα συγκεκριμένο "not found" error αν προτιμάς
	}

	return distributions, nil
}
