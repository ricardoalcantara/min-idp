package rbac

import (
	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
	rbac_repositories "github.com/ricardoalcantara/min-idp/internal/rbac/repositories"
)

type RBACService struct {
	repo *rbac_repositories.RBACRepository
}

func NewRBACService(repo *rbac_repositories.RBACRepository) *RBACService {
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
