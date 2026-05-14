package authn

import (
	"errors"
	"testing"

	"github.com/ricardoalcantara/min-idp/internal/users"
	user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock ---

type mockUserAuthenticator struct {
	user *user_entities.User
	err  error
}

func (m *mockUserAuthenticator) Authenticate(_, _ string) (*user_entities.User, error) {
	return m.user, m.err
}

func newTestAuthnSvc(auth UserAuthenticator) *AuthnService {
	return NewAuthnService(auth)
}

// --- Authenticate ---

func TestAuthnService_Authenticate_Success(t *testing.T) {
	u := &user_entities.User{Email: "a@b.com"}
	svc := newTestAuthnSvc(&mockUserAuthenticator{user: u})
	got, err := svc.Authenticate("a@b.com", "pass")
	require.NoError(t, err)
	assert.Equal(t, "a@b.com", got.Email)
}

func TestAuthnService_Authenticate_InvalidCredentials(t *testing.T) {
	svc := newTestAuthnSvc(&mockUserAuthenticator{err: users.ErrInvalidCredentials})
	_, err := svc.Authenticate("a@b.com", "wrong")
	assert.ErrorIs(t, err, errInvalidCredentials)
}

func TestAuthnService_Authenticate_AccountNotActive(t *testing.T) {
	svc := newTestAuthnSvc(&mockUserAuthenticator{err: users.ErrAccountNotActive})
	_, err := svc.Authenticate("a@b.com", "pass")
	assert.ErrorIs(t, err, errInvalidCredentials)
}

func TestAuthnService_Authenticate_InfraError(t *testing.T) {
	infraErr := errors.New("db down")
	svc := newTestAuthnSvc(&mockUserAuthenticator{err: infraErr})
	_, err := svc.Authenticate("a@b.com", "pass")
	assert.ErrorIs(t, err, infraErr)
	assert.NotErrorIs(t, err, errInvalidCredentials)
}
