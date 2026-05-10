package user_entities

import "github.com/ricardoalcantara/min-idp/internal/db"

type User struct {
	db.Model
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Status       string `gorm:"default:'active'"`
}
