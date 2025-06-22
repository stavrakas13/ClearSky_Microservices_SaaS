package model

import "time"

type User struct {
	ID           string `gorm:"primaryKey"`
	Username     string `gorm:"uniqueIndex"`
	PasswordHash string
	Role         string
	StudentID    string // optional school ID for students
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
