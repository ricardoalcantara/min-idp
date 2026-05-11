package session

import (
	"context"
	"fmt"
	"time"

	"github.com/ricardoalcantara/min-idp/internal/config"
	session_entities "github.com/ricardoalcantara/min-idp/internal/session/entities"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
)

type SessionRepository interface {
	Create(s *session_entities.Session) error
	FindByUUID(uuid string) (*session_entities.Session, error)
	FindActiveByUserID(userID uint) ([]session_entities.Session, error)
	RevokeByUUID(ctx context.Context, sessionUUID string) (*session_entities.Session, error)
	RevokeAllExceptUUID(ctx context.Context, userID uint, exceptUUID string) ([]session_entities.Session, error)
	RevokeAll(ctx context.Context, userID uint) error
}

type SessionService struct {
	repo SessionRepository
	kv   kvstore.KVStore
	cfg  *config.Config
}

func NewSessionService(repo SessionRepository, kv kvstore.KVStore, cfg *config.Config) *SessionService {
	return &SessionService{repo: repo, kv: kv, cfg: cfg}
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
	return sess, nil
}

// Revoke revokes a session by UUID and adds it to the KV revocation list.
func (s *SessionService) Revoke(ctx context.Context, sessionUUID string) error {
	sess, err := s.repo.RevokeByUUID(ctx, sessionUUID)
	if err != nil {
		return err
	}
	return s.addToRevocationList(ctx, sess.UUID.String(), sess.ExpiresAt)
}

// RevokeAll revokes all sessions for a user (used by admin force-logout).
func (s *SessionService) RevokeAll(ctx context.Context, userID uint) error {
	return s.repo.RevokeAll(ctx, userID)
}

// RevokeAllExcept revokes all sessions for a user except the given session UUID.
func (s *SessionService) RevokeAllExcept(ctx context.Context, userID uint, exceptSessionUUID string) error {
	revoked, err := s.repo.RevokeAllExceptUUID(ctx, userID, exceptSessionUUID)
	if err != nil {
		return err
	}
	for _, sess := range revoked {
		_ = s.addToRevocationList(ctx, sess.UUID.String(), sess.ExpiresAt)
	}
	return nil
}

func (s *SessionService) List(ctx context.Context, userID uint) ([]session_entities.Session, error) {
	return s.repo.FindActiveByUserID(userID)
}

func (s *SessionService) addToRevocationList(ctx context.Context, sessionUUID string, expiresAt time.Time) error {
	ttl := time.Until(expiresAt)
	if ttl <= 0 {
		return nil
	}
	return s.kv.Set(ctx, "revoked:"+sessionUUID, []byte("1"), ttl)
}
