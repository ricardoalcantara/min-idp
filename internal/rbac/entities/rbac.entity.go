package rbac_entities

import "github.com/ricardoalcantara/min-idp/internal/db"

type Role struct {
	db.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	System      bool `gorm:"default:false"`
}

type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}
