package rbac_repositories

import (
	"errors"

	"github.com/go-minstack/repository"
	"github.com/ricardoalcantara/min-idp/internal/db"
	rbac_entities "github.com/ricardoalcantara/min-idp/internal/rbac/entities"
	"gorm.io/gorm"
)

type RBACRepository struct {
	*repository.Repository[rbac_entities.Role]
	db *gorm.DB
}

func NewRBACRepository(d *gorm.DB) *RBACRepository {
	return &RBACRepository{Repository: repository.NewRepository[rbac_entities.Role](d), db: d}
}

func (r *RBACRepository) CreateRole(role *rbac_entities.Role) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(role).Error; err != nil {
			return err
		}
		return tx.Create(&db.Subject{Type: db.SubjectTypeRole, EntityID: role.ID}).Error
	})
}

func (r *RBACRepository) FindRoleByName(name string) (*rbac_entities.Role, error) {
	role, err := r.FindOne(repository.Where("name = ?", name))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return role, err
}

func (r *RBACRepository) FindRoleByID(id uint) (*rbac_entities.Role, error) {
	role, err := r.FindByID(id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return role, err
}

func (r *RBACRepository) FindRoleByUUID(uuid string) (*rbac_entities.Role, error) {
	role, err := r.FindOne(repository.Where("uuid = ?", uuid))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, db.ErrEntityNotFound
	}
	return role, err
}

func (r *RBACRepository) ListRoles() ([]rbac_entities.Role, error) {
	return r.FindAll()
}

func (r *RBACRepository) UpdateRole(role *rbac_entities.Role) error {
	return r.Update(role)
}

func (r *RBACRepository) DeleteRole(id uint) error {
	return r.DeleteByID(id)
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

// GetSubjectIDsForUser returns the subjects.id values for every principal the
// user belongs to: their own user-subject, all role-subjects, all group-subjects.
// Used by the SSO gate to match access_rules.subject_id in one query.
func (r *RBACRepository) GetSubjectIDsForUser(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Raw(`
		SELECT s.id FROM subjects s
		WHERE (s.type = 'user' AND s.entity_id = ?)
		   OR (s.type = 'role' AND s.entity_id IN (
		           SELECT role_id FROM user_roles WHERE user_id = ?
		       ))
		   OR (s.type = 'group' AND s.entity_id IN (
		           SELECT group_id FROM user_groups WHERE user_id = ?
		       ))
	`, userID, userID, userID).Scan(&ids).Error
	return ids, err
}
