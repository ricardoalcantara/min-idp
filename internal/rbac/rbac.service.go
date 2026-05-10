package rbac

import rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"

type RBACRepository interface {
	CreateRole(role *rbac_entities.Role) error
	FindRoleByName(name string) (*rbac_entities.Role, error)
	FindRoleByID(id uint) (*rbac_entities.Role, error)
	CreatePermission(perm *rbac_entities.Permission) error
	AssignPermissionToRole(roleID, permID uint) error
	AssignRoleToUser(userID, roleID uint) error
	UserHasPermission(userID uint, permission string) (bool, error)
}

type RBACService struct {
	repo RBACRepository
}

func NewRBACService(repo RBACRepository) *RBACService {
	return &RBACService{repo: repo}
}

func (s *RBACService) CreateRole(name, description string, system bool) (*rbac_entities.Role, error) {
	role := &rbac_entities.Role{Name: name, Description: description, System: system}
	return role, s.repo.CreateRole(role)
}

func (s *RBACService) FindRoleByName(name string) (*rbac_entities.Role, error) {
	return s.repo.FindRoleByName(name)
}

func (s *RBACService) FindRoleByID(id uint) (*rbac_entities.Role, error) {
	return s.repo.FindRoleByID(id)
}

func (s *RBACService) CreatePermission(name string) (*rbac_entities.Permission, error) {
	perm := &rbac_entities.Permission{Name: name}
	return perm, s.repo.CreatePermission(perm)
}

func (s *RBACService) AssignPermissionToRole(roleID, permID uint) error {
	return s.repo.AssignPermissionToRole(roleID, permID)
}

func (s *RBACService) AssignRoleToUser(userID, roleID uint) error {
	return s.repo.AssignRoleToUser(userID, roleID)
}

func (s *RBACService) UserHasPermission(userID uint, permission string) (bool, error) {
	return s.repo.UserHasPermission(userID, permission)
}
