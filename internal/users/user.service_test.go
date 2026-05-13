package users

import (
	"errors"
	"testing"

	"github.com/ricardoalcantara/min-idp/internal/crypto"
	"github.com/ricardoalcantara/min-idp/internal/db"
	user_entities "github.com/ricardoalcantara/min-idp/internal/users/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock ---

type mockUserRepo struct {
	user *user_entities.User
	err  error
}

func (m *mockUserRepo) Create(u *user_entities.User) error               { return m.err }
func (m *mockUserRepo) FindByID(_ uint) (*user_entities.User, error)     { return m.user, m.err }
func (m *mockUserRepo) FindByUUID(_ string) (*user_entities.User, error)  { return m.user, m.err }
func (m *mockUserRepo) FindByEmail(_ string) (*user_entities.User, error) { return m.user, m.err }
func (m *mockUserRepo) Update(_ *user_entities.User) error               { return m.err }
func (m *mockUserRepo) List(_, _ int) ([]*user_entities.User, int64, error) {
	if m.user != nil {
		return []*user_entities.User{m.user}, 1, m.err
	}
	return nil, 0, m.err
}
func (m *mockUserRepo) Delete(_ uint) error { return m.err }

func newTestUserSvc(repo UserRepository) *UserService {
	return NewUserService(repo)
}

// --- Create ---

func TestUserService_Create_Success(t *testing.T) {
	svc := newTestUserSvc(&mockUserRepo{})
	u, err := svc.Create("a@b.com", "", "secret")
	require.NoError(t, err)
	assert.Equal(t, "a@b.com", u.Email)
	assert.Equal(t, "active", u.Status)
}

func TestUserService_Create_RepoError(t *testing.T) {
	repoErr := errors.New("db down")
	svc := newTestUserSvc(&mockUserRepo{err: repoErr})
	_, err := svc.Create("a@b.com", "", "secret")
	assert.ErrorIs(t, err, repoErr)
}

// --- Authenticate ---

func TestUserService_Authenticate_Success(t *testing.T) {
	hash, _ := crypto.HashPassword("correct")
	svc := newTestUserSvc(&mockUserRepo{
		user: &user_entities.User{Email: "a@b.com", PasswordHash: hash, Status: "active"},
	})
	u, err := svc.Authenticate("a@b.com", "correct")
	require.NoError(t, err)
	assert.Equal(t, "a@b.com", u.Email)
}

func TestUserService_Authenticate_UserNotFound(t *testing.T) {
	svc := newTestUserSvc(&mockUserRepo{err: db.ErrEntityNotFound})
	_, err := svc.Authenticate("nobody@b.com", "x")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestUserService_Authenticate_WrongPassword(t *testing.T) {
	hash, _ := crypto.HashPassword("correct")
	svc := newTestUserSvc(&mockUserRepo{
		user: &user_entities.User{Email: "a@b.com", PasswordHash: hash, Status: "active"},
	})
	_, err := svc.Authenticate("a@b.com", "wrong")
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestUserService_Authenticate_AccountDisabled(t *testing.T) {
	hash, _ := crypto.HashPassword("correct")
	svc := newTestUserSvc(&mockUserRepo{
		user: &user_entities.User{Email: "a@b.com", PasswordHash: hash, Status: "disabled"},
	})
	_, err := svc.Authenticate("a@b.com", "correct")
	assert.ErrorIs(t, err, ErrAccountNotActive)
}

func TestUserService_Authenticate_DBError(t *testing.T) {
	infraErr := errors.New("connection refused")
	svc := newTestUserSvc(&mockUserRepo{err: infraErr})
	_, err := svc.Authenticate("a@b.com", "x")
	assert.ErrorIs(t, err, infraErr)
}
