package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"stats_backend/models"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=db user=myuser password=mypass dbname=mydb port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("❌ Could not connect to the database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	var grades []models.Grade
	result := db.Find(&grades)
	if result.Error != nil {
		log.Fatalf("❌ Failed to fetch grades: %v", result.Error)
	}

	for _, g := range grades {
		log.Printf("📊 StudentID: %s | CourseID: %s | Grade: %.2f", g.StudentID, g.CourseID, g.Grade)
	}
}
