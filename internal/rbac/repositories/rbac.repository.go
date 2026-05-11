package rbac_repositories

import (
	"errors"

	"github.com/ricardoalcantara/min-idp/internal/db"
	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
	"gorm.io/gorm"
)

type RBACRepository struct {
	db *gorm.DB
}

func NewRBACRepository(d *gorm.DB) *RBACRepository {
	return &RBACRepository{db: d}
}

func (r *RBACRepository) CreateRole(role *rbac_entities.Role) error {
	return r.db.Create(role).Error
}

func (r *RBACRepository) FindRoleByName(name string) (*rbac_entities.Role, error) {
	var role rbac_entities.Role
	err := r.db.Where("name = ?", name).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &role, err
}

func (r *RBACRepository) FindRoleByID(id uint) (*rbac_entities.Role, error) {
	var role rbac_entities.Role
	err := r.db.First(&role, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &role, err
}

func (r *RBACRepository) FindRoleByUUID(uuid string) (*rbac_entities.Role, error) {
	var role rbac_entities.Role
	err := r.db.Where("uuid = ?", uuid).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &role, err
}

func (r *RBACRepository) ListRoles() ([]rbac_entities.Role, error) {
	var roles []rbac_entities.Role
	return roles, r.db.Find(&roles).Error
}

func (r *RBACRepository) UpdateRole(role *rbac_entities.Role) error {
	return r.db.Save(role).Error
}

func (r *RBACRepository) DeleteRole(id uint) error {
	return r.db.Delete(&rbac_entities.Role{}, id).Error
}

func (r *RBACRepository) CreatePermission(perm *rbac_entities.Permission) error {
	return r.db.FirstOrCreate(perm, rbac_entities.Permission{Name: perm.Name}).Error
}

func (r *RBACRepository) FindPermissionByUUID(uuid string) (*rbac_entities.Permission, error) {
	var perm rbac_entities.Permission
	err := r.db.Where("uuid = ?", uuid).First(&perm).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return &perm, err
}

func (r *RBACRepository) GetPermissionsByRole(roleID uint) ([]rbac_entities.Permission, error) {
	var perms []rbac_entities.Permission
	err := r.db.
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&perms).Error
	return perms, err
}

func (r *RBACRepository) AssignPermissionToRole(roleID, permID uint) error {
	return r.db.Where("role_id = ? AND permission_id = ?", roleID, permID).
		FirstOrCreate(&rbac_entities.RolePermission{RoleID: roleID, PermissionID: permID}).Error
}

func (r *RBACRepository) RemovePermissionFromRole(roleID, permID uint) error {
	return r.db.Where("role_id = ? AND permission_id = ?", roleID, permID).
		Delete(&rbac_entities.RolePermission{}).Error
}

func (r *RBACRepository) AssignRoleToUser(userID, roleID uint) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		FirstOrCreate(&rbac_entities.UserRole{UserID: userID, RoleID: roleID}).Error
}

func (r *RBACRepository) RemoveRoleFromUser(userID, roleID uint) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&rbac_entities.UserRole{}).Error
}

func (r *RBACRepository) GetRolesByUser(userID uint) ([]rbac_entities.Role, error) {
	var roles []rbac_entities.Role
	err := r.db.
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&roles).Error
	return roles, err
}

func (r *RBACRepository) UserHasPermission(userID uint, permission string) (bool, error) {
	var count int64
	err := r.db.Model(&rbac_entities.Permission{}).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ? AND permissions.name = ?", userID, permission).
		Count(&count).Error
	return count > 0, err
}

func (r *RBACRepository) GetUserPermissions(userID uint) ([]string, error) {
	var names []string
	err := r.db.Model(&rbac_entities.Permission{}).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
		Distinct("permissions.name").
		Pluck("name", &names).Error
	return names, err
}
