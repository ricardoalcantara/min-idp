package audit_repositories

import (
	audit_entities "github.com/ricardoalcantara/min-idp/internal/audit/entities"
	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Insert(e *audit_entities.Event) error {
	return r.db.Create(e).Error
}

func (r *AuditRepository) List(offset, limit int) ([]audit_entities.Event, int64, error) {
	var events []audit_entities.Event
	var total int64
	r.db.Model(&audit_entities.Event{}).Count(&total)
	err := r.db.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&events).Error
	return events, total, err
}
