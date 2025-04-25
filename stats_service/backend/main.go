package main

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"backend/services"
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

	err = services.CalculateDistributions(db, "ΤΕΧΝΟΛΟΓΙΑ ΛΟΓΙΣΜΙΚΟΥ   (3205)", "2024-2025 ΧΕΙΜ 2024")
	if err != nil {
		log.Fatal("Σφάλμα στον υπολογισμό κατανομής:", err)
	}

}
