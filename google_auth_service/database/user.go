package database

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex"`
	Name     string
	Picture  string
	Provider string
}
