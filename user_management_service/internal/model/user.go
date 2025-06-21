package model

type User struct {
	ID           string `gorm:"primaryKey"`
	Email        string `gorm:"unique"`
	Username     string `gorm:"unique"`
	PasswordHash string
	Role         string
	StudentID    string // optional school ID for students
}
