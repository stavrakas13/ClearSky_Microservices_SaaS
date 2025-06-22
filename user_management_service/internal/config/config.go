package config

import (
	"os"
	"user_management_service/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// SetupDatabase opens the DB, runs migrations, and seeds a default admin user
func SetupDatabase() *gorm.DB {
	// Use auth_service.db as default
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = "auth_service.db"
	}
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	// Migrate the User model
	db.AutoMigrate(&model.User{})

	// Seed default admin user (username: admin / password: admin)
	passHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	admin := model.User{Username: "admin", PasswordHash: string(passHash), Role: "institution_representative"}
	db.FirstOrCreate(&admin, model.User{Username: "admin"})

	return db
}
