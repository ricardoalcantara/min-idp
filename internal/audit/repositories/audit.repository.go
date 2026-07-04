package audit_repositories

import (
	"time"

	"github.com/go-minstack/go-minstack/repository"
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
	*repository.Repository[audit_entities.Event]
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{repository.NewRepository[audit_entities.Event](db)}
}

func buildFilterOpts(filter AuditFilter) []repository.QueryOption {
	var opts []repository.QueryOption
	if filter.Action != "" {
		opts = append(opts, repository.Where("action = ?", filter.Action))
	}
	if filter.TargetType != "" {
		opts = append(opts, repository.Where("target_type = ?", filter.TargetType))
	}
	if filter.Result != "" {
		opts = append(opts, repository.Where("result = ?", filter.Result))
	}
	if !filter.Since.IsZero() {
		opts = append(opts, repository.Where("timestamp >= ?", filter.Since))
	}
	if !filter.Until.IsZero() {
		opts = append(opts, repository.Where("timestamp <= ?", filter.Until))
	}
	return opts
}

func (r *AuditRepository) List(filter AuditFilter, page, pageSize int) ([]audit_entities.Event, int64, error) {
	opts := buildFilterOpts(filter)

	total, err := r.Count(opts...)
	if err != nil {
		return nil, 0, err
	}

	queryOpts := append(opts,
		repository.Order("timestamp", true),
		repository.Paginate(repository.NewPagination(page, pageSize)),
	)
	events, err := r.FindAll(queryOpts...)
	return events, total, err
}
