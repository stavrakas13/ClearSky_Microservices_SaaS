package user_management_service

import (
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var DB *gorm.DB

// User represents the user model
type User struct {
	gorm.Model
	Username  string
	Password  string
	Role      string
	StudentID string `gorm:"unique"` // Add StudentID field
}

// InitDB initializes the database connection and performs migrations
func InitDB() {
	// ...existing DB connection and migrations...
	DB.AutoMigrate(&User{})
	seedDefaultUsers()
}

// seedDefaultUsers ensures default users exist for all roles
func seedDefaultUsers() {
	seedAdminUser()
	seedStudentUser()
	seedInstructorUser()
}

// seedAdminUser ensures a default admin exists
func seedAdminUser() {
	passHash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	u := User{Username: "admin", Password: string(passHash), Role: "institution_representative"}
	DB.FirstOrCreate(&u, User{Username: "admin"})
}

// seedStudentUser ensures a default student exists
func seedStudentUser() {
	passHash, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
	u := User{
		Username:  "student",
		Password:  string(passHash),
		Role:      "student",
		StudentID: "03181121",
	}
	DB.FirstOrCreate(&u, User{Username: "student"})
}

// seedInstructorUser ensures a default instructor exists
func seedInstructorUser() {
	passHash, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
	u := User{Username: "instructor", Password: string(passHash), Role: "instructor"}
	DB.FirstOrCreate(&u, User{Username: "instructor"})
}
