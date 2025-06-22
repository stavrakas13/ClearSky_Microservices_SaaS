package database

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email     string `gorm:"unique;not null"`
	Name      string
	Picture   string
	Provider  string `gorm:"default:'google'"`
	Role      string `gorm:"default:'institution_representative'"`
	StudentID string `gorm:"unique"`
}
