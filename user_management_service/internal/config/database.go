package config

import (
    "log"

    "golang.org/x/crypto/bcrypt"
    "github.com/google/uuid"
    "user_management_service/internal/model"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func SetupDatabase() *gorm.DB {
    db, err := gorm.Open(sqlite.Open("auth_service.db"), &gorm.Config{})
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }

    // Auto-migrate the User model
    if err := db.AutoMigrate(&model.User{}); err != nil {
        log.Fatal("Failed to migrate database:", err)
    }

    // Seed an 'admin' user if it doesn't already exist
    password := []byte("admin")
    hash, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
    if err != nil {
        log.Fatalf("Failed to hash admin password: %v", err)
    }

    admin := model.User{
        ID:           uuid.NewString(),
        Username:     "admin",
        PasswordHash: string(hash),
        Role:         "representative",
    }
    // Only create if Username = "admin" is not found
    if err := db.FirstOrCreate(&admin, model.User{Username: "admin"}).Error; err != nil {
        log.Fatalf("Failed to seed admin user: %v", err)
    }

    log.Println("Database connection, migration, and admin seeding completed.")
    return db
}
