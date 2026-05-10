package session_dto

import (
	"time"

	session_entities "github.com/ricardoalcantara/min-idp/internal/session/entities"
)

type SessionDto struct {
	ID         string     `json:"id"`
	IP         string     `json:"ip"`
	UserAgent  string     `json:"user_agent"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  time.Time  `json:"expires_at"`
	LastSeenAt time.Time  `json:"last_seen_at"`
}

func NewSessionDto(s *session_entities.Session) SessionDto {
	return SessionDto{
		ID:         s.UUID.String(),
		IP:         s.IP,
		UserAgent:  s.UserAgent,
		CreatedAt:  s.CreatedAt,
		ExpiresAt:  s.ExpiresAt,
		LastSeenAt: s.LastSeenAt,
	}
}
