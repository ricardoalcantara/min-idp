package session

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/ricardoalcantara/min-idp/internal/config"
	"github.com/ricardoalcantara/min-idp/internal/db"
	session_entities "github.com/ricardoalcantara/min-idp/internal/session/entities"
	"github.com/ricardoalcantara/min-idp/internal/kvstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mocks ---

type mockSessionRepo struct {
	session *session_entities.Session
	list    []session_entities.Session
	err     error
}

func (m *mockSessionRepo) Create(s *session_entities.Session) error { return m.err }
func (m *mockSessionRepo) FindByUUID(_ string) (*session_entities.Session, error) {
	return m.session, m.err
}
func (m *mockSessionRepo) FindActiveByUserID(_ uint) ([]session_entities.Session, error) {
	return m.list, m.err
}
func (m *mockSessionRepo) RevokeByUUID(_ context.Context, _ string) (*session_entities.Session, error) {
	return m.session, m.err
}
func (m *mockSessionRepo) RevokeAllExceptUUID(_ context.Context, _ uint, _ string) ([]session_entities.Session, error) {
	return nil, m.err
}
func (m *mockSessionRepo) RevokeAll(_ context.Context, _ uint) error { return m.err }

type mockKV struct{}

func (m *mockKV) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error      { return nil }
func (m *mockKV) Get(_ context.Context, _ string) ([]byte, error)                       { return nil, kvstore.ErrNotFound }
func (m *mockKV) Delete(_ context.Context, _ string) error                              { return nil }
func (m *mockKV) SetNX(_ context.Context, _ string, _ []byte, _ time.Duration) (bool, error) {
	return true, nil
}

func newTestSessionSvc(repo SessionRepository) *SessionService {
	return &SessionService{repo: repo, kv: &mockKV{}, cfg: &config.Config{SessionTTL: 12 * time.Hour}}
}

// --- Create ---

func TestSessionService_Create_Success(t *testing.T) {
	svc := newTestSessionSvc(&mockSessionRepo{})
	sess, err := svc.Create(context.Background(), 1, "127.0.0.1", "agent")
	require.NoError(t, err)
	assert.Equal(t, uint(1), sess.UserID)
	assert.True(t, sess.ExpiresAt.After(time.Now()))
}

func TestSessionService_Create_RepoError(t *testing.T) {
	svc := newTestSessionSvc(&mockSessionRepo{err: errors.New("db down")})
	_, err := svc.Create(context.Background(), 1, "", "")
	assert.Error(t, err)
}

// --- Validate ---

func TestSessionService_Validate_Success(t *testing.T) {
	sess := &session_entities.Session{
		UserID:    1,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	svc := newTestSessionSvc(&mockSessionRepo{session: sess})
	got, err := svc.Validate(context.Background(), "some-uuid")
	require.NoError(t, err)
	assert.Equal(t, uint(1), got.UserID)
}

func TestSessionService_Validate_NotFound(t *testing.T) {
	svc := newTestSessionSvc(&mockSessionRepo{err: db.ErrEntityNotFound})
	_, err := svc.Validate(context.Background(), "bad-uuid")
	assert.Error(t, err)
}

func TestSessionService_Validate_Expired(t *testing.T) {
	sess := &session_entities.Session{
		UserID:    1,
		ExpiresAt: time.Now().Add(-time.Hour),
	}
	svc := newTestSessionSvc(&mockSessionRepo{session: sess})
	_, err := svc.Validate(context.Background(), "some-uuid")
	assert.Error(t, err)
}

func TestSessionService_Validate_Revoked(t *testing.T) {
	revokedAt := time.Now().Add(-time.Minute)
	sess := &session_entities.Session{
		UserID:    1,
		ExpiresAt: time.Now().Add(time.Hour),
		RevokedAt: &revokedAt,
	}
	svc := newTestSessionSvc(&mockSessionRepo{session: sess})
	_, err := svc.Validate(context.Background(), "some-uuid")
	assert.Error(t, err)
}
