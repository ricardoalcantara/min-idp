package bootstrap_repositories

import (
	"context"
	"errors"

	bootstrap_entities "github.com/ricardoalcantara/min-idp/internal/bootstrap/entities"
	"gorm.io/gorm"
)

type BootstrapRepository struct {
	db *gorm.DB
}

func NewBootstrapRepository(db *gorm.DB) *BootstrapRepository {
	return &BootstrapRepository{db: db}
}

func (r *BootstrapRepository) IsInitialized(ctx context.Context) (bool, error) {
	var state bootstrap_entities.BootstrapState
	err := r.db.WithContext(ctx).Where("key = ?", "initialized").First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return state.Value == "true", nil
}

func (r *BootstrapRepository) SetInitialized(ctx context.Context) error {
	return r.db.WithContext(ctx).Save(&bootstrap_entities.BootstrapState{
		Key:   "initialized",
		Value: "true",
	}).Error
}
