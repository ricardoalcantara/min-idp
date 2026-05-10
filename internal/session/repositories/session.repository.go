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

func (r *SessionRepository) Revoke(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&session_entities.Session{}).
		Where("id = ?", id).Update("revoked_at", time.Now().UTC()).Error
}

func (r *SessionRepository) RevokeAll(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Model(&session_entities.Session{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", time.Now().UTC()).Error
}

func (r *SessionRepository) RevokeAllExcept(ctx context.Context, userID, exceptID uint) error {
	return r.db.WithContext(ctx).Model(&session_entities.Session{}).
		Where("user_id = ? AND id != ? AND revoked_at IS NULL", userID, exceptID).
		Update("revoked_at", time.Now().UTC()).Error
}

func (r *SessionRepository) Touch(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Model(&session_entities.Session{}).
		Where("id = ?", id).Update("last_seen_at", time.Now().UTC()).Error
}
