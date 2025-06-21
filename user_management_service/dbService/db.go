package user_management_service

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

// User represents the user model
type User struct {
	gorm.Model
	Username string
	Password string
	Role     string
}

// InitDB initializes the database connection and performs migrations
func InitDB() {
	// ...existing DB connection and migrations...
	DB.AutoMigrate(&User{})
	seedAdminUser()
}

// seedAdminUser ensures a default admin exists
func seedAdminUser() {
	passHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	u := User{Username: "admin", Password: string(passHash), Role: "representative"}
	DB.FirstOrCreate(&u, User{Username: "admin"})
}
