package keystore_repositories

import (
	"context"
	"errors"
	"time"

	"github.com/ricardoalcantara/min-idp/internal/db"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	gormdb "gorm.io/gorm"
)

type KeyRepository struct {
	db *gormdb.DB
}

func NewKeyRepository(d *gormdb.DB) *KeyRepository {
	return &KeyRepository{db: d}
}

func (r *KeyRepository) Insert(ctx context.Context, key *keystore_entities.SigningKey) error {
	return r.db.WithContext(ctx).Create(key).Error
}

func (r *KeyRepository) GetActive(ctx context.Context, protocol string) (*keystore_entities.SigningKey, error) {
	var key keystore_entities.SigningKey
	err := r.db.WithContext(ctx).
		Where("protocol = ? AND status = ?", protocol, keystore_entities.StatusActive).
		First(&key).Error
	if errors.Is(err, gormdb.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &key, err
}

func (r *KeyRepository) ListPublished(ctx context.Context, protocol string) ([]*keystore_entities.SigningKey, error) {
	var keys []*keystore_entities.SigningKey
	err := r.db.WithContext(ctx).
		Where("protocol = ? AND status IN ?", protocol, []string{keystore_entities.StatusActive, keystore_entities.StatusPrevious}).
		Find(&keys).Error
	return keys, err
}

func (r *KeyRepository) SetStatus(ctx context.Context, id uint, status string, ts time.Time) error {
	updates := map[string]any{"status": status}
	if status == keystore_entities.StatusRetired {
		updates["retired_at"] = ts
	}
	return r.db.WithContext(ctx).Model(&keystore_entities.SigningKey{}).
		Where("id = ?", id).Updates(updates).Error
}
