// db/init.go
package db

import (
	"fmt"
	"log"
	"os"

	"stats_service/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func init() {
	// load .env in local dev; in Docker this just no-ops if .env isn’t present
	_ = godotenv.Load()
}

// InitDB returns a gorm DB, reading all settings from environment.
func InitDB() (*gorm.DB, error) {
	// 1) If user set a full connection string, use it:
	dsn := os.Getenv("DB_DSN")

	// 2) Otherwise build it piecewise:
	if dsn == "" {
		// SUPPORT BOTH LOCAL & DOCKER HOST NAMES
		host := os.Getenv("DB_HOST")
		if host == "" {
			host = "localhost"
		}
		dsn = fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			host,
			os.Getenv("POSTGRES_USER"),
			os.Getenv("POSTGRES_PASSWORD"),
			os.Getenv("POSTGRES_DB"),
			os.Getenv("DB_PORT"),
		)
	}

	// OPEN THE CONNECTION
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("could not connect to postgres: %w", err)
	}

	// AUTO‐MIGRATE YOUR SCHEMA
	if err := db.AutoMigrate(
		&models.Exam{},
		&models.Grade{},
		//&models.GradeDistribution{},
	); err != nil {
		log.Printf("⚠️  Migration warning: %v", err)
	}

	return db, nil
}
