package session

import (
	"context"
	"fmt"
	"time"

	"github.com/ricardoalcantara/min-idp/internal/config"
	session_entities "github.com/ricardoalcantara/min-idp/internal/session/entities"
)

type SessionRepository interface {
	Create(s *session_entities.Session) error
	FindByUUID(uuid string) (*session_entities.Session, error)
	FindActiveByUserID(userID uint) ([]session_entities.Session, error)
	Touch(ctx context.Context, id uint) error
	Revoke(ctx context.Context, id uint) error
	RevokeAll(ctx context.Context, userID uint) error
	RevokeAllExcept(ctx context.Context, userID, exceptID uint) error
}

type SessionService struct {
	repo SessionRepository
	cfg  *config.Config
}

func NewSessionService(repo SessionRepository, cfg *config.Config) *SessionService {
	return &SessionService{repo: repo, cfg: cfg}
}

func (s *SessionService) Create(ctx context.Context, userID uint, ip, ua string) (*session_entities.Session, error) {
	now := time.Now().UTC()
	sess := &session_entities.Session{
		UserID:     userID,
		ExpiresAt:  now.Add(s.cfg.SessionTTL),
		LastSeenAt: now,
		IP:         ip,
		UserAgent:  ua,
	}
	if err := s.repo.Create(sess); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *SessionService) GetByUUID(ctx context.Context, sessionUUID string) (*session_entities.Session, error) {
	return s.repo.FindByUUID(sessionUUID)
}

func (s *SessionService) Validate(ctx context.Context, sessionUUID string) (*session_entities.Session, error) {
	sess, err := s.repo.FindByUUID(sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("session: not found")
	}
	if !sess.IsValid() {
		return nil, fmt.Errorf("session: expired or revoked")
	}
	_ = s.repo.Touch(ctx, sess.ID)
	return sess, nil
}

func (s *SessionService) Revoke(ctx context.Context, id uint) error {
	return s.repo.Revoke(ctx, id)
}

func (s *SessionService) RevokeAll(ctx context.Context, userID uint) error {
	return s.repo.RevokeAll(ctx, userID)
}

func (s *SessionService) RevokeAllExcept(ctx context.Context, userID, exceptID uint) error {
	return s.repo.RevokeAllExcept(ctx, userID, exceptID)
}

func (s *SessionService) List(ctx context.Context, userID uint) ([]session_entities.Session, error) {
	return s.repo.FindActiveByUserID(userID)
}
