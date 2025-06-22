package user_management_service

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
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
	loadEnvVariables()
	var err error
	dsn := os.Getenv("DB_DSN")
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Migrate the schema
	DB.AutoMigrate(&User{})
	seedDefaultUsers()
}

// loadEnvVariables loads environment variables from a .env file
func loadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}
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
	var existingUser User
	result := DB.Where("username = ?", "student").First(&existingUser)
	if result.Error != nil {
		// User doesn't exist, create it
		passHash, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
		u := User{
			Username:  "student",
			Password:  string(passHash),
			Role:      "student",
			StudentID: "03181121",
		}
		DB.Create(&u)
		log.Println("Created default student user")
	} else {
		log.Println("Student user already exists")
	}
}

// seedInstructorUser ensures a default instructor exists
func seedInstructorUser() {
	var existingUser User
	result := DB.Where("username = ?", "instructor").First(&existingUser)
	if result.Error != nil {
		// User doesn't exist, create it
		passHash, _ := bcrypt.GenerateFromPassword([]byte("1234"), bcrypt.DefaultCost)
		u := User{
			Username: "instructor",
			Password: string(passHash),
			Role:     "instructor",
		}
		DB.Create(&u)
		log.Println("Created default instructor user")
	} else {
		log.Println("Instructor user already exists")
	}
}
