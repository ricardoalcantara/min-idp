package audit_repositories

import (
	"time"

	audit_entities "github.com/ricardoalcantara/min-idp/internal/audit/entities"
	"gorm.io/gorm"
)

type AuditFilter struct {
	Action     string
	TargetType string
	Result     string
	Since      time.Time
	Until      time.Time
}

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Insert(e *audit_entities.Event) error {
	return r.db.Create(e).Error
}

func (r *AuditRepository) List(filter AuditFilter, offset, limit int) ([]audit_entities.Event, int64, error) {
	q := r.db.Model(&audit_entities.Event{})

	if filter.Action != "" {
		q = q.Where("action = ?", filter.Action)
	}
	if filter.TargetType != "" {
		q = q.Where("target_type = ?", filter.TargetType)
	}
	if filter.Result != "" {
		q = q.Where("result = ?", filter.Result)
	}
	if !filter.Since.IsZero() {
		q = q.Where("timestamp >= ?", filter.Since)
	}
	if !filter.Until.IsZero() {
		q = q.Where("timestamp <= ?", filter.Until)
	}

	var total int64
	q.Count(&total)

	var events []audit_entities.Event
	err := q.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&events).Error
	return events, total, err
}
