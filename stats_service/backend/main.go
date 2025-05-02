package main

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"backend/services"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "host=db user=myuser password=mypass dbname=mydb port=5432 sslmode=disable"
	}

	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Println("❌ Waiting for DB...")
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		log.Fatalf("❌ Could not connect to the database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	// ✅ Προσοχή: Χωρίς περιττά κενά!
	classID := "ΤΕΧΝΟΛΟΓΙΑ ΛΟΓΙΣΜΙΚΟΥ (3205)"
	examDate := "2024-2025 ΧΕΙΜ 2024"

	err = services.CalculateDistributions(db, classID, examDate)
	if err != nil {
		log.Fatalf("Σφάλμα στον υπολογισμό κατανομής: %v", err)
	}
}
