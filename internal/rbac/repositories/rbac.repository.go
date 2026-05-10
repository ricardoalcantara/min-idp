package rbac_repositories

import (
	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
	"gorm.io/gorm"
)

type RBACRepository struct {
	db *gorm.DB
}

func NewRBACRepository(db *gorm.DB) *RBACRepository {
	return &RBACRepository{db: db}
}

func (r *RBACRepository) CreateRole(role *rbac_entities.Role) error {
	return r.db.Create(role).Error
}

func (r *RBACRepository) FindRoleByName(name string) (*rbac_entities.Role, error) {
	var role rbac_entities.Role
	return &role, r.db.Where("name = ?", name).First(&role).Error
}

func (r *RBACRepository) FindRoleByID(id uint) (*rbac_entities.Role, error) {
	var role rbac_entities.Role
	return &role, r.db.First(&role, id).Error
}

func (r *RBACRepository) ListRoles() ([]rbac_entities.Role, error) {
	var roles []rbac_entities.Role
	return roles, r.db.Find(&roles).Error
}

func (r *RBACRepository) CreatePermission(perm *rbac_entities.Permission) error {
	return r.db.FirstOrCreate(perm, rbac_entities.Permission{Name: perm.Name}).Error
}

func (r *RBACRepository) AssignPermissionToRole(roleID, permID uint) error {
	return r.db.Where("role_id = ? AND permission_id = ?", roleID, permID).
		FirstOrCreate(&rbac_entities.RolePermission{RoleID: roleID, PermissionID: permID}).Error
}

func (r *RBACRepository) AssignRoleToUser(userID, roleID uint) error {
	return r.db.Where("user_id = ? AND role_id = ?", userID, roleID).
		FirstOrCreate(&rbac_entities.UserRole{UserID: userID, RoleID: roleID}).Error
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
