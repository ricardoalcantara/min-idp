package oidc_entities

import (
	"time"

	"github.com/ricardoalcantara/min-idp/internal/db"
)

type OAuthToken struct {
	db.Model
	Type        string     `gorm:"not null;index"` // "access" or "refresh"
	TokenHash   string     `gorm:"not null;uniqueIndex"` // Hash of refresh token or JTI of access token
	ClientID    string     `gorm:"not null;index"`
	UserID      uint       `gorm:"not null;index"`
	SessionUUID string     `gorm:"index"`
	Scope       string     `gorm:"type:text"`
	ExpiresAt   time.Time  `gorm:"index"`
	RevokedAt   *time.Time
	ParentID    *uint      // For refresh token rotation tracking
}

func (OAuthToken) TableName() string { return "oauth_tokens" }
