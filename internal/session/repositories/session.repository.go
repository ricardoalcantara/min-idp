package session_repositories

import (
	"context"
	"errors"
	"time"

	"github.com/go-minstack/repository"
	"github.com/ricardoalcantara/min-idp/internal/db"
	session_entities "github.com/ricardoalcantara/min-idp/internal/session/entities"
	gormdb "gorm.io/gorm"
)

type SessionRepository struct {
	*repository.Repository[session_entities.Session]
	db *gormdb.DB
}

func NewSessionRepository(d *gormdb.DB) *SessionRepository {
	return &SessionRepository{Repository: repository.NewRepository[session_entities.Session](d), db: d}
}

func (r *SessionRepository) FindByUUID(uuid string) (*session_entities.Session, error) {
	s, err := r.FindOne(repository.Where("uuid = ?", uuid))
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return s, err
}

func (r *SessionRepository) FindActiveByUserID(userID uint) ([]session_entities.Session, error) {
	return r.FindAll(repository.Where("user_id = ? AND revoked_at IS NULL", userID))
}

// RevokeByUUID marks a session as revoked by its UUID and returns its expiry for KV TTL.
func (r *SessionRepository) RevokeByUUID(ctx context.Context, sessionUUID string) (*session_entities.Session, error) {
	s, err := r.FindByUUID(sessionUUID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	if err := r.db.WithContext(ctx).Model(s).Update("revoked_at", now).Error; err != nil {
		return nil, err
	}
	return s, nil
}

// RevokeAllExceptUUID revokes all active sessions for a user except the given session UUID.
// Returns the UUIDs and expiries of revoked sessions (for KV blocklist insertion).
func (r *SessionRepository) RevokeAllExceptUUID(ctx context.Context, userID uint, exceptUUID string) ([]session_entities.Session, error) {
	var sessions []session_entities.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND uuid != ? AND revoked_at IS NULL", userID, exceptUUID).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	if len(sessions) == 0 {
		return nil, nil
	}
	now := time.Now().UTC()
	uuids := make([]string, len(sessions))
	for i, s := range sessions {
		uuids[i] = s.UUID.String()
	}
	err = r.db.WithContext(ctx).Model(&session_entities.Session{}).
		Where("user_id = ? AND uuid IN ? AND revoked_at IS NULL", userID, uuids).
		Update("revoked_at", now).Error
	return sessions, err
}

func (r *SessionRepository) RevokeAll(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&session_entities.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now().UTC()).Error
}
