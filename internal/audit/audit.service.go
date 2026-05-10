package audit

import (
	"log/slog"
	"time"

	audit_entities "github.com/ricardoalcantara/min-idp/internal/audit/entities"
	audit_repositories "github.com/ricardoalcantara/min-idp/internal/audit/repositories"
)

type AuditService struct {
	repo *audit_repositories.AuditRepository
	log  *slog.Logger
}

func NewAuditService(repo *audit_repositories.AuditRepository, log *slog.Logger) *AuditService {
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
