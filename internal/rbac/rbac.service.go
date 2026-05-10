package rbac

import rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"

type RBACRepository interface {
	CreateRole(role *rbac_entities.Role) error
	FindRoleByName(name string) (*rbac_entities.Role, error)
	FindRoleByID(id uint) (*rbac_entities.Role, error)
	FindRoleByUUID(uuid string) (*rbac_entities.Role, error)
	ListRoles() ([]rbac_entities.Role, error)
	UpdateRole(role *rbac_entities.Role) error
	DeleteRole(id uint) error
	CreatePermission(perm *rbac_entities.Permission) error
	FindPermissionByUUID(uuid string) (*rbac_entities.Permission, error)
	GetPermissionsByRole(roleID uint) ([]rbac_entities.Permission, error)
	AssignPermissionToRole(roleID, permID uint) error
	RemovePermissionFromRole(roleID, permID uint) error
	AssignRoleToUser(userID, roleID uint) error
	RemoveRoleFromUser(userID, roleID uint) error
	GetRolesByUser(userID uint) ([]rbac_entities.Role, error)
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

func (s *RBACService) FindRoleByUUID(uuid string) (*rbac_entities.Role, error) {
	return s.repo.FindRoleByUUID(uuid)
}

func (s *RBACService) ListRoles() ([]rbac_entities.Role, error) {
	return s.repo.ListRoles()
}

func (s *RBACService) UpdateRole(id uint, name, description *string) (*rbac_entities.Role, error) {
	role, err := s.repo.FindRoleByID(id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		role.Name = *name
	}
	if description != nil {
		role.Description = *description
	}
	return role, s.repo.UpdateRole(role)
}

func (s *RBACService) DeleteRole(id uint) error {
	return s.repo.DeleteRole(id)
}

func (s *RBACService) CreatePermission(name string) (*rbac_entities.Permission, error) {
	perm := &rbac_entities.Permission{Name: name}
	return perm, s.repo.CreatePermission(perm)
}

func (s *RBACService) GetPermissionsByRole(roleID uint) ([]rbac_entities.Permission, error) {
	return s.repo.GetPermissionsByRole(roleID)
}

func (s *RBACService) AssignPermissionToRole(roleID, permID uint) error {
	return s.repo.AssignPermissionToRole(roleID, permID)
}

func (s *RBACService) AssignPermissionToRoleByUUID(roleUUID, permName string) error {
	role, err := s.repo.FindRoleByUUID(roleUUID)
	if err != nil {
		return err
	}
	perm, err := s.CreatePermission(permName)
	if err != nil {
		return err
	}
	return s.repo.AssignPermissionToRole(role.ID, perm.ID)
}

func (s *RBACService) RemovePermissionFromRole(roleID, permID uint) error {
	return s.repo.RemovePermissionFromRole(roleID, permID)
}

func (s *RBACService) AssignRoleToUser(userID, roleID uint) error {
	return s.repo.AssignRoleToUser(userID, roleID)
}

func (s *RBACService) AssignRoleToUserByUUID(userID uint, roleUUID string) error {
	role, err := s.repo.FindRoleByUUID(roleUUID)
	if err != nil {
		return err
	}
	return s.repo.AssignRoleToUser(userID, role.ID)
}

func (s *RBACService) RemoveRoleFromUserByUUID(userID uint, roleUUID string) error {
	role, err := s.repo.FindRoleByUUID(roleUUID)
	if err != nil {
		return err
	}
	return s.repo.RemoveRoleFromUser(userID, role.ID)
}

func (s *RBACService) GetRolesByUser(userID uint) ([]rbac_entities.Role, error) {
	return s.repo.GetRolesByUser(userID)
}

func (s *RBACService) UserHasPermission(userID uint, permission string) (bool, error) {
	return s.repo.UserHasPermission(userID, permission)
}
