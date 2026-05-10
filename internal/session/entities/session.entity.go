package session_entities

import (
	"time"

	"github.com/ricardoalcantara/min-idp/internal/db"
	"gorm.io/gorm"
)

type Session struct {
	db.Model
	UserID     uint       `gorm:"not null;index"`
	ExpiresAt  time.Time  `gorm:"index"`
	LastSeenAt time.Time
	IP         string
	UserAgent  string
	RevokedAt  *time.Time
}

func (s *Session) IsValid() bool {
	return s.RevokedAt == nil && time.Now().Before(s.ExpiresAt)
}

type SPSession struct {
	gorm.Model
	SessionID uint   `gorm:"not null;uniqueIndex:idx_sp_session"`
	SPID      uint   `gorm:"not null;uniqueIndex:idx_sp_session"`
	Sub       string `gorm:"not null"`
}
