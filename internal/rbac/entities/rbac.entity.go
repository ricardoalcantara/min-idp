package rbac_entities

import (
	"github.com/ricardoalcantara/min-idp/internal/db"
	"gorm.io/gorm"
)

type Role struct {
	db.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	System      bool `gorm:"default:false"`
}

type Permission struct {
	db.Model
	Name string `gorm:"uniqueIndex;not null"`
}

type Group struct {
	db.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
}

// Join tables — no UUID needed
type RolePermission struct {
	RoleID       uint `gorm:"primaryKey"`
	PermissionID uint `gorm:"primaryKey"`
}

type UserRole struct {
	UserID uint `gorm:"primaryKey"`
	RoleID uint `gorm:"primaryKey"`
}

type UserGroup struct {
	UserID  uint `gorm:"primaryKey"`
	GroupID uint `gorm:"primaryKey"`
}

// RBACRepository uses gorm.DB directly for join queries; keep a thin query model
type PermissionQuery struct {
	gorm.Model
	Name string
}

func (PermissionQuery) TableName() string { return "permissions" }
