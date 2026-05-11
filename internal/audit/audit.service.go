package audit

import (
	"log/slog"
	"time"

	audit_entities "github.com/ricardoalcantara/min-idp/internal/audit/entities"
	audit_repositories "github.com/ricardoalcantara/min-idp/internal/audit/repositories"
)

type AuditRepository interface {
	Insert(e *audit_entities.Event) error
	List(filter audit_repositories.AuditFilter, offset, limit int) ([]audit_entities.Event, int64, error)
}

type AuditService struct {
	repo AuditRepository
	log  *slog.Logger
}

func NewAuditService(repo AuditRepository, log *slog.Logger) *AuditService {
	return &AuditService{repo: repo, log: log}
}

func (s *AuditService) Log(e audit_entities.Event) {
	e.Timestamp = time.Now().UTC()
	go func() {
		if err := s.repo.Insert(&e); err != nil {
			s.log.Error("audit: failed to write event", "action", e.Action, "err", err)
		}
	}()
}

func (s *AuditService) List(filter audit_repositories.AuditFilter, page, pageSize int) ([]audit_entities.Event, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	return s.repo.List(filter, (page-1)*pageSize, pageSize)
}
