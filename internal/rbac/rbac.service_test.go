package rbac

import (
	"errors"
	"testing"

	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- mock ---

type mockRBACRepo struct {
	role    *rbac_entities.Role
	hasRole bool
	err     error
}

func (m *mockRBACRepo) CreateRole(role *rbac_entities.Role) error                  { return m.err }
func (m *mockRBACRepo) FindRoleByName(_ string) (*rbac_entities.Role, error)       { return m.role, m.err }
func (m *mockRBACRepo) FindRoleByID(_ uint) (*rbac_entities.Role, error)           { return m.role, m.err }
func (m *mockRBACRepo) FindRoleByUUID(_ string) (*rbac_entities.Role, error)       { return m.role, m.err }
func (m *mockRBACRepo) ListRoles() ([]rbac_entities.Role, error)                   { return nil, m.err }
func (m *mockRBACRepo) UpdateRole(_ *rbac_entities.Role) error                     { return m.err }
func (m *mockRBACRepo) DeleteRole(_ uint) error                                    { return m.err }
func (m *mockRBACRepo) AssignRoleToUser(_, _ uint) error                           { return m.err }
func (m *mockRBACRepo) RemoveRoleFromUser(_, _ uint) error                         { return m.err }
func (m *mockRBACRepo) GetRolesByUser(_ uint) ([]rbac_entities.Role, error)        { return nil, m.err }
func (m *mockRBACRepo) UserHasRole(_ uint, _ string) (bool, error)                 { return m.hasRole, m.err }
func (m *mockRBACRepo) GetUserRoleNames(_ uint) ([]string, error)                  { return nil, m.err }
func (m *mockRBACRepo) GetSubjectIDsForUser(_ uint) ([]uint, error)                { return nil, m.err }

func newTestRBACSvc(repo RBACRepository) *RBACService {
	return NewRBACService(repo)
}

// --- CreateRole ---

func TestRBACService_CreateRole_Success(t *testing.T) {
	svc := newTestRBACSvc(&mockRBACRepo{})
	role, err := svc.CreateRole("system:admin", "Admin role", true)
	require.NoError(t, err)
	assert.Equal(t, "system:admin", role.Name)
	assert.True(t, role.System)
}

func TestRBACService_CreateRole_RepoError(t *testing.T) {
	svc := newTestRBACSvc(&mockRBACRepo{err: errors.New("db down")})
	_, err := svc.CreateRole("system:admin", "", true)
	assert.Error(t, err)
}

// --- UserHasRole ---

func TestRBACService_UserHasRole_True(t *testing.T) {
	svc := newTestRBACSvc(&mockRBACRepo{hasRole: true})
	ok, err := svc.UserHasRole(1, "system:admin")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRBACService_UserHasRole_False(t *testing.T) {
	svc := newTestRBACSvc(&mockRBACRepo{hasRole: false})
	ok, err := svc.UserHasRole(1, "system:admin")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRBACService_UserHasRole_DBError(t *testing.T) {
	svc := newTestRBACSvc(&mockRBACRepo{err: errors.New("db down")})
	_, err := svc.UserHasRole(1, "system:admin")
	assert.Error(t, err)
}
