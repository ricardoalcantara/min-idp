package audit

import (
	"io"
	"log/slog"
	"testing"
	"time"

	audit_entities "github.com/ricardoalcantara/min-idp/internal/audit/entities"
	audit_repositories "github.com/ricardoalcantara/min-idp/internal/audit/repositories"
	"github.com/stretchr/testify/assert"
)

// --- mock ---

type mockAuditRepo struct {
	inserted []*audit_entities.Event
	err      error
}

func (m *mockAuditRepo) Create(e *audit_entities.Event) error {
	m.inserted = append(m.inserted, e)
	return m.err
}

func (m *mockAuditRepo) List(_ audit_repositories.AuditFilter, _, _ int) ([]audit_entities.Event, int64, error) {
	return nil, 0, m.err
}

func newTestAuditSvc(repo AuditRepository) *AuditService {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewAuditService(repo, log)
}

// --- Log ---

func TestAuditService_Log_SetsTimestamp(t *testing.T) {
	repo := &mockAuditRepo{}
	svc := newTestAuditSvc(repo)

	before := time.Now()
	svc.Log(audit_entities.Event{Action: "test.action"})
	time.Sleep(10 * time.Millisecond) // allow goroutine to flush

	assert.Len(t, repo.inserted, 1)
	assert.True(t, repo.inserted[0].Timestamp.After(before))
	assert.Equal(t, "test.action", repo.inserted[0].Action)
}

func TestAuditService_Log_RepoErrorSilent(t *testing.T) {
	// repo errors must not propagate to caller — Log is fire-and-forget
	svc := newTestAuditSvc(&mockAuditRepo{err: assert.AnError})
	assert.NotPanics(t, func() {
		svc.Log(audit_entities.Event{Action: "test.action"})
		time.Sleep(10 * time.Millisecond)
	})
}
