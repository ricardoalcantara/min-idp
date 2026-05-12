package keystore_repositories

import (
	"context"
	"errors"
	"time"

	"github.com/go-minstack/repository"
	"github.com/ricardoalcantara/min-idp/internal/db"
	keystore_entities "github.com/ricardoalcantara/min-idp/internal/keystore/entities"
	gormdb "gorm.io/gorm"
)

type KeyRepository struct {
	*repository.Repository[keystore_entities.SigningKey]
	db *gormdb.DB // kept for SetStatus (conditional UPDATE with context)
}

func NewKeyRepository(d *gormdb.DB) *KeyRepository {
	return &KeyRepository{Repository: repository.NewRepository[keystore_entities.SigningKey](d), db: d}
}

func (r *KeyRepository) GetActive(ctx context.Context, protocol string) (*keystore_entities.SigningKey, error) {
	key, err := r.FindOne(repository.Where("protocol = ? AND status = ?", protocol, keystore_entities.StatusActive))
	if err != nil {
		if errors.Is(err, gormdb.ErrRecordNotFound) {
			return nil, db.ErrEntityNotFound
		}
		return nil, err
	}
	return key, nil
}

func (r *KeyRepository) ListAll(ctx context.Context) ([]*keystore_entities.SigningKey, error) {
	keys, err := r.FindAll(
		repository.Order("protocol"),
		repository.Order("created_at", true),
	)
	if err != nil {
		return nil, err
	}
	ptrs := make([]*keystore_entities.SigningKey, len(keys))
	for i := range keys {
		ptrs[i] = &keys[i]
	}
	return ptrs, nil
}

func (r *KeyRepository) ListByProtocol(ctx context.Context, protocol string) ([]*keystore_entities.SigningKey, error) {
	keys, err := r.FindAll(
		repository.Where("protocol = ?", protocol),
		repository.Order("created_at", true),
	)
	if err != nil {
		return nil, err
	}
	ptrs := make([]*keystore_entities.SigningKey, len(keys))
	for i := range keys {
		ptrs[i] = &keys[i]
	}
	return ptrs, nil
}

func (r *KeyRepository) ListPublished(ctx context.Context, protocol string) ([]*keystore_entities.SigningKey, error) {
	keys, err := r.FindAll(
		repository.Where("protocol = ? AND status IN ?", protocol, []string{keystore_entities.StatusActive, keystore_entities.StatusPrevious}),
	)
	if err != nil {
		return nil, err
	}
	ptrs := make([]*keystore_entities.SigningKey, len(keys))
	for i := range keys {
		ptrs[i] = &keys[i]
	}
	return ptrs, nil
}

func (r *KeyRepository) SetStatus(ctx context.Context, id uint, status string, ts time.Time) error {
	updates := map[string]any{"status": status}
	if status == keystore_entities.StatusRetired {
		updates["retired_at"] = ts
	}
	return r.db.WithContext(ctx).Model(&keystore_entities.SigningKey{}).
		Where("id = ?", id).Updates(updates).Error
}
