package kvstore

import (
	"context"
	"errors"
	"log/slog"
	"time"

	kvstore_entities "github.com/ricardoalcantara/min-idp/internal/kvstore/entities"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DBKVStore struct {
	db  *gorm.DB
	log *slog.Logger
}

func NewDBKVStore(db *gorm.DB, lc fx.Lifecycle, log *slog.Logger) *DBKVStore {
	s := &DBKVStore{db: db, log: log}
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go s.janitor(ctx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
	return s
}

func (s *DBKVStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().UTC().Add(ttl)
		expiresAt = &t
	}
	return s.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "key"}},
		DoUpdates: clause.AssignmentColumns([]string{"value", "expires_at"}),
	}).Create(&kvstore_entities.KVEntry{Key: key, Value: value, ExpiresAt: expiresAt}).Error
}

func (s *DBKVStore) Get(ctx context.Context, key string) ([]byte, error) {
	var entry kvstore_entities.KVEntry
	err := s.db.WithContext(ctx).Where("key = ?", key).First(&entry).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if entry.ExpiresAt != nil && time.Now().After(*entry.ExpiresAt) {
		_ = s.db.WithContext(ctx).Delete(&kvstore_entities.KVEntry{}, "key = ?", key)
		return nil, ErrNotFound
	}
	return entry.Value, nil
}

func (s *DBKVStore) Delete(ctx context.Context, key string) error {
	return s.db.WithContext(ctx).Delete(&kvstore_entities.KVEntry{}, "key = ?", key).Error
}

func (s *DBKVStore) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	var expiresAt *time.Time
	if ttl > 0 {
		t := time.Now().UTC().Add(ttl)
		expiresAt = &t
	}
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).
		Create(&kvstore_entities.KVEntry{Key: key, Value: value, ExpiresAt: expiresAt})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected == 1, nil
}

func (s *DBKVStore) janitor(ctx context.Context) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.db.Delete(&kvstore_entities.KVEntry{}, "expires_at IS NOT NULL AND expires_at < ?", time.Now().UTC())
		}
	}
}
