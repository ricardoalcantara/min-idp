package user_entities

import (
	"github.com/ricardoalcantara/min-idp/internal/db"
	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
)

type User struct {
	db.Model
	Email        string               `gorm:"uniqueIndex;not null"`
	Username     string               `gorm:"uniqueIndex"`
	Name         string
	PasswordHash string               `gorm:"not null"`
	Status       string               `gorm:"default:'active'"`
	Roles        []rbac_entities.Role `gorm:"many2many:user_roles;"`
}
