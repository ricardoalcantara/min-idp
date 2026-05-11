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
	role       *rbac_entities.Role
	permission *rbac_entities.Permission
	hasPerm    bool
	err        error
}

func (m *mockRBACRepo) CreateRole(role *rbac_entities.Role) error                     { return m.err }
func (m *mockRBACRepo) FindRoleByName(_ string) (*rbac_entities.Role, error)            { return m.role, m.err }
func (m *mockRBACRepo) FindRoleByID(_ uint) (*rbac_entities.Role, error)               { return m.role, m.err }
func (m *mockRBACRepo) FindRoleByUUID(_ string) (*rbac_entities.Role, error)           { return m.role, m.err }
func (m *mockRBACRepo) ListRoles() ([]rbac_entities.Role, error)                       { return nil, m.err }
func (m *mockRBACRepo) UpdateRole(_ *rbac_entities.Role) error                         { return m.err }
func (m *mockRBACRepo) DeleteRole(_ uint) error                                        { return m.err }
func (m *mockRBACRepo) CreatePermission(p *rbac_entities.Permission) error             { return m.err }
func (m *mockRBACRepo) FindPermissionByUUID(_ string) (*rbac_entities.Permission, error) { return m.permission, m.err }
func (m *mockRBACRepo) GetPermissionsByRole(_ uint) ([]rbac_entities.Permission, error) { return nil, m.err }
func (m *mockRBACRepo) AssignPermissionToRole(_, _ uint) error                         { return m.err }
func (m *mockRBACRepo) RemovePermissionFromRole(_, _ uint) error                       { return m.err }
func (m *mockRBACRepo) AssignRoleToUser(_, _ uint) error                               { return m.err }
func (m *mockRBACRepo) RemoveRoleFromUser(_, _ uint) error                             { return m.err }
func (m *mockRBACRepo) GetRolesByUser(_ uint) ([]rbac_entities.Role, error)            { return nil, m.err }
func (m *mockRBACRepo) UserHasPermission(_ uint, _ string) (bool, error)              { return m.hasPerm, m.err }
func (m *mockRBACRepo) GetUserPermissions(_ uint) ([]string, error)                   { return nil, m.err }

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

// --- UserHasPermission ---

func TestRBACService_UserHasPermission_True(t *testing.T) {
	svc := newTestRBACSvc(&mockRBACRepo{hasPerm: true})
	ok, err := svc.UserHasPermission(1, "system:admin")
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestRBACService_UserHasPermission_False(t *testing.T) {
	svc := newTestRBACSvc(&mockRBACRepo{hasPerm: false})
	ok, err := svc.UserHasPermission(1, "system:admin")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestRBACService_UserHasPermission_DBError(t *testing.T) {
	svc := newTestRBACSvc(&mockRBACRepo{err: errors.New("db down")})
	_, err := svc.UserHasPermission(1, "system:admin")
	assert.Error(t, err)
}
